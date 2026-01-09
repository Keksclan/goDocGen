package pdf

import "fmt"

// renderTOC rendert das Inhaltsverzeichnis mit klickbaren Links und Seitenzahlen.
// Im Measurement-Modus gibt es die Anzahl der benötigten Seiten zurück.
func (g *Generator) renderTOC(isMeasurement bool) int {
	if !g.cfg.TOC.Enabled {
		return 0
	}

	startPage := g.pdf.PageNo()
	g.inTOC = true
	g.pdf.AddPage()
	g.inTOC = false

	if isMeasurement {
		if len(g.toc) == 0 {
			// Falls noch keine Einträge da sind (erster Lauf), reservieren wir eine Seite
			g.pdf.AddPage()
			return g.pdf.PageNo() - startPage
		}

		// Im ersten Durchgang messen wir, wie viele Seiten das TOC tatsächlich einnimmt
		g.pdf.SetY(40)
		g.pdf.Ln(15) // Titel
		g.pdf.Ln(10) // Linie

		for _, entry := range g.toc {
			h := 8.5
			if entry.Level == 1 {
				h = 10.5
			}
			g.checkPageBreak(h)
			g.pdf.Ln(h)
		}
		g.pdf.AddPage()
		return g.pdf.PageNo() - startPage
	}

	// Titel des Inhaltsverzeichnisses
	g.pdf.SetY(40)
	g.safeSetFont("main", "B", 24)
	r, green, b := hexToRGB(g.cfg.Colors.Title)
	g.pdf.SetTextColor(r, green, b)
	g.pdf.CellFormat(0, 15, "Inhaltsverzeichnis", "", 1, "L", false, 0, "")

	// Dekorative Linie
	left, _, right, _ := g.pdf.GetMargins()
	g.pdf.SetDrawColor(r, green, b)
	g.pdf.SetLineWidth(0.5)
	g.pdf.Line(left, g.pdf.GetY(), 210-right, g.pdf.GetY())
	g.pdf.Ln(10)

	g.setPrimaryTextColor()

	for _, entry := range g.toc {
		indent := float64((entry.Level - 1) * 8)
		g.pdf.SetX(left + indent)

		fontSize := 12.0
		h := 8.5
		style := ""
		if entry.Level == 1 {
			style = "B"
			g.safeSetFont("main", style, 12)
			g.pdf.Ln(1)
			h = 10.5
		} else {
			g.safeSetFont("main", "", 11)
			fontSize = 11.0
		}

		text := ""
		if g.cfg.TOC.ShowNumbers {
			text = entry.Number
		}
		text += entry.Text

		// Eintragstext als Link
		g.safeWriteLinkID(h, text, "main", style, entry.Link)

		// Punkte zwischen Text und Seitenzahl
		if g.cfg.TOC.ShowDots {
			g.safeSetFont("main", "", 10)
			g.pdf.SetTextColor(180, 180, 180)
			dotX := g.pdf.GetX() + 2
			dotEndX := 210 - right - 10
			remaining := dotEndX - dotX
			if remaining > 0 {
				dots := ""
				dotW := g.pdf.GetStringWidth(".")
				for i := 0; float64(i)*dotW < remaining; i++ {
					dots += "."
				}
				g.pdf.CellFormat(remaining, h, dots, "", 0, "L", false, 0, "")
			}
		}

		// Seitenzahl
		g.setPrimaryTextColor()
		g.safeSetFont("main", "B", fontSize)
		g.pdf.SetX(210 - right - 8)
		displayPage := entry.Page - g.cfg.PageNumbers.StartPage + 1
		if displayPage < 1 {
			displayPage = 1
		}
		g.pdf.CellFormat(8, h, fmt.Sprintf("%d", displayPage), "", 1, "R", false, entry.Link, "")
	}

	// Sicherstellen, dass der nächste Inhalt auf einer neuen Seite beginnt
	g.pdf.AddPage()
	return g.pdf.PageNo() - startPage
}

// measureTOC berechnet die Anzahl der Seiten, die das Inhaltsverzeichnis einnehmen wird,
// ohne das eigentliche PDF-Dokument zu verändern.
func (g *Generator) measureTOC() int {
	if !g.cfg.TOC.Enabled || len(g.toc) == 0 {
		return 0
	}

	_, top, _, bottom := g.pdf.GetMargins()
	pageHeight := 297.0
	usableHeight := pageHeight - top - bottom

	currentY := 40.0 // Start Y laut renderTOC
	currentY += 15.0 // Titel Höhe
	currentY += 10.0 // Linie und Abstand

	pages := 1
	for _, entry := range g.toc {
		h := 8.5
		if entry.Level == 1 {
			h = 10.5
			currentY += 1.0 // Extra-Abstand für Top-Level laut renderTOC
		}

		// checkPageBreak Logik simulieren
		if currentY+h > usableHeight+top-10 { // -10 als Sicherheitspuffer
			pages++
			currentY = top + 20.0 // Neue Seite
		}
		currentY += h
	}

	return pages
}
