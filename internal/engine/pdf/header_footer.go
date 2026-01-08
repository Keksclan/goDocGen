package pdf

import (
	"fmt"
	"time"
)

func (g *Generator) drawBackground() {
	if g.cfg.Colors.Background != "" {
		r, green, b := hexToRGB(g.cfg.Colors.Background)
		g.pdf.SetFillColor(r, green, b)
		g.pdf.Rect(0, 0, 210, 297, "F")
	}
}

func (g *Generator) setupHeaderFooter() {
	g.pdf.SetHeaderFunc(func() {
		g.drawBackground()
		if g.pdf.PageNo() == 1 {
			return // No header on front page
		}
		g.pdf.SetY(10)
		r, green, b := hexToRGB(g.cfg.Colors.Header)
		g.pdf.SetTextColor(r, green, b)
		g.pdf.SetFont("Main", "", 8)

		if g.cfg.Header.Image != "" {
			g.pdf.Image(g.cfg.Header.Image, 10, 10, 20, 0, false, "", 0, "")
			g.pdf.SetX(35)
		} else {
			g.pdf.SetX(10)
		}

		g.pdf.CellFormat(0, 10, g.cfg.Header.Text, "", 0, "L", false, 0, "")
		g.pdf.Ln(15)
	})

	g.pdf.SetFooterFunc(func() {
		if g.pdf.PageNo() < g.cfg.PageNumbers.StartPage {
			return
		}
		g.pdf.SetY(-20)
		g.pdf.SetFont("Main", "", 8)
		g.pdf.SetTextColor(128, 128, 128)

		if g.cfg.Footer.Image != "" {
			g.pdf.Image(g.cfg.Footer.Image, 10, 277, 20, 0, false, "", 0, "")
			g.pdf.SetX(35)
		} else {
			g.pdf.SetX(10)
		}

		g.pdf.CellFormat(0, 10, g.cfg.Footer.Text, "", 0, "L", false, 0, "")

		displayPage := g.pdf.PageNo() - g.cfg.PageNumbers.StartPage + 1
		totalDisplayPages := g.totalPages - g.cfg.PageNumbers.StartPage + 1

		if totalDisplayPages < 1 {
			totalDisplayPages = 1 // Fallback
		}

		g.pdf.SetX(-40)
		g.pdf.CellFormat(0, 10, fmt.Sprintf("%d / %d", displayPage, totalDisplayPages), "", 0, "R", false, 0, "")
	})
}

func (g *Generator) renderFrontPage() {
	g.pdf.AddPage()

	r, green, b := hexToRGB(g.cfg.Colors.Title)

	// Decorative background element
	g.pdf.SetFillColor(r, green, b)
	g.pdf.Rect(0, 0, 10, 297, "F") // Side bar

	g.pdf.SetY(60)
	g.pdf.SetX(30)
	g.pdf.SetFont("Main", "B", 40)
	g.pdf.SetTextColor(r, green, b)
	g.pdf.MultiCell(0, 15, g.cfg.Title, "", "L", false)

	if g.cfg.Subtitle != "" {
		g.pdf.Ln(10)
		g.pdf.SetX(30)
		g.pdf.SetFont("Main", "", 20)
		g.pdf.SetTextColor(100, 100, 100)
		g.pdf.MultiCell(0, 12, g.cfg.Subtitle, "", "L", false)
	}

	// Bottom info
	g.pdf.SetY(250)
	g.pdf.SetX(30)
	g.pdf.SetFont("Main", "", 12)
	g.pdf.SetTextColor(128, 128, 128)
	g.pdf.CellFormat(0, 10, fmt.Sprintf("Datum: %s", time.Now().Format("02.01.2006")), "", 1, "L", false, 0, "")
}
