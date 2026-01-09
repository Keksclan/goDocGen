// Package pdf enthält die Logik zum Rendern der verschiedenen Dokumentblöcke.
package pdf

import (
	"godocgen/internal/blocks"
	"fmt"
)

// renderBlock entscheidet anhand des Typs des Blocks, welche Rendering-Funktion aufgerufen wird.
func (g *Generator) renderBlock(block blocks.DocBlock, isMeasurement bool) {
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
	case blocks.PageBreakBlock:
		g.pdf.AddPage()
	}
}

// safeSetFont setzt die Schriftart sicher und fällt auf Fallback-Schriften zurück, falls die gewünschte fehlt.
func (g *Generator) safeSetFont(family string, style string, size float64) {
	key := family
	if style != "" {
		key = family + style
	}

	if g.registeredFonts[key] {
		g.pdf.SetFont(family, style, size)
	} else if g.registeredFonts["Main"] {
		g.pdf.SetFont("Main", "", size)
	} else {
		g.pdf.SetFont("Arial", style, size)
	}
}

// renderHeading rendert eine Überschrift mit automatischer Nummerierung und Inhaltsverzeichniseintrag.
func (g *Generator) renderHeading(h blocks.HeadingBlock, isMeasurement bool) {
	if h.Level > 0 && h.Level <= len(g.headingCounts) {
		g.headingCounts[h.Level-1]++
		for i := h.Level; i < len(g.headingCounts); i++ {
			g.headingCounts[i] = 0
		}
	}

	numbering := ""
	if g.cfg.Layout.HeaderNumbering {
		for i := 0; i < h.Level; i++ {
			count := g.headingCounts[i]
			if count == 0 {
				count = 1
			}
			numbering += fmt.Sprintf("%d.", count)
		}
		if numbering != "" {
			numbering += " "
		}
	}

	link := g.pdf.AddLink()
	if isMeasurement {
		g.toc = append(g.toc, TOCEntry{
			Level:  h.Level,
			Number: numbering,
			Text:   h.Text,
			Page:   g.pdf.PageNo(),
			Link:   link,
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
	g.safeSetFont("Main", "B", size)
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

	displayText := numbering + h.Text
	align := g.getAlign(g.cfg.Layout.Body)
	g.pdf.MultiCell(0, 10, displayText, "", align, false)
	g.pdf.Ln(3)
}

// safeWrite schreibt Text sicher in das PDF und fängt Panics der PDF-Bibliothek ab.
func (g *Generator) safeWrite(size float64, text string, family string, style string) {
	if text == "" {
		return
	}

	key := family
	if style != "" {
		key = family + style
	}

	if !g.registeredFonts[key] {
		if g.registeredFonts["Main"] {
			g.pdf.SetFont("Main", "", g.cfg.FontSize)
		} else {
			g.pdf.SetFont("Arial", style, g.cfg.FontSize)
		}
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in gofpdf abgefangen (Write): %v (Text: %s)\n", r, text)
		}
	}()

	g.pdf.Write(size, text)
}

// safeWriteLinkID schreibt einen klickbaren Link sicher in das PDF.
func (g *Generator) safeWriteLinkID(size float64, text string, family string, style string, link int) {
	if text == "" {
		return
	}

	key := family
	if style != "" {
		key = family + style
	}

	if !g.registeredFonts[key] {
		if g.registeredFonts["Main"] {
			g.pdf.SetFont("Main", "", g.cfg.FontSize)
		} else {
			g.pdf.SetFont("Arial", style, g.cfg.FontSize)
		}
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in gofpdf abgefangen (WriteLinkID): %v (Text: %s)\n", r, text)
		}
	}()

	g.pdf.WriteLinkID(size, text, link)
}

// renderParagraph rendert einen Textabsatz mit Unterstützung für Fett, Kursiv und Inline-Code.
func (g *Generator) renderParagraph(p blocks.ParagraphBlock) {
	g.safeSetFont("Main", "", g.cfg.FontSize)
	g.setPrimaryTextColor()

	for _, seg := range p.Content {
		style := ""
		if seg.Bold {
			style += "B"
		}
		if seg.Italic {
			style += "I"
		}

		if seg.Code {
			fontFamily := "Main"
			if g.cfg.Fonts.Mono != "" {
				fontFamily = "Mono"
			}
			g.safeSetFont(fontFamily, "I", g.cfg.FontSize)
			g.pdf.SetFillColor(240, 240, 240)
			if seg.Text != "" {
				g.safeWrite(g.cfg.FontSize/2, seg.Text, fontFamily, "I")
			}
		} else {
			g.safeSetFont("Main", style, g.cfg.FontSize)
			if seg.Text != "" {
				g.safeWrite(g.cfg.FontSize/2, seg.Text, "Main", style)
			}
		}
	}
	g.pdf.Ln(g.cfg.FontSize/2 + 5)
}

// renderCode rendert einen Codeblock mit Syntax-Highlighting und abgerundeten Ecken.
func (g *Generator) renderCode(c blocks.CodeBlock) {
	fontFamily := "Main"
	if g.cfg.Fonts.Mono != "" {
		fontFamily = "Mono"
	}

	g.safeSetFont(fontFamily, "I", g.cfg.FontSize)
	bgR, bgG, bgB := 245, 245, 245
	if c.BgColor != "" {
		r, green, b := hexToRGB(c.BgColor)
		if r < 250 || green < 250 || b < 250 {
			bgR, bgG, bgB = r, green, b
		}
	}
	g.pdf.SetFillColor(bgR, bgG, bgB)
	g.pdf.SetDrawColor(200, 200, 200)

	lineCount := 0
	for _, seg := range c.Segments {
		for _, r := range seg.Text {
			if r == '\n' {
				lineCount++
			}
		}
	}
	if lineCount == 0 && len(c.Segments) > 0 {
		lineCount = 1
	} else if len(c.Segments) > 0 && c.Segments[len(c.Segments)-1].Text != "" && c.Segments[len(c.Segments)-1].Text[len(c.Segments[len(c.Segments)-1].Text)-1] != '\n' {
		lineCount++
	}

	lineHeight := g.cfg.FontSize * 0.5
	rectHeight := float64(lineCount)*lineHeight + 10

	g.checkPageBreak(rectHeight + 10)

	x := g.pdf.GetX()
	y := g.pdf.GetY()
	left, _, right, _ := g.pdf.GetMargins()
	width := 210 - left - right

	g.pdf.RoundedRect(x, y, width, rectHeight, 4, "1234", "DF")

	if c.Language != "" {
		g.safeSetFont("Main", "B", 7)
		g.pdf.SetTextColor(150, 150, 150)
		labelW := g.pdf.GetStringWidth(c.Language) + 4
		g.pdf.SetXY(x+width-labelW-2, y+2)
		g.pdf.CellFormat(labelW, 4, c.Language, "", 0, "R", false, 0, "")
	}

	g.pdf.SetX(x + 5)
	g.pdf.SetY(y + 5)

	for _, seg := range c.Segments {
		if seg.Color != "" {
			r, green, b := hexToRGB(seg.Color)
			g.pdf.SetTextColor(r, green, b)
		} else {
			g.setPrimaryTextColor()
		}

		text := seg.Text
		for {
			idx := -1
			for i, r := range text {
				if r == '\n' {
					idx = i
					break
				}
			}

			if idx == -1 {
				if text != "" {
					g.safeWrite(lineHeight, text, fontFamily, "I")
				}
				break
			}

			if idx > 0 {
				g.safeWrite(lineHeight, text[:idx], fontFamily, "I")
			}
			g.pdf.Ln(lineHeight)
			g.pdf.SetX(x + 5)
			text = text[idx+1:]
			if text == "" {
				break
			}
		}
	}
	g.pdf.SetY(y + rectHeight + 5)
	g.pdf.Ln(2)
}

// renderImage rendert ein Bild mit automatischer Skalierung und optionalem Titel.
func (g *Generator) renderImage(i blocks.ImageBlock) {
	g.pdf.RegisterImage(i.Path, "")
	info := g.pdf.GetImageInfo(i.Path)
	left, top, right, bottom := g.pdf.GetMargins()
	maxWidth := 210 - left - right
	maxPageHeight := 297 - top - bottom - 40

	widthOnPage := maxWidth - 20
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
		g.pdf.SetFont("Main", "B", 10)
		g.pdf.SetTextColor(100, 100, 100)
		g.pdf.CellFormat(0, 8, i.Title, "", 1, "C", false, 0, "")
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
	g.pdf.SetFont("Main", "", g.cfg.FontSize)
	g.setPrimaryTextColor()
	for i, item := range l.Items {
		prefix := "• "
		if l.Ordered {
			prefix = fmt.Sprintf("%d. ", i+1)
		}
		g.pdf.SetX(15)
		g.pdf.Write(g.cfg.FontSize/2, prefix)
		for _, seg := range item.Content {
			g.pdf.Write(g.cfg.FontSize/2, seg.Text)
		}
		g.pdf.Ln(g.cfg.FontSize/2 + 2)
	}
	g.pdf.Ln(5)
}

// renderTable rendert eine Tabelle mit Kopfzeile und automatischer Spaltenbreite.
func (g *Generator) renderTable(t blocks.TableBlock) {
	if len(t.Rows) == 0 {
		return
	}

	left, _, right, _ := g.pdf.GetMargins()
	width := 210 - left - right
	colCount := len(t.Rows[0])
	if colCount == 0 {
		return
	}
	colWidth := width / float64(colCount)

	g.pdf.SetFont("Main", "B", g.cfg.FontSize)
	g.setPrimaryTextColor()

	for _, row := range t.Rows {
		maxH := 0.0
		for _, cell := range row {
			cellText := ""
			for _, seg := range cell.Content {
				cellText += seg.Text
			}
			h := float64(len(g.pdf.SplitLines([]byte(cellText), colWidth))) * (g.cfg.FontSize * 0.5)
			if h > maxH {
				maxH = h
			}
		}
		maxH += 4

		g.checkPageBreak(maxH)

		for i, cell := range row {
			x, y := g.pdf.GetX(), g.pdf.GetY()
			if i == 0 {
				g.pdf.SetX(left)
				x = left
			}

			if cell.Header {
				g.pdf.SetFillColor(240, 240, 240)
				g.pdf.Rect(x, y, colWidth, maxH, "F")
				g.pdf.SetFont("Main", "B", g.cfg.FontSize)
			} else {
				g.pdf.SetFont("Main", "", g.cfg.FontSize)
			}

			g.pdf.Rect(x, y, colWidth, maxH, "D")

			cellText := ""
			for _, seg := range cell.Content {
				cellText += seg.Text
			}

			g.pdf.MultiCell(colWidth, g.cfg.FontSize*0.5, cellText, "", "L", false)
			g.pdf.SetXY(x+colWidth, y)
		}
		g.pdf.Ln(maxH)
	}
	g.pdf.Ln(5)
}
