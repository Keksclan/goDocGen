package pdf

import (
	"strconv"
)

// hexToRGB wandelt einen Hexadezimal-Farbcode in RGB-Werte um.
func hexToRGB(hex string) (int, int, int) {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return int(r), int(g), int(b)
}

// setTextColor setzt die Textfarbe basierend auf einem Hex-Code.
func (g *Generator) setTextColor(hex string) {
	r, green, b := hexToRGB(hex)
	g.pdf.SetTextColor(r, green, b)
}

// setPrimaryTextColor setzt die Standard-Textfarbe aus der Konfiguration.
func (g *Generator) setPrimaryTextColor() {
	if g.cfg.Colors.Text != "" {
		g.setTextColor(g.cfg.Colors.Text)
	} else {
		g.pdf.SetTextColor(0, 0, 0)
	}
}

// setFillColor setzt die Füllfarbe basierend auf einem Hex-Code.
func (g *Generator) setFillColor(hex string) {
	r, green, b := hexToRGB(hex)
	g.pdf.SetFillColor(r, green, b)
}

// getAlign wandelt lesbare Ausrichtungsnamen in gofpdf-Konstanten um.
func (g *Generator) getAlign(align string) string {
	switch align {
	case "center":
		return "C"
	case "right":
		return "R"
	case "justify":
		return "J"
	default:
		return "L"
	}
}

// checkPageBreak prüft, ob die verbleibende Höhe auf der Seite ausreicht, und fügt ggf. eine neue Seite hinzu.
func (g *Generator) checkPageBreak(h float64) {
	_, _, _, bottom := g.pdf.GetMargins()
	if g.pdf.GetY()+h > 297-bottom {
		g.pdf.AddPage()
	}
}

// drawGradient zeichnet einen linearen Farbverlauf auf der aktuellen Seite.
func (g *Generator) drawGradient(startColor, endColor string, orientation string) {
	sr, sg, sb := hexToRGB(startColor)
	er, eg, eb := hexToRGB(endColor)

	steps := 100
	if orientation == "horizontal" {
		w := 210.0 / float64(steps)
		for i := 0; i < steps; i++ {
			ratio := float64(i) / float64(steps)
			currR := int(float64(sr) + ratio*float64(er-sr))
			currG := int(float64(sg) + ratio*float64(eg-sg))
			currB := int(float64(sb) + ratio*float64(eb-sb))
			g.pdf.SetFillColor(currR, currG, currB)
			g.pdf.Rect(float64(i)*w, 0, w+0.1, 297, "F")
		}
	} else {
		// Vertikal (Standard)
		h := 297.0 / float64(steps)
		for i := 0; i < steps; i++ {
			ratio := float64(i) / float64(steps)
			currR := int(float64(sr) + ratio*float64(er-sr))
			currG := int(float64(sg) + ratio*float64(eg-sg))
			currB := int(float64(sb) + ratio*float64(eb-sb))
			g.pdf.SetFillColor(currR, currG, currB)
			g.pdf.Rect(0, float64(i)*h, 210, h+0.1, "F")
		}
	}
}
