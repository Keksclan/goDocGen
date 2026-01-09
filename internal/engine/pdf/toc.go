package pdf

import "fmt"

// getEntryHeight berechnet die Höhe eines Inhaltsverzeichniseintrags basierend auf seiner Ebene.
func getEntryHeight(level int, scale float64) float64 {
	fontSize := 12.0 * scale
	extra := 0.0
	if level == 1 {
		fontSize = 13.0 * scale
		extra = 2.0 * scale
	} else if level == 2 {
		fontSize = 11.0 * scale
		extra = 0.0
	} else {
		fontSize = 10.0 * scale
		extra = 0.0
	}
	return fontSize*0.8 + 2*scale + extra
}

// renderTOC rendert das Inhaltsverzeichnis mit klickbaren Links und Seitenzahlen.
func (g *Generator) renderTOC(isMeasurement bool) {
	if !g.cfg.TOC.Enabled {
		return
	}

	g.inTOC = true
	g.pdf.AddPage()
	g.inTOC = false

	if isMeasurement {
		// Im ersten Durchgang messen wir nur, wie viele Seiten das TOC einnimmt
		// Wir simulieren das Rendern der Einträge
		g.pdf.SetY(40)
		g.pdf.Ln(15) // Titel
		g.pdf.Ln(10) // Linie

		for _, entry := range g.toc {
			// Wir nutzen eine vereinfachte Höhenberechnung
			h := 10.0
			if entry.Level == 1 {
				h = 12.0
			}
			g.checkPageBreak(h)
			g.pdf.Ln(h)
		}
		g.pdf.AddPage()
		return
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
		style := ""
		if entry.Level == 1 {
			style = "B"
			g.safeSetFont("main", style, 12)
			g.pdf.Ln(2)
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
		g.safeWriteLinkID(fontSize, text, "main", style, entry.Link)

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
				g.pdf.CellFormat(remaining, fontSize, dots, "", 0, "L", false, 0, "")
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
		g.pdf.CellFormat(8, fontSize, fmt.Sprintf("%d", displayPage), "", 1, "R", false, entry.Link, "")
	}

	// Sicherstellen, dass der nächste Inhalt auf einer neuen Seite beginnt
	g.pdf.AddPage()
}
