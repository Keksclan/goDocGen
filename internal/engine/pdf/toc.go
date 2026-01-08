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
	// Disable auto page break for TOC to ensure it stays on one page
	oldAuto, oldMargin := g.pdf.GetAutoPageBreak()
	g.pdf.SetAutoPageBreak(false, 0)
	defer g.pdf.SetAutoPageBreak(oldAuto, oldMargin)

	// Title
	g.pdf.SetY(40)
	g.pdf.SetFont("Main", "B", 28)
	r, green, b := hexToRGB(g.cfg.Colors.Title)
	g.pdf.SetTextColor(r, green, b)
	g.pdf.CellFormat(0, 20, "Inhaltsverzeichnis", "", 1, "L", false, 0, "")

	// Decorative line
	left, _, right, bottom := g.pdf.GetMargins()
	g.pdf.SetDrawColor(r, green, b)
	g.pdf.SetLineWidth(1)
	g.pdf.Line(left, g.pdf.GetY(), 210-right, g.pdf.GetY())
	g.pdf.Ln(15)

	startY := g.pdf.GetY()
	availHeight := (297 - bottom) - startY

	totalHeightNormal := 0.0
	for _, entry := range g.toc {
		totalHeightNormal += getEntryHeight(entry.Level, 1.0)
	}

	scaleFactor := 1.0
	useColumns := false

	if totalHeightNormal > availHeight {
		useColumns = true
		if totalHeightNormal > 2*availHeight {
			scaleFactor = (2 * availHeight) / totalHeightNormal
			if scaleFactor < 0.5 {
				scaleFactor = 0.5 // Minimum scale to keep it readable
			}
		}
	}

	fullWidth := 210 - left - right
	columnWidth := fullWidth
	if useColumns {
		columnWidth = (fullWidth - 4) / 2
	}

	colX := []float64{left, left + columnWidth + 4}
	currentCol := 0
	currentY := startY
	g.setPrimaryTextColor()

	totalScaledHeight := totalHeightNormal * scaleFactor

	for i, entry := range g.toc {
		// Decide if we need to switch column
		if useColumns && currentCol == 0 {
			heightSoFar := 0.0
			for j := 0; j <= i; j++ {
				heightSoFar += getEntryHeight(g.toc[j].Level, scaleFactor)
			}
			if heightSoFar > totalScaledHeight/2+2 {
				currentCol = 1
				currentY = startY
			}
		}

		indent := float64((entry.Level-1)*3) * scaleFactor
		if !useColumns {
			indent = float64((entry.Level-1)*6) * scaleFactor
		}

		g.pdf.SetX(colX[currentCol] + indent)
		g.pdf.SetY(currentY)

		fontSize := 12.0 * scaleFactor
		if entry.Level == 1 {
			g.pdf.SetFont("Main", "B", 13*scaleFactor)
			g.pdf.Ln(2 * scaleFactor)
			currentY += 2 * scaleFactor
		} else if entry.Level == 2 {
			g.pdf.SetFont("Main", "", 11*scaleFactor)
			fontSize = 11.0 * scaleFactor
		} else {
			g.pdf.SetFont("Main", "I", 10*scaleFactor)
			fontSize = 10.0 * scaleFactor
		}

		text := entry.Number + entry.Text
		maxWidth := columnWidth - indent - 8
		for g.pdf.GetStringWidth(text) > maxWidth && len(text) > 5 {
			text = text[:len(text)-4] + "..."
		}

		tw := g.pdf.GetStringWidth(text)

		// Entry Text
		g.pdf.SetX(colX[currentCol] + indent)
		g.pdf.WriteLinkID(fontSize, text, entry.Link)

		// Dots
		g.pdf.SetFont("Main", "", 10*scaleFactor)
		g.pdf.SetTextColor(150, 150, 150)
		dotX := colX[currentCol] + indent + tw + 2
		dotEndX := colX[currentCol] + columnWidth - 7
		remaining := dotEndX - dotX
		if remaining > 2 {
			g.pdf.SetX(dotX)
			dots := ""
			dotW := g.pdf.GetStringWidth(".")
			for i := 0; float64(i)*dotW < remaining; i++ {
				dots += "."
			}
			g.pdf.CellFormat(remaining, fontSize, dots, "", 0, "L", false, 0, "")
		}

		// Page Number
		g.setPrimaryTextColor()
		g.pdf.SetFont("Main", "B", fontSize)
		g.pdf.SetX(colX[currentCol] + columnWidth - 6)
		g.pdf.CellFormat(6, fontSize, fmt.Sprintf("%d", entry.Page), "", 0, "R", false, entry.Link, "")

		lineHeight := fontSize*0.8 + 2*scaleFactor
		currentY += lineHeight
	}
}
