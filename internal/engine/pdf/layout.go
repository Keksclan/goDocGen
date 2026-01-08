package pdf

import (
	"strconv"
)

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

func (g *Generator) setTextColor(hex string) {
	r, green, b := hexToRGB(hex)
	g.pdf.SetTextColor(r, green, b)
}

func (g *Generator) setPrimaryTextColor() {
	if g.cfg.Colors.Text != "" {
		g.setTextColor(g.cfg.Colors.Text)
	} else {
		g.pdf.SetTextColor(0, 0, 0)
	}
}

func (g *Generator) setFillColor(hex string) {
	r, green, b := hexToRGB(hex)
	g.pdf.SetFillColor(r, green, b)
}

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

func (g *Generator) checkPageBreak(h float64) {
	_, _, _, bottom := g.pdf.GetMargins()
	if g.pdf.GetY()+h > 297-bottom {
		g.pdf.AddPage()
	}
}
