package pdf

import (
	"docgen/internal/blocks"
	"docgen/internal/config"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

type Generator struct {
	pdf           *gofpdf.Fpdf
	cfg           *config.Config
	blocks        []blocks.DocBlock
	toc           []TOCEntry
	fontDir       string
	totalPages    int
	headingCounts []int
}

type TOCEntry struct {
	Level  int
	Number string
	Text   string
	Page   int
	Link   int
}

func NewGenerator(cfg *config.Config, blocks []blocks.DocBlock, fontDir string) *Generator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(cfg.Layout.Margins.Left, cfg.Layout.Margins.Top, cfg.Layout.Margins.Right)
	pdf.SetAutoPageBreak(true, cfg.Layout.Margins.Bottom)

	// Register Fonts
	registerFonts(pdf, cfg, fontDir)

	return &Generator{
		pdf:           pdf,
		cfg:           cfg,
		blocks:        blocks,
		fontDir:       fontDir,
		headingCounts: make([]int, 6),
	}
}

func registerFonts(pdf *gofpdf.Fpdf, cfg *config.Config, fontDir string) {
	regularPath := filepath.Join(fontDir, cfg.Fonts.Regular)
	if _, err := os.Stat(regularPath); err == nil {
		pdf.AddUTF8Font("Main", "", regularPath)
	}

	if cfg.Fonts.Bold != "" {
		boldPath := filepath.Join(fontDir, cfg.Fonts.Bold)
		if _, err := os.Stat(boldPath); err == nil {
			pdf.AddUTF8Font("Main", "B", boldPath)
		}
	}
	if cfg.Fonts.Italic != "" {
		italicPath := filepath.Join(fontDir, cfg.Fonts.Italic)
		if _, err := os.Stat(italicPath); err == nil {
			pdf.AddUTF8Font("Main", "I", italicPath)
		}
	}
	if cfg.Fonts.Mono != "" {
		monoPath := filepath.Join(fontDir, cfg.Fonts.Mono)
		if _, err := os.Stat(monoPath); err == nil {
			pdf.AddUTF8Font("Mono", "", monoPath)
			pdf.AddUTF8Font("Mono", "I", monoPath)
			pdf.AddUTF8Font("Mono", "B", monoPath)
			pdf.AddUTF8Font("Mono", "BI", monoPath)
		}
	}
}

func (g *Generator) Generate(outputPath string) error {
	// Pass 1: Measure and collect TOC
	g.headingCounts = make([]int, 6)
	g.renderAll(true)
	g.totalPages = g.pdf.PageNo()

	// Reset for Pass 2
	g.pdf = gofpdf.New("P", "mm", "A4", "")
	g.pdf.SetMargins(g.cfg.Layout.Margins.Left, g.cfg.Layout.Margins.Top, g.cfg.Layout.Margins.Right)
	g.pdf.SetAutoPageBreak(true, g.cfg.Layout.Margins.Bottom)
	registerFonts(g.pdf, g.cfg, g.fontDir)
	g.headingCounts = make([]int, 6)

	// Pass 2: Final render
	g.renderAll(false)

	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return err
	}

	return g.pdf.OutputFileAndClose(outputPath)
}

func (g *Generator) renderAll(isMeasurement bool) {
	g.setupHeaderFooter()

	// Front Page
	g.renderFrontPage()

	// TOC Page
	if !isMeasurement {
		g.renderTOC()
	} else {
		// Just a placeholder to count pages
		g.pdf.AddPage()
	}

	// Content
	for _, block := range g.blocks {
		g.renderBlock(block, isMeasurement)
	}
}
