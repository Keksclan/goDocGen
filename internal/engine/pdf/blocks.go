package pdf

import (
	"godocgen/internal/blocks"
	"fmt"
)

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
		// Will be handled during builder phase (rendered to PNG)
	case blocks.ListBlock:
		g.renderList(b)
	case blocks.TableBlock:
		g.renderTable(b)
	case blocks.PageBreakBlock:
		g.pdf.AddPage()
	}
}

func (g *Generator) renderHeading(h blocks.HeadingBlock, isMeasurement bool) {
	// Update heading counts for numbering
	if h.Level > 0 && h.Level <= len(g.headingCounts) {
		g.headingCounts[h.Level-1]++
		// Reset sub-levels
		for i := h.Level; i < len(g.headingCounts); i++ {
			g.headingCounts[i] = 0
		}
	}

	// Generate numbering string (e.g., "1.2.3 ")
	numbering := ""
	for i := 0; i < h.Level; i++ {
		numbering += fmt.Sprintf("%d.", g.headingCounts[i])
	}
	if numbering != "" {
		numbering += " "
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

	size := 18.0
	if h.Level == 1 {
		size = 22.0
		g.pdf.Ln(10)
	} else if h.Level == 2 {
		size = 18.0
		g.pdf.Ln(5)
	} else {
		size = 14.0
		g.pdf.Ln(3)
	}

	g.pdf.SetFont("Main", "B", size)
	g.checkPageBreak(size + 15)
	r, green, b := hexToRGB(g.cfg.Colors.Title)
	g.pdf.SetTextColor(r, green, b)

	// Modern accent: small line before heading
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
	g.pdf.MultiCell(0, 10, displayText, "", "L", false)
	g.pdf.Ln(3)
}

func (g *Generator) renderParagraph(p blocks.ParagraphBlock) {
	g.pdf.SetFont("Main", "", g.cfg.FontSize)
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
			g.pdf.SetFont(fontFamily, "I", g.cfg.FontSize)
			g.pdf.SetFillColor(240, 240, 240)
			if seg.Text != "" {
				g.pdf.Write(g.cfg.FontSize/2, seg.Text)
			}
		} else {
			g.pdf.SetFont("Main", style, g.cfg.FontSize)
			if seg.Text != "" {
				g.pdf.Write(g.cfg.FontSize/2, seg.Text)
			}
		}
	}
	g.pdf.Ln(g.cfg.FontSize/2 + 5)
}

func (g *Generator) renderCode(c blocks.CodeBlock) {
	fontFamily := "Main"
	if g.cfg.Fonts.Mono != "" {
		fontFamily = "Mono"
	}

	// Language header
	if c.Language != "" {
		g.pdf.SetFont("Main", "B", 7)
		g.pdf.SetTextColor(150, 150, 150)
		g.pdf.CellFormat(0, 4, c.Language, "", 1, "R", false, 0, "")
	}

	g.pdf.SetFont(fontFamily, "I", g.cfg.FontSize)
	bgR, bgG, bgB := 245, 245, 245 // Light grey background for container
	if c.BgColor != "" {
		r, green, b := hexToRGB(c.BgColor)
		// If background is very light (white), keep our light grey for container visibility
		if r < 250 || green < 250 || b < 250 {
			bgR, bgG, bgB = r, green, b
		}
	}
	g.pdf.SetFillColor(bgR, bgG, bgB)
	g.pdf.SetDrawColor(200, 200, 200) // Light grey border

	// Draw a box for code
	x := g.pdf.GetX()
	y := g.pdf.GetY()

	// Simple height estimation
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

	// Intelligent page break for code block
	g.checkPageBreak(rectHeight + 10)

	x = g.pdf.GetX()
	y = g.pdf.GetY()
	left, _, right, _ := g.pdf.GetMargins()
	width := 210 - left - right

	g.pdf.RoundedRect(x, y, width, rectHeight, 4, "1234", "DF") // Rounded corners
	g.pdf.SetX(x + 5)                                           // Padding
	g.pdf.SetY(y + 5)                                           // Padding

	for _, seg := range c.Segments {
		if seg.Color != "" {
			r, green, b := hexToRGB(seg.Color)
			g.pdf.SetTextColor(r, green, b)
		} else {
			g.setPrimaryTextColor()
		}

		// Handle newlines to keep X position
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
					g.pdf.Write(lineHeight, text)
				}
				break
			}

			if idx > 0 {
				g.pdf.Write(lineHeight, text[:idx])
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

func (g *Generator) renderImage(i blocks.ImageBlock) {
	// Register image to get info
	g.pdf.RegisterImage(i.Path, "")
	info := g.pdf.GetImageInfo(i.Path)
	left, top, right, bottom := g.pdf.GetMargins()
	maxWidth := 210 - left - right
	maxPageHeight := 297 - top - bottom - 40 // Adjusted for header/footer

	widthOnPage := maxWidth - 20
	var h float64
	if info != nil && info.Width() > 0 {
		h = (info.Height() / info.Width()) * widthOnPage
	} else {
		h = 60.0
	}

	// Add title height if exists
	titleHeight := 0.0
	if i.Title != "" {
		titleHeight = 10.0
	}

	// Intelligent scaling and page break
	if h+titleHeight > maxPageHeight {
		// Scale down to fit one whole page
		h = maxPageHeight - titleHeight
		widthOnPage = 0 // Auto-width based on h
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

	// Render Title
	if i.Title != "" {
		g.pdf.SetFont("Main", "B", 10)
		g.pdf.SetTextColor(100, 100, 100)
		g.pdf.CellFormat(0, 8, i.Title, "", 1, "C", false, 0, "")
		g.pdf.Ln(2)
	}

	// Center image container
	imgX := x + (maxWidth-containerW)/2

	// Background for diagram/image container
	g.pdf.SetFillColor(255, 255, 255)
	if g.cfg.Colors.Background != "" {
		// If page has background, maybe make container slightly different or keep white
		// Let's use a very light grey if white is the background
		g.pdf.SetFillColor(250, 250, 250)
	}
	g.pdf.SetDrawColor(220, 220, 220)
	g.pdf.RoundedRect(imgX, g.pdf.GetY(), containerW, containerH, 5, "1234", "DF")

	g.pdf.Image(i.Path, imgX+padding, g.pdf.GetY()+padding, imgW, h, false, "", 0, "")

	g.pdf.SetY(y + containerH + titleHeight + 5)
	g.pdf.Ln(2)
}

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
			// Similar to paragraph
			g.pdf.Write(g.cfg.FontSize/2, seg.Text)
		}
		g.pdf.Ln(g.cfg.FontSize/2 + 2)
	}
	g.pdf.Ln(5)
}

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
		// Calculate max height for this row
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
		maxH += 4 // Padding

		g.checkPageBreak(maxH)

		for i, cell := range row {
			x, y := g.pdf.GetX(), g.pdf.GetY()
			if i == 0 {
				g.pdf.SetX(left)
				x = left
			}

			// Background for header
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

