package pdf

import "fmt"

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

func (g *Generator) renderTOC() {
	g.pdf.AddPage()

	// Title
	g.pdf.SetY(40)
	g.safeSetFont("Main", "B", 24)
	r, green, b := hexToRGB(g.cfg.Colors.Title)
	g.pdf.SetTextColor(r, green, b)
	g.pdf.CellFormat(0, 15, "Inhaltsverzeichnis", "", 1, "L", false, 0, "")

	// Decorative line
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
			g.safeSetFont("Main", style, 12)
			g.pdf.Ln(2)
		} else {
			g.safeSetFont("Main", "", 11)
			fontSize = 11.0
		}

		text := entry.Number + entry.Text

		// Entry Text
		g.safeWriteLinkID(fontSize, text, "Main", style, entry.Link)

		// Dots
		g.safeSetFont("Main", "", 10)
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

		// Page Number
		g.setPrimaryTextColor()
		g.safeSetFont("Main", "B", fontSize)
		g.pdf.SetX(210 - right - 8)
		g.pdf.CellFormat(8, fontSize, fmt.Sprintf("%d", entry.Page), "", 1, "R", false, entry.Link, "")
	}
}
