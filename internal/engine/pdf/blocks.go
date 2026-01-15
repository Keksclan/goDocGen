// Package pdf enthält die Logik zum Rendern der verschiedenen Dokumentblöcke.
package pdf

import (
	"fmt"
	"godocgen/internal/blocks"
	"godocgen/internal/util"
	"strings"
	"unicode"
)

// cleanCodeText entfernt problematische Unicode-Zeichen aus Code-Text.
// Dies verhindert Kästchen und seltsame Darstellungen im PDF.
func cleanCodeText(text string) string {
	var result strings.Builder
	for _, r := range text {
		// Tabs durch 4 Leerzeichen ersetzen (Standard für Code)
		if r == '\t' {
			result.WriteString("    ")
		} else if r == '\n' || r == '\r' {
			result.WriteRune(r)
		} else if r >= 32 && r < 127 {
			// Standard ASCII druckbare Zeichen
			result.WriteRune(r)
		} else if r >= 128 && r < 256 {
			// Latin-1 Supplement (Umlaute etc.)
			result.WriteRune(r)
		} else if unicode.IsPrint(r) && !unicode.IsControl(r) {
			// Andere druckbare Unicode-Zeichen (aber keine Steuerzeichen)
			// Ersetze durch ASCII-Äquivalent wenn möglich
			switch r {
			case '→':
				result.WriteString("->")
			case '←':
				result.WriteString("<-")
			case '≥':
				result.WriteString(">=")
			case '≤':
				result.WriteString("<=")
			case '≠':
				result.WriteString("!=")
			case '…':
				result.WriteString("...")
			case '\u00A0': // Non-breaking space
				result.WriteRune(' ')
			case '\uFEFF': // BOM
				// Ignorieren
			case '\u200B', '\u200C', '\u200D': // Zero-width spaces
				// Ignorieren
			default:
				// Für andere Unicode-Zeichen: prüfen ob im erweiterten Latin-Bereich
				if r < 0x2000 {
					result.WriteRune(r)
				}
				// Sonst ignorieren (z.B. Emoji, spezielle Symbole)
			}
		}
		// Steuerzeichen und andere problematische Zeichen werden ignoriert
	}
	return result.String()
}

func (g *Generator) fixSegmentSpacing(content []blocks.TextSegment) {
	for i := 0; i < len(content); i++ {
		seg := &content[i]
		// Punctuation Spacing Korrektur (nur für normalen Text, nicht für Inline-Code)
		if !seg.Code && seg.Link == "" {
			seg.Text = util.FixPunctuationSpacing(seg.Text)

			// Segmentgrenzen-Korrektur: Wenn dieses Segment mit einem Satzzeichen endet
			// und das nächste Segment mit einem Buchstaben/Ziffer beginnt.
			if i < len(content)-1 {
				nextSeg := &content[i+1]
				if len(seg.Text) > 0 && len(nextSeg.Text) > 0 {
					lastChar := seg.Text[len(seg.Text)-1]
					nextChar := nextSeg.Text[0]

					// Wenn letztes Zeichen ein Satzzeichen [,?!:;.] ist und nächstes Zeichen kein Leerzeichen
					if strings.ContainsAny(string(lastChar), ",?!:;.") &&
						nextChar != ' ' && nextChar != '\n' && nextChar != '\r' && nextChar != '\t' {

						// Ausnahmen wie bei FixPunctuationSpacing (vereinfacht für Grenzen)
						isException := false
						if lastChar == '.' && nextChar >= '0' && nextChar <= '9' {
							isException = true
						}
						if lastChar == ':' && (nextChar == '/' || nextChar == '\\') {
							isException = true
						}

						if !isException {
							seg.Text += " "
						}
					}
				}
			}
		}
	}
}

// renderBlock entscheidet anhand des Typs des Blocks, welche Rendering-Funktion aufgerufen wird.
func (g *Generator) renderBlock(block blocks.DocBlock, isMeasurement bool) {
	// Wir klonen den Block nicht, um Speicher zu sparen, aber wir müssen vorsichtig sein,
	// da wir h.Text in renderHeading temporär verändern (durch fixPunctuationSpacing).
	// Da renderAll zweimal aufgerufen wird (einmal für Messung, einmal für PDF),
	// könnte fixPunctuationSpacing doppelt angewendet werden.
	// Das ist bei unserer Regex-Logik aber unproblematisch (idempotent).

	switch b := block.(type) {
	case blocks.HeadingBlock:
		g.renderHeading(b, isMeasurement)
	case blocks.ParagraphBlock:
		g.renderParagraph(b)
	case blocks.CodeBlock:
		g.renderCode(b)
	case blocks.ImageBlock:
		g.renderImage(b)
	case blocks.MermaidBlock:
		// Mermaid-Blöcke wurden bereits im Builder zu PNGs umgewandelt.
	case blocks.ListBlock:
		g.renderList(b)
	case blocks.TableBlock:
		g.renderTable(b)
	case blocks.BlockquoteBlock:
		g.renderBlockquote(b)
	case blocks.PageBreakBlock:
		g.pdf.AddPage()
	}
}

// safeSetFont setzt die Schriftart sicher und fällt auf Fallback-Schriften zurück, falls die gewünschte fehlt.
func (g *Generator) safeSetFont(family string, style string, size float64) {
	family = strings.ToLower(family)
	key := family
	if style != "" {
		key = family + style
	}

	if g.registeredFonts[key] {
		g.pdf.SetFont(family, style, size)
		g.currentFontIsUTF8 = true
	} else if style != "" && g.registeredFonts["main"+style] {
		// Fallback: Versuche den Style mit dem main-Font
		g.pdf.SetFont("main", style, size)
		g.currentFontIsUTF8 = true
	} else if g.registeredFonts["main"] {
		// Letzter Fallback: main ohne Style
		g.pdf.SetFont("main", "", size)
		g.currentFontIsUTF8 = true
	} else {
		g.pdf.SetFont("Arial", style, size)
		g.currentFontIsUTF8 = false
	}
}

// renderHeading rendert eine Überschrift mit automatischer Nummerierung und Inhaltsverzeichniseintrag.
// Überschriften mit ExcludeFromTOC=true werden ohne Nummerierung gerendert und nicht im TOC angezeigt.
func (g *Generator) renderHeading(h blocks.HeadingBlock, isMeasurement bool) {
	// Heading-Zähler nur erhöhen wenn die Überschrift NICHT vom TOC ausgeschlossen ist
	// So wird die Nummerierung für normale Überschriften nicht beeinflusst
	if !h.ExcludeFromTOC {
		if h.Level > 0 && h.Level <= len(g.headingCounts) {
			g.headingCounts[h.Level-1]++
			for i := h.Level; i < len(g.headingCounts); i++ {
				g.headingCounts[i] = 0
			}
		}
	}

	numbering := ""
	text := util.FixPunctuationSpacing(h.Text)

	// Keine Nummerierung für ausgeschlossene Überschriften
	if !h.ExcludeFromTOC {
		// Versuche Nummerierung aus dem Text zu extrahieren
		extractedNum, remainingText := splitNumberingAndText(text)

		if extractedNum != "" {
			// Wenn im Text eine Nummerierung gefunden wurde, verwenden wir diese bevorzugt
			numbering = extractedNum
			text = remainingText
		} else if g.cfg.Layout.HeaderNumbering {
			// Ansonsten automatische Nummerierung, falls aktiviert
			if h.ParentNumbering != "" {
				numbering = h.ParentNumbering
				// Wir verwenden Punkte für die Hierarchie (1.1.1)
				if !strings.HasSuffix(numbering, ".") && numbering != "" {
					numbering += "."
				}

				// Wir hängen nur an, wenn ParentNumbering nicht schon die Nummer für diesen Header ist
				// In der neuen Logik ist ParentNumbering oft die Nummer des ersten Headers der Datei
				if h.Level > 1 {
					for i := 1; i < h.Level; i++ {
						count := g.headingCounts[i]
						if count == 0 {
							count = 1
						}
						numbering += fmt.Sprintf("%d", count)
						if i < h.Level-1 {
							numbering += "."
						}
					}
				}
				if numbering != "" && !strings.HasSuffix(numbering, ".") {
					numbering += "."
				}
			} else {
				// Klassische hierarchische Nummerierung
				for i := 0; i < h.Level; i++ {
					count := g.headingCounts[i]
					if count == 0 {
						count = 1
					}
					numbering += fmt.Sprintf("%d.", count)
				}
			}
		}
	}

	if numbering != "" && !strings.HasSuffix(numbering, " ") {
		numbering += " "
	}

	link := g.pdf.AddLink()

	// Anchor-Link registrieren für interne Verlinkungen
	if h.AnchorID != "" {
		g.anchorLinks[h.AnchorID] = link
	}

	// Nur zum TOC hinzufügen wenn nicht ausgeschlossen
	if isMeasurement && !h.ExcludeFromTOC {
		// Berechne globales Level für korrekte Einrückung im TOC
		globalLevel := h.Level
		if h.ParentNumbering != "" {
			// Zähle Punkte in ParentNumbering für die Basistiefe
			dots := strings.Count(h.ParentNumbering, ".")
			globalLevel += dots
		}

		g.toc = append(g.toc, TOCEntry{
			Level:    globalLevel,
			Number:   numbering,
			Text:     text,
			Page:     g.pdf.PageNo(),
			Link:     link,
			AnchorID: h.AnchorID,
		})
	}
	g.pdf.SetLink(link, g.pdf.GetY(), -1)

	size := 14.0
	spacing := 3.0
	if h.Level == 1 {
		size = 22.0
		spacing = 10.0
	} else if h.Level == 2 {
		size = 18.0
		spacing = 5.0
	}

	// Bessere Seitenumbrüche für Überschriften (verhindert Orphan-Headings)
	g.checkPageBreak(size + spacing + 20)

	g.pdf.Ln(spacing)
	g.safeSetFont("main", "B", size)
	r, green, b := hexToRGB(g.cfg.Colors.Title)
	g.pdf.SetTextColor(r, green, b)

	if h.Level == 1 {
		left, _, _, _ := g.pdf.GetMargins()
		accentR, accentG, accentB := r, green, b
		if g.cfg.Colors.Accent != "" {
			accentR, accentG, accentB = hexToRGB(g.cfg.Colors.Accent)
		}
		g.pdf.SetFillColor(accentR, accentG, accentB)
		g.pdf.Rect(left, g.pdf.GetY()+2, 2, 10, "F")
		g.pdf.SetX(left + 5)
	}

	displayText := numbering + text
	g.pdf.MultiCell(0, 10, g.prepareText(displayText), "", "L", false)
	g.pdf.Ln(3)
}

// splitNumberingAndText extrahiert eine führende Nummerierung und den restlichen Text.
func splitNumberingAndText(s string) (string, string) {
	s = strings.TrimSpace(s)
	// Wir suchen nach einem Muster am Anfang: Zahlen, Punkte, gefolgt von Leerzeichen
	i := 0
	foundDigit := false
	for i < len(s) {
		r := s[i]
		if r >= '0' && r <= '9' {
			foundDigit = true
			i++
		} else if r == '.' || r == ')' {
			i++
		} else if r == ' ' || r == '\t' {
			// Ende der Nummerierung erreicht
			if foundDigit {
				return s[:i], strings.TrimSpace(s[i:])
			}
			break
		} else {
			break
		}
	}
	return "", s
}

// trimLeadingNumbering entfernt führende Nummern wie "1. ", "1.1 " oder "1) " aus einem String.
func trimLeadingNumbering(s string) string {
	_, text := splitNumberingAndText(s)
	return text
}

// safeWrite schreibt Text sicher in das PDF und fängt Panics der PDF-Bibliothek ab.
func (g *Generator) safeWrite(size float64, text string, family string, style string, link string) {
	g.safeWriteWithFontSize(size, text, family, style, link, g.cfg.FontSize)
}

// safeWriteWithFontSize schreibt Text sicher in das PDF mit einer spezifischen Schriftgröße.
func (g *Generator) safeWriteWithFontSize(size float64, text string, family string, style string, link string, fontSize float64) {
	if text == "" {
		return
	}

	g.safeSetFont(family, style, fontSize)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in gofpdf abgefangen (Write): %v (Text: %s)\n", r, text)
		}
	}()

	if link != "" {
		// Prüfen ob es ein Anchor-Link ist (beginnt mit #)
		if strings.HasPrefix(link, "#") {
			anchorID := strings.TrimPrefix(link, "#")
			if linkID, ok := g.anchorLinks[anchorID]; ok {
				// Interner Link zu einer Überschrift
				g.pdf.WriteLinkID(size, g.prepareText(text), linkID)
			} else {
				// Anchor nicht gefunden, als normalen Text rendern
				g.pdf.Write(size, g.prepareText(text))
			}
		} else {
			// Externer Link (URL)
			g.pdf.WriteLinkString(size, g.prepareText(text), link)
		}
	} else {
		g.pdf.Write(size, g.prepareText(text))
	}
}

// safeWriteLinkID schreibt einen klickbaren Link sicher in das PDF.
func (g *Generator) safeWriteLinkID(size float64, text string, family string, style string, link int) {
	if text == "" {
		return
	}

	g.safeSetFont(family, style, g.cfg.FontSize)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in gofpdf abgefangen (WriteLinkID): %v (Text: %s)\n", r, text)
		}
	}()

	g.pdf.WriteLinkID(size, g.prepareText(text), link)
}

// getLineHeight berechnet die Zeilenhöhe basierend auf Schriftgröße und konfiguriertem Zeilenabstand.
func (g *Generator) getLineHeight() float64 {
	spacing := g.cfg.Layout.LineSpacing
	if spacing <= 0 {
		spacing = 1.0
	}
	lh := g.cfg.FontSize * 0.5 * spacing
	if lh < 5 {
		lh = 5
	}
	return lh
}

// renderParagraph rendert einen Textabsatz mit Unterstützung für Fett, Kursiv, Durchgestrichen und Inline-Code.
func (g *Generator) renderParagraph(p blocks.ParagraphBlock) {
	g.safeSetFont("main", "", g.cfg.FontSize)
	g.setPrimaryTextColor()

	lineHeight := g.getLineHeight()

	// Prüfen, ob der Absatz Formatierungen enthält
	hasFormatting := false
	fullText := ""
	for _, seg := range p.Content {
		if seg.Bold || seg.Italic || seg.Strikethrough || seg.Code || seg.Link != "" {
			hasFormatting = true
		}
		fullText += seg.Text
	}

	// Wenn keine Formatierung vorhanden ist, nutzen wir MultiCell für die konfigurierte Ausrichtung
	if !hasFormatting {
		align := g.getAlign(g.cfg.Layout.Body)
		// Wir verwenden echten Blocksatz für Paragraphen, wenn "justify" eingestellt ist
		g.pdf.MultiCell(0, lineHeight, g.prepareText(fullText), "", align, false)
		g.pdf.Ln(2)
		return
	}

	g.fixSegmentSpacing(p.Content)
	for _, seg := range p.Content {
		style := ""
		if seg.Bold {
			style += "B"
		}
		if seg.Italic {
			style += "I"
		}

		if seg.Code {
			fontFamily := "main"
			if g.cfg.Fonts.Mono != "" {
				fontFamily = "mono"
			}
			// Inline-Code verwendet eine etwas kleinere Schriftgröße als normaler Text
			codeFontSize := g.cfg.Code.FontSize
			if codeFontSize <= 0 {
				codeFontSize = g.cfg.FontSize * 0.9 // 90% der normalen Schriftgröße für bessere Lesbarkeit
			}
			g.safeSetFont(fontFamily, "", codeFontSize)

			if seg.Text != "" {
				// Berechne Textbreite für den Hintergrund-Chip
				textWidth := g.pdf.GetStringWidth(g.prepareText(seg.Text))
				chipPaddingH := 2.5              // Horizontales Padding (links und rechts)
				chipPaddingV := 1.2              // Vertikales Padding (oben und unten)
				chipHeight := codeFontSize * 0.5 // Höhe basierend auf Schriftgröße

				// Position merken
				startX := g.pdf.GetX()
				startY := g.pdf.GetY()

				// Moderner Hintergrund-Chip (sanftes Grau, ohne sichtbaren Rahmen)
				g.pdf.SetFillColor(243, 244, 246)                          // Sanftes Grau (wie Tailwind gray-100)
				chipY := startY + (lineHeight-chipHeight-2*chipPaddingV)/2 // Vertikal zentrieren
				g.pdf.RoundedRect(startX, chipY, textWidth+2*chipPaddingH, chipHeight+2*chipPaddingV, 1.5, "1234", "F")

				// Text auf dem Chip schreiben
				g.pdf.SetX(startX + chipPaddingH)
				g.pdf.SetTextColor(55, 65, 81) // Dunkles Grau für Code-Text (wie Tailwind gray-700)
				codeLineHeight := codeFontSize * 0.5
				g.safeWriteWithFontSize(codeLineHeight, seg.Text, fontFamily, "", seg.Link, codeFontSize)

				// Position nach dem Chip setzen (mit etwas Abstand)
				g.pdf.SetX(startX + textWidth + 2*chipPaddingH + 1.5)

				// Textfarbe zurücksetzen für nachfolgenden Text
				g.setPrimaryTextColor()
			}
		} else {
			g.safeSetFont("main", style, g.cfg.FontSize)
			if seg.Text != "" {
				// Strikethrough: Position vor dem Text merken
				startX := g.pdf.GetX()
				startY := g.pdf.GetY()

				g.safeWrite(lineHeight, seg.Text, "main", style, seg.Link)

				// Strikethrough-Linie zeichnen
				if seg.Strikethrough {
					endX := g.pdf.GetX()
					// Linie in der Mitte des Textes zeichnen
					strikeY := startY + g.cfg.FontSize*0.35
					g.pdf.SetDrawColor(0, 0, 0)
					g.pdf.SetLineWidth(0.3)
					g.pdf.Line(startX, strikeY, endX, strikeY)
				}
			}
		}
	}
	g.pdf.Ln(lineHeight + 2)
}

// calculateCodeFontSize berechnet die optimale Schriftgröße für einen Code-Block
// basierend auf Zeilenanzahl, maximaler Zeilenlänge und verfügbarer Seitenhöhe.
func (g *Generator) calculateCodeFontSize(lineCount int, maxLineLen int) float64 {
	// Basis-Schriftgröße ermitteln - Standard ist 6 für sehr kompakte Code-Darstellung
	baseFontSize := 6.0
	if g.cfg.Code.FontSize > 0 {
		baseFontSize = g.cfg.Code.FontSize
	}

	// Wenn AutoScale deaktiviert ist, Basis-Schriftgröße zurückgeben
	if !g.cfg.Code.AutoScale {
		return baseFontSize
	}

	// Standardwerte für Schwellenwerte
	maxLen := 80
	minFontSize := 4.0 // Reduziert von 6.0 für sehr lange Code-Blöcke

	if g.cfg.Code.MaxLineLen > 0 {
		maxLen = g.cfg.Code.MaxLineLen
	}
	if g.cfg.Code.MinFontSize > 0 {
		minFontSize = g.cfg.Code.MinFontSize
	}

	fontSize := baseFontSize

	// Berechne verfügbare Seitenhöhe
	_, top, _, bottom := g.pdf.GetMargins()
	_, pageHeight := g.pdf.GetPageSize()
	availableHeight := pageHeight - top - bottom - 30 // 30 für Header/Footer und Padding

	// Berechne benötigte Höhe bei aktueller Schriftgröße
	// lineHeight = fontSize * 0.35, rectHeight = lineCount * lineHeight + 10
	neededHeight := float64(lineCount)*(fontSize*0.35) + 10

	// Skalierung basierend auf Seitenhöhe (wichtigste Anpassung)
	if neededHeight > availableHeight {
		// Berechne die maximale Schriftgröße, die auf die Seite passt
		// availableHeight = lineCount * (fontSize * 0.35) + 10
		// availableHeight - 10 = lineCount * fontSize * 0.35
		// fontSize = (availableHeight - 10) / (lineCount * 0.35)
		maxFontForHeight := (availableHeight - 10) / (float64(lineCount) * 0.35)
		if maxFontForHeight < fontSize {
			fontSize = maxFontForHeight
		}
	}

	// Skalierung basierend auf Zeilenlänge
	if maxLineLen > maxLen {
		// Berechne verfügbare Breite
		left, _, right, _ := g.pdf.GetMargins()
		w, _ := g.pdf.GetPageSize()
		availableWidth := w - left - right - 10 // 10 für Padding

		// Schätze benötigte Breite bei aktueller Schriftgröße
		charWidth := fontSize * 0.5 // Ungefähre Zeichenbreite für Monospace
		neededWidth := float64(maxLineLen) * charWidth

		if neededWidth > availableWidth {
			lenRatio := availableWidth / neededWidth
			newFontSize := fontSize * lenRatio
			if newFontSize < fontSize {
				fontSize = newFontSize
			}
		}
	}

	// Minimale Schriftgröße einhalten
	if fontSize < minFontSize {
		fontSize = minFontSize
	}

	return fontSize
}

// coloredSegment repräsentiert ein Textsegment mit Farbinformation für Syntax-Highlighting.
type coloredSegment struct {
	text  string
	color string
}

// renderCode rendert einen Codeblock mit Syntax-Highlighting und abgerundeten Ecken.
// Bei sehr langen Code-Blöcken wird der Code automatisch auf mehrere Seiten aufgeteilt.
func (g *Generator) renderCode(c blocks.CodeBlock) {
	fontFamily := "main"
	if g.cfg.Fonts.Mono != "" {
		fontFamily = "mono"
	}

	bgR, bgG, bgB := 245, 245, 245
	if c.BgColor != "" {
		r, green, b := hexToRGB(c.BgColor)
		if r < 250 || green < 250 || b < 250 {
			bgR, bgG, bgB = r, green, b
		}
	}

	// Alle Code-Zeilen extrahieren mit Segmenten für Syntax-Highlighting
	var allLines [][]coloredSegment
	var currentLine []coloredSegment

	for _, seg := range c.Segments {
		segColor := seg.Color
		// Bereinige den Text von problematischen Unicode-Zeichen
		text := cleanCodeText(seg.Text)
		currentText := ""

		for _, r := range text {
			if r == '\n' {
				// Aktuelles Segment zur Zeile hinzufügen (falls Text vorhanden)
				if currentText != "" {
					currentLine = append(currentLine, coloredSegment{text: currentText, color: segColor})
					currentText = ""
				}
				// Zeile abschließen
				allLines = append(allLines, currentLine)
				currentLine = nil
			} else {
				currentText += string(r)
			}
		}
		// Restlichen Text als Segment hinzufügen
		if currentText != "" {
			currentLine = append(currentLine, coloredSegment{text: currentText, color: segColor})
		}
	}
	// Letzte Zeile hinzufügen
	if len(currentLine) > 0 || len(allLines) == 0 {
		allLines = append(allLines, currentLine)
	}

	lineCount := len(allLines)
	if lineCount == 0 {
		lineCount = 1
	}

	// Maximale Zeilenlänge ermitteln
	maxLineLen := 0
	for _, line := range allLines {
		lineLen := 0
		for _, seg := range line {
			lineLen += len(seg.text)
		}
		if lineLen > maxLineLen {
			maxLineLen = lineLen
		}
	}

	// Verfügbare Seitenhöhe berechnen
	_, top, _, bottom := g.pdf.GetMargins()
	_, pageHeight := g.pdf.GetPageSize()
	availableHeight := pageHeight - top - bottom - 40 // Platz für Header/Footer

	// Schriftgröße berechnen (mit AutoScale)
	codeFontSize := g.calculateCodeFontSize(lineCount, maxLineLen)
	lineHeight := codeFontSize * 0.35 // Reduzierter Zeilenabstand für kompaktere Darstellung

	// Berechne wie viele Zeilen auf eine Seite passen
	linesPerPage := int((availableHeight - 20) / lineHeight) // 20 für Padding im Rechteck
	if linesPerPage < 5 {
		linesPerPage = 5 // Mindestens 5 Zeilen pro Seite
	}

	// Seitenränder und Breite
	left, _, right, _ := g.pdf.GetMargins()
	w, _ := g.pdf.GetPageSize()
	width := w - left - right

	// Code in Chunks aufteilen und rendern
	totalChunks := (lineCount + linesPerPage - 1) / linesPerPage
	if totalChunks < 1 {
		totalChunks = 1
	}

	for chunkIdx := 0; chunkIdx < totalChunks; chunkIdx++ {
		startLine := chunkIdx * linesPerPage
		endLine := startLine + linesPerPage
		if endLine > lineCount {
			endLine = lineCount
		}
		chunkLines := allLines[startLine:endLine]
		chunkLineCount := len(chunkLines)

		// Rechteckhöhe für diesen Chunk
		rectHeight := float64(chunkLineCount)*lineHeight + 10

		// Seitenumbruch prüfen
		g.checkPageBreak(rectHeight + 10)

		g.pdf.SetFillColor(bgR, bgG, bgB)
		g.pdf.SetDrawColor(200, 200, 200)

		x := g.pdf.GetX()
		y := g.pdf.GetY()

		// Rechteck zeichnen
		g.pdf.RoundedRect(x, y, width, rectHeight, 4, "1234", "DF")

		// Sprach-Label (nur beim ersten Chunk) oder Fortsetzungsmarkierung
		if chunkIdx == 0 && c.Language != "" {
			g.safeSetFont("main", "B", 7)
			g.pdf.SetTextColor(150, 150, 150)
			labelW := g.pdf.GetStringWidth(c.Language) + 4
			g.pdf.SetXY(x+width-labelW-2, y+2)
			g.pdf.CellFormat(labelW, 4, g.prepareText(c.Language), "", 0, "R", false, 0, "")
		} else if chunkIdx > 0 {
			// Fortsetzungsmarkierung
			g.safeSetFont("main", "I", 6)
			g.pdf.SetTextColor(150, 150, 150)
			contLabel := fmt.Sprintf("... (Teil %d/%d)", chunkIdx+1, totalChunks)
			labelW := g.pdf.GetStringWidth(contLabel) + 4
			g.pdf.SetXY(x+width-labelW-2, y+2)
			g.pdf.CellFormat(labelW, 4, g.prepareText(contLabel), "", 0, "R", false, 0, "")
		}

		// Code-Zeilen rendern mit Syntax-Highlighting
		g.safeSetFont(fontFamily, "", codeFontSize)
		g.pdf.SetX(x + 5)
		g.pdf.SetY(y + 5)

		for i, line := range chunkLines {
			// Jedes Segment der Zeile mit seiner eigenen Farbe rendern
			for _, seg := range line {
				if seg.color != "" {
					r, green, b := hexToRGB(seg.color)
					g.pdf.SetTextColor(r, green, b)
				} else {
					g.setPrimaryTextColor()
				}

				if seg.text != "" {
					g.safeWriteWithFontSize(lineHeight, seg.text, fontFamily, "", "", codeFontSize)
				}
			}

			if i < len(chunkLines)-1 {
				g.pdf.Ln(lineHeight)
				g.pdf.SetX(x + 5)
			}
		}

		g.pdf.SetY(y + rectHeight + 5)
		g.pdf.Ln(2)
	}
}

// renderImage rendert ein Bild mit automatischer Skalierung und optionalem Titel.
// Unterstützt konfigurierbare Breite und Skalierung über ImageBlock.Width und ImageBlock.Scale.
func (g *Generator) renderImage(i blocks.ImageBlock) {
	g.pdf.RegisterImage(i.Path, "")
	info := g.pdf.GetImageInfo(i.Path)
	left, top, right, bottom := g.pdf.GetMargins()
	w, h_page := g.pdf.GetPageSize()
	maxWidth := w - left - right
	maxPageHeight := h_page - top - bottom - 40

	// Berechne Bildbreite basierend auf Konfiguration
	var widthOnPage float64
	if i.Width > 0 {
		// Explizite Breite angegeben
		widthOnPage = i.Width
		if widthOnPage > maxWidth-20 {
			widthOnPage = maxWidth - 20
		}
	} else if i.Scale > 0 && i.Scale != 1.0 {
		// Skalierungsfaktor angegeben
		widthOnPage = (maxWidth - 20) * i.Scale
	} else {
		// Standard: fast volle Breite
		widthOnPage = maxWidth - 20
	}

	var h float64
	if info != nil && info.Width() > 0 {
		h = (info.Height() / info.Width()) * widthOnPage
	} else {
		h = 60.0
	}

	titleHeight := 0.0
	if i.Title != "" {
		titleHeight = 10.0
	}

	if h+titleHeight > maxPageHeight {
		h = maxPageHeight - titleHeight
		widthOnPage = 0
	}

	padding := 5.0
	imgW := widthOnPage
	if imgW == 0 && info != nil {
		imgW = (info.Width() / info.Height()) * h
	}
	containerH := h + 2*padding
	containerW := imgW + 2*padding

	g.checkPageBreak(containerH + titleHeight + 10)

	x := g.pdf.GetX()
	y := g.pdf.GetY()

	if i.Title != "" {
		g.safeSetFont("main", "B", 10)
		g.pdf.SetTextColor(100, 100, 100)
		g.pdf.CellFormat(0, 8, g.prepareText(i.Title), "", 1, "C", false, 0, "")
		g.pdf.Ln(2)
	}

	imgX := x + (maxWidth-containerW)/2

	g.pdf.SetFillColor(255, 255, 255)
	if g.cfg.Colors.Background != "" {
		g.pdf.SetFillColor(250, 250, 250)
	}
	g.pdf.SetDrawColor(220, 220, 220)
	g.pdf.RoundedRect(imgX, g.pdf.GetY(), containerW, containerH, 5, "1234", "DF")

	g.pdf.Image(i.Path, imgX+padding, g.pdf.GetY()+padding, imgW, h, false, "", 0, "")

	g.pdf.SetY(y + containerH + titleHeight + 5)
	g.pdf.Ln(2)
}

// renderList rendert eine Aufzählung oder nummerierte Liste.
func (g *Generator) renderList(l blocks.ListBlock) {
	g.renderListWithIndent(l, 0)
	g.pdf.Ln(5)
}

// renderListWithIndent rendert eine Liste mit der angegebenen Einrückungsebene.
func (g *Generator) renderListWithIndent(l blocks.ListBlock, indentLevel int) {
	g.safeSetFont("main", "", g.cfg.FontSize)
	g.setPrimaryTextColor()
	lineHeight := g.getLineHeight()

	// Basis-Einrückung entspricht dem linken Margin für konsistente Ausrichtung mit normalem Text
	left, _, _, _ := g.pdf.GetMargins()
	baseIndent := left
	indentStep := 8.0 // Einrückung pro Verschachtelungsebene
	currentIndent := baseIndent + float64(indentLevel)*indentStep

	for i, item := range l.Items {
		g.fixSegmentSpacing(item.Content)

		prefix := "• "
		if l.Ordered {
			prefix = fmt.Sprintf("%d. ", i+1)
		}

		// Prüfen, ob der Listeneintrag Formatierungen enthält
		hasFormatting := false
		fullText := prefix
		for _, seg := range item.Content {
			if seg.Bold || seg.Italic || seg.Strikethrough || seg.Code || seg.Link != "" {
				hasFormatting = true
			}
			fullText += seg.Text
		}

		if !hasFormatting {
			align := g.getAlign(g.cfg.Layout.Body)
			g.pdf.SetX(currentIndent)
			// Wir berechnen die Breite für MultiCell unter Berücksichtigung der Einrückung
			w, _ := g.pdf.GetPageSize()
			_, _, right, _ := g.pdf.GetMargins()
			width := w - currentIndent - right
			g.pdf.MultiCell(width, lineHeight, g.prepareText(fullText), "", align, false)
			g.pdf.Ln(1)
		} else {
			g.pdf.SetX(currentIndent)
			g.pdf.Write(lineHeight, g.prepareText(prefix))

			for _, seg := range item.Content {
				if seg.Code {
					// Inline-Code mit Hintergrund-Chip rendern
					fontFamily := "main"
					if g.cfg.Fonts.Mono != "" {
						fontFamily = "mono"
					}
					codeFontSize := g.cfg.Code.FontSize
					if codeFontSize <= 0 {
						codeFontSize = g.cfg.FontSize * 0.9 // 90% der normalen Schriftgröße für bessere Lesbarkeit
					}
					g.safeSetFont(fontFamily, "", codeFontSize)

					if seg.Text != "" {
						textWidth := g.pdf.GetStringWidth(g.prepareText(seg.Text))
						chipPaddingH := 2.5              // Horizontales Padding (links und rechts)
						chipPaddingV := 1.2              // Vertikales Padding (oben und unten)
						chipHeight := codeFontSize * 0.5 // Höhe basierend auf Schriftgröße

						startX := g.pdf.GetX()
						startY := g.pdf.GetY()

						// Moderner Hintergrund-Chip (sanftes Grau, ohne sichtbaren Rahmen)
						g.pdf.SetFillColor(243, 244, 246)                          // Sanftes Grau (wie Tailwind gray-100)
						chipY := startY + (lineHeight-chipHeight-2*chipPaddingV)/2 // Vertikal zentrieren
						g.pdf.RoundedRect(startX, chipY, textWidth+2*chipPaddingH, chipHeight+2*chipPaddingV, 1.5, "1234", "F")

						g.pdf.SetX(startX + chipPaddingH)
						g.pdf.SetTextColor(55, 65, 81) // Dunkles Grau für Code-Text (wie Tailwind gray-700)
						codeLineHeight := codeFontSize * 0.5
						g.safeWriteWithFontSize(codeLineHeight, seg.Text, fontFamily, "", seg.Link, codeFontSize)

						g.pdf.SetX(startX + textWidth + 2*chipPaddingH + 1.5)

						// Textfarbe zurücksetzen für nachfolgenden Text
						g.setPrimaryTextColor()
					}
				} else {
					style := ""
					if seg.Bold {
						style += "B"
					}
					if seg.Italic {
						style += "I"
					}
					g.safeSetFont("main", style, g.cfg.FontSize)

					// Strikethrough: Position vor dem Text merken
					startX := g.pdf.GetX()
					startY := g.pdf.GetY()

					g.safeWrite(lineHeight, seg.Text, "main", style, seg.Link)

					// Strikethrough-Linie zeichnen
					if seg.Strikethrough {
						endX := g.pdf.GetX()
						strikeY := startY + g.cfg.FontSize*0.35
						g.pdf.SetDrawColor(0, 0, 0)
						g.pdf.SetLineWidth(0.3)
						g.pdf.Line(startX, strikeY, endX, strikeY)
					}
				}
			}
			g.pdf.Ln(lineHeight + 2)
		}

		// Verschachtelte Liste rendern
		if item.SubList != nil {
			g.renderListWithIndent(*item.SubList, indentLevel+1)
		}
	}
}

// renderBlockquote rendert ein Zitat mit linkem Rand und Einrückung.
func (g *Generator) renderBlockquote(b blocks.BlockquoteBlock) {
	// Speichere aktuelle Position
	left, _, _, _ := g.pdf.GetMargins()
	startY := g.pdf.GetY()

	// Einrückung für Blockquote
	quoteIndent := 10.0
	borderWidth := 2.0
	borderColor := 150 // Grau

	// Setze linken Rand für den Inhalt
	g.pdf.SetLeftMargin(left + quoteIndent + borderWidth + 3)
	g.pdf.SetX(left + quoteIndent + borderWidth + 3)

	// Rendere den Inhalt des Blockquotes
	for _, block := range b.Content {
		switch content := block.(type) {
		case blocks.ParagraphBlock:
			// Blockquote-Text kursiv rendern
			g.safeSetFont("main", "I", g.cfg.FontSize)
			g.setPrimaryTextColor()
			lineHeight := g.getLineHeight()

			fullText := ""
			for _, seg := range content.Content {
				fullText += seg.Text
			}
			align := g.getAlign(g.cfg.Layout.Body)
			g.pdf.MultiCell(0, lineHeight, g.prepareText(fullText), "", align, false)
			g.pdf.Ln(2)
		case blocks.ListBlock:
			g.renderList(content)
		}
	}

	endY := g.pdf.GetY()

	// Zeichne den linken Rand
	g.pdf.SetDrawColor(borderColor, borderColor, borderColor)
	g.pdf.SetLineWidth(borderWidth)
	g.pdf.Line(left+quoteIndent, startY, left+quoteIndent, endY)

	// Setze Margins zurück
	g.pdf.SetLeftMargin(left)
	g.pdf.Ln(5)
}

// renderTable rendert eine Tabelle mit Kopfzeile und automatischer Spaltenbreite.
// Verbesserte Darstellung mit schöneren Rahmen, Padding und Zebra-Streifen.
func (g *Generator) renderTable(t blocks.TableBlock) {
	if len(t.Rows) == 0 {
		return
	}

	left, _, right, _ := g.pdf.GetMargins()
	w, pageHeight := g.pdf.GetPageSize()
	width := w - left - right
	colCount := 0
	for _, row := range t.Rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}
	if colCount == 0 {
		return
	}

	// Tabellen-Styling Konstanten
	cellPadding := 4.0
	headerBgR, headerBgG, headerBgB := 52, 73, 94 // Elegantes Dunkelblau für Header (Fallback)
	if g.cfg.Colors.Accent != "" {
		headerBgR, headerBgG, headerBgB = hexToRGB(g.cfg.Colors.Accent)
	}
	headerTextR, headerTextG, headerTextB := 255, 255, 255 // Weißer Text für Header
	evenRowR, evenRowG, evenRowB := 245, 247, 250          // Sehr helles Blau-Grau für Zebra
	oddRowR, oddRowG, oddRowB := 255, 255, 255             // Weiß
	borderR, borderG, borderB := 200, 200, 210             // Dezenter Rahmen

	// Berechne dynamische Spaltenbreiten
	colWidths := make([]float64, colCount)
	g.safeSetFont("main", "B", g.cfg.FontSize)

	// 1. Berechne benötigte Breite für jede Spalte
	for _, row := range t.Rows {
		for i, cell := range row {
			if i >= colCount {
				break
			}
			cellText := ""
			for _, seg := range cell.Content {
				cellText += seg.Text
			}
			// Breite des Textes in einer Zeile messen
			textWidth := g.pdf.GetStringWidth(cellText) + 2*cellPadding + 4
			if textWidth > colWidths[i] {
				colWidths[i] = textWidth
			}
		}
	}

	// 2. Proportionale Anpassung an die Gesamtbreite
	totalNeededWidth := 0.0
	for _, cw := range colWidths {
		totalNeededWidth += cw
	}

	// Tabellen sollten immer die volle Breite nutzen, wenn sie nicht winzig sind.
	// Das sorgt für ein konsistentes Look & Feel.
	if totalNeededWidth > 0 {
		scaleFactor := width / totalNeededWidth
		// Wenn die Tabelle natürlich sehr schmal wäre, skalieren wir sie nur moderat,
		// es sei denn, der User möchte volle Breite (was bei Markdown-Tabellen meist erwartet wird).
		if totalNeededWidth < width*0.5 {
			// Begrenze das "Aufblasen" auf maximal 1.5x der natürlichen Breite,
			// damit schmale Tabellen nicht absurd breit werden.
			if scaleFactor > 1.5 {
				scaleFactor = 1.5
			}
		}

		for i := range colWidths {
			colWidths[i] *= scaleFactor
		}
	}

	// Sicherstellen, dass die Gesamtbreite nicht das Limit überschreitet
	currentTotal := 0.0
	for _, cw := range colWidths {
		currentTotal += cw
	}
	if currentTotal > width {
		shrink := width / currentTotal
		for i := range colWidths {
			colWidths[i] *= shrink
		}
	}

	// Falls nach dem Shrinken immer noch winzige Rundungsdifferenzen bestehen,
	// passen wir die letzte Spalte an.
	currentTotal = 0.0
	for _, cw := range colWidths {
		currentTotal += cw
	}
	if len(colWidths) > 0 && currentTotal < width-0.1 {
		colWidths[len(colWidths)-1] += (width - currentTotal)
	}

	g.setPrimaryTextColor()

	// Berechne die Zeilenhöhen basierend auf den neuen Spaltenbreiten
	rowHeights := make([]float64, len(t.Rows))
	totalTableHeight := 0.0
	lineHeightFactor := 1.2 // Faktor für Zeilenabstand innerhalb der Zelle
	for rowIdx, row := range t.Rows {
		maxH := 0.0
		for i, cell := range row {
			if i >= colCount {
				break
			}
			g.fixSegmentSpacing(cell.Content)
			cellText := ""
			for _, seg := range cell.Content {
				cellText += seg.Text
			}

			// Font für Messung setzen (Header ist Fett)
			if cell.Header {
				g.safeSetFont("main", "B", g.cfg.FontSize)
			} else {
				g.safeSetFont("main", "", g.cfg.FontSize)
			}

			lines := g.pdf.SplitLines([]byte(g.prepareText(cellText)), colWidths[i]-2*cellPadding)
			h := float64(len(lines)) * (g.cfg.FontSize * 0.35 * lineHeightFactor)
			if h > maxH {
				maxH = h
			}
		}
		maxH += 2 * cellPadding // Padding oben und unten
		// Mindesthöhe für eine Zeile
		minH := g.cfg.FontSize*0.35*lineHeightFactor + 2*cellPadding
		if maxH < minH {
			maxH = minH
		}
		rowHeights[rowIdx] = maxH
		totalTableHeight += maxH
	}
	totalTableHeight += 5 // Abstand nach der Tabelle

	// Berechne verfügbare Höhe auf der aktuellen Seite
	_, top, _, bottom := g.pdf.GetMargins()
	availableHeight := pageHeight - g.pdf.GetY() - bottom - 10

	// Seitenumbruch-Logik
	maxPageHeight := pageHeight - top - bottom - 20
	if totalTableHeight > availableHeight && totalTableHeight <= maxPageHeight {
		g.pdf.AddPage()
	}

	rowIndex := 0
	headerRow := -1
	for i, row := range t.Rows {
		if len(row) > 0 && row[0].Header {
			headerRow = i
			break
		}
	}

	for rowIdx, row := range t.Rows {
		maxH := rowHeights[rowIdx]
		isHeader := len(row) > 0 && row[0].Header

		// Prüfe auf Seitenumbruch innerhalb der Tabelle
		// Wir lassen etwas Puffer (10mm statt 5mm) am Seitenende für Stabilität
		if g.pdf.GetY()+maxH > pageHeight-bottom-10 {
			g.pdf.AddPage()
			// Wenn wir eine neue Seite anfangen, wiederholen wir den Header
			if headerRow != -1 && !isHeader {
				g.renderTableRow(t.Rows[headerRow], colWidths, rowHeights[headerRow], true, rowIndex, headerBgR, headerBgG, headerBgB, headerTextR, headerTextG, headerTextB, evenRowR, evenRowG, evenRowB, oddRowR, oddRowG, oddRowB, borderR, borderG, borderB, t.Alignments, cellPadding)
			}
		}

		g.renderTableRow(row, colWidths, maxH, isHeader, rowIndex, headerBgR, headerBgG, headerBgB, headerTextR, headerTextG, headerTextB, evenRowR, evenRowG, evenRowB, oddRowR, oddRowG, oddRowB, borderR, borderG, borderB, t.Alignments, cellPadding)

		if !isHeader {
			rowIndex++
		}
	}
	g.pdf.Ln(5)
}

// renderTableRow rendert eine einzelne Tabellenzeile (Hilfsfunktion für renderTable)
func (g *Generator) renderTableRow(row []blocks.TableRow, colWidths []float64, maxH float64, isHeader bool, rowIndex int, hBgR, hBgG, hBgB, hTextR, hTextG, hTextB, eRowR, eRowG, eRowB, oRowR, oRowG, oRowB, bR, bG, bB int, alignments []blocks.Align, cellPadding float64) {
	left, _, _, _ := g.pdf.GetMargins()
	startY := g.pdf.GetY()
	colCount := len(colWidths)

	// Zeichne zuerst die Hintergründe und Rahmen für alle Zellen der Zeile
	currentX := left
	for i := range colWidths {
		if i >= len(row) && !isHeader { // Falls die Zeile weniger Zellen hat als Spalten (sollte nicht sein)
			break
		}

		g.pdf.SetXY(currentX, startY)

		// Hintergrundfarbe setzen
		if isHeader {
			g.pdf.SetFillColor(hBgR, hBgG, hBgB)
		} else {
			if rowIndex%2 == 0 {
				g.pdf.SetFillColor(eRowR, eRowG, eRowB)
			} else {
				g.pdf.SetFillColor(oRowR, oRowG, oRowB)
			}
		}

		// Zelle mit Hintergrund zeichnen
		g.pdf.Rect(currentX, startY, colWidths[i], maxH, "F")

		// Rahmen zeichnen
		g.pdf.SetDrawColor(bR, bG, bB)
		g.pdf.SetLineWidth(0.15)
		g.pdf.Rect(currentX, startY, colWidths[i], maxH, "D")

		currentX += colWidths[i]
	}

	// Dann den Text in die Zellen rendern
	currentX = left
	for i, cell := range row {
		if i >= colCount {
			break
		}

		// Textfarbe und Font setzen
		if cell.Header || isHeader {
			g.pdf.SetTextColor(hTextR, hTextG, hTextB)
			g.safeSetFont("main", "B", g.cfg.FontSize)
		} else {
			g.setPrimaryTextColor()
			g.safeSetFont("main", "", g.cfg.FontSize)
		}

		cellText := ""
		for _, seg := range cell.Content {
			cellText += seg.Text
		}

		align := g.getAlign(g.cfg.Layout.Body)
		if i < len(alignments) {
			switch alignments[i] {
			case blocks.AlignCenter:
				align = "C"
			case blocks.AlignRight:
				align = "R"
			case blocks.AlignLeft:
				align = "L"
			}
		}

		// Vertikales Zentrieren
		lines := g.pdf.SplitLines([]byte(g.prepareText(cellText)), colWidths[i]-2*cellPadding)
		lineHeight := g.cfg.FontSize * 0.35 * 1.2
		textHeight := float64(len(lines)) * lineHeight
		verticalOffset := (maxH - textHeight) / 2
		if verticalOffset < cellPadding {
			verticalOffset = cellPadding
		}

		g.pdf.SetXY(currentX+cellPadding, startY+verticalOffset)

		// Wenn align "J" (Justify) ist, müssen wir sicherstellen, dass MultiCell
		// die Breite der Zelle abzüglich Padding nutzt.
		g.pdf.MultiCell(colWidths[i]-2*cellPadding, lineHeight, g.prepareText(cellText), "", align, false)

		currentX += colWidths[i]
	}

	// Nach der Zeile setzen wir den Y-Cursor absolut auf das Ende der Zeile
	g.pdf.SetXY(left, startY+maxH)
}
