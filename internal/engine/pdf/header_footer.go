package pdf

import (
	"fmt"
	"strings"
	"time"
)

// drawBackground zeichnet den Seitenhintergrund (Farbverlauf oder Vollfarbe).
func (g *Generator) drawBackground() {
	if g.cfg.Gradient.Enabled && g.cfg.Gradient.Global {
		g.drawGradient(g.cfg.Gradient.Start, g.cfg.Gradient.End, g.cfg.Gradient.Orientation)
	} else if g.cfg.Colors.Background != "" {
		r, green, b := hexToRGB(g.cfg.Colors.Background)
		g.pdf.SetFillColor(r, green, b)
		g.pdf.Rect(0, 0, 210, 297, "F")
	}
}

// setupHeaderFooter konfiguriert die Header- und Footer-Funktionen für das PDF.
func (g *Generator) setupHeaderFooter() {
	g.pdf.SetHeaderFunc(func() {
		g.drawBackground()
		if g.inTOC || g.pdf.PageNo() == 1 || g.pdf.PageNo() < g.cfg.PageNumbers.StartPage {
			return // Kein Header auf Titelseite, TOC oder vor Startseite
		}
		g.pdf.SetY(10)
		r, green, b := hexToRGB(g.cfg.Colors.Header)
		g.pdf.SetTextColor(r, green, b)
		g.safeSetFont("main", "", 8)

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
		if g.inTOC || g.pdf.PageNo() == 1 || g.pdf.PageNo() < g.cfg.PageNumbers.StartPage {
			return
		}
		g.pdf.SetY(-15)
		g.safeSetFont("main", "", 8)
		g.pdf.SetTextColor(128, 128, 128)

		left, _, right, _ := g.pdf.GetMargins()
		width := 210 - left - right

		if g.cfg.Footer.Image != "" {
			g.pdf.Image(g.cfg.Footer.Image, left, 282, 15, 0, false, "", 0, "")
		}

		// Zonen rendern
		if g.cfg.Footer.Left != "" {
			g.pdf.SetX(left)
			g.pdf.CellFormat(width, 10, g.replacePlaceholders(g.cfg.Footer.Left), "", 0, "L", false, 0, "")
		}
		if g.cfg.Footer.Center != "" {
			g.pdf.SetX(left)
			g.pdf.CellFormat(width, 10, g.replacePlaceholders(g.cfg.Footer.Center), "", 0, "C", false, 0, "")
		}
		if g.cfg.Footer.Right != "" {
			g.pdf.SetX(left)
			g.pdf.CellFormat(width, 10, g.replacePlaceholders(g.cfg.Footer.Right), "", 0, "R", false, 0, "")
		}
	})
}

// replacePlaceholders ersetzt Variablen wie {page}, {total}, {title} durch ihre aktuellen Werte.
func (g *Generator) replacePlaceholders(text string) string {
	displayPage := g.pdf.PageNo() - g.cfg.PageNumbers.StartPage + 1
	totalDisplayPages := g.totalPages - g.cfg.PageNumbers.StartPage + 1
	if totalDisplayPages < 1 {
		totalDisplayPages = 1
	}

	text = strings.ReplaceAll(text, "{page}", fmt.Sprintf("%d", displayPage))
	text = strings.ReplaceAll(text, "{total}", fmt.Sprintf("%d", totalDisplayPages))
	text = strings.ReplaceAll(text, "{title}", g.cfg.Title)
	text = strings.ReplaceAll(text, "{author}", g.cfg.Author)
	text = strings.ReplaceAll(text, "{date}", time.Now().Format("02.01.2006"))
	return text
}

// renderFrontPage rendert das Deckblatt des Dokuments.
func (g *Generator) renderFrontPage() {
	g.pdf.AddPage()

	r, green, b := hexToRGB(g.cfg.Colors.Title)

	if g.cfg.Gradient.Enabled {
		g.drawGradient(g.cfg.Gradient.Start, g.cfg.Gradient.End, g.cfg.Gradient.Orientation)
		g.pdf.SetTextColor(255, 255, 255)
		r, green, b = 255, 255, 255
	} else {
		if !g.cfg.Gradient.Enabled && g.cfg.Colors.Background == "" {
		} else if !g.cfg.Gradient.Enabled {
			g.pdf.SetFillColor(r, green, b)
			g.pdf.Rect(0, 0, 10, 297, "F")
		}
		g.pdf.SetTextColor(r, green, b)
	}

	g.pdf.SetY(60)
	g.pdf.SetX(30)
	g.safeSetFont("main", "B", 40)
	align := g.getAlign(g.cfg.Layout.StartPage)
	g.pdf.MultiCell(0, 15, g.cfg.Title, "", align, false)

	if g.cfg.Subtitle != "" {
		g.pdf.Ln(5)
		g.pdf.SetX(30)
		g.safeSetFont("main", "", 20)
		if !g.cfg.Gradient.Enabled {
			g.pdf.SetTextColor(100, 100, 100)
		}
		g.pdf.MultiCell(0, 12, g.cfg.Subtitle, "", align, false)
	}

	if g.cfg.Author != "" {
		g.pdf.Ln(5)
		g.pdf.SetX(30)
		g.safeSetFont("main", "I", 14)
		if !g.cfg.Gradient.Enabled {
			g.pdf.SetTextColor(120, 120, 120)
		}
		g.pdf.MultiCell(0, 10, fmt.Sprintf("Autor: %s", g.cfg.Author), "", align, false)
	}

	g.pdf.SetY(250)
	g.pdf.SetX(30)
	g.safeSetFont("main", "", 12)
	if !g.cfg.Gradient.Enabled {
		g.pdf.SetTextColor(128, 128, 128)
	}
	if g.cfg.Author != "" {
		g.pdf.MultiCell(0, 10, fmt.Sprintf("Erstellt von: %s", g.cfg.Author), "", align, false)
	}
	g.pdf.MultiCell(0, 10, fmt.Sprintf("Datum: %s", time.Now().Format("02.01.2006")), "", align, false)
}
