package engine

import (
	"docgen/internal/blocks"
	"docgen/internal/config"
	"docgen/internal/engine/code"
	"docgen/internal/engine/fonts"
	"docgen/internal/engine/markdown"
	"docgen/internal/engine/mermaid"
	"docgen/internal/engine/pdf"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Builder struct {
	ProjectDir string
	OutDir     string
	CacheDir   string
	ConfigName string
}

func NewBuilder(projectDir, outDir string) *Builder {
	return &Builder{
		ProjectDir: projectDir,
		OutDir:     outDir,
		CacheDir:   ".cache",
		ConfigName: "docgen.yml",
	}
}

func (b *Builder) Build() error {
	// 1. Load Config
	cfgPath := filepath.Join(b.ProjectDir, b.ConfigName)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// 2. Extract Fonts
	fontZip := filepath.Join(b.ProjectDir, cfg.Fonts.Zip)
	fontDir, err := fonts.ExtractFonts(fontZip, b.CacheDir)
	if err != nil {
		return fmt.Errorf("font error: %w", err)
	}

	// 3. Load Markdown files (recursive)
	contentDir := filepath.Join(b.ProjectDir, "content")
	var mdFiles []string
	err = filepath.WalkDir(contentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not read content dir: %w", err)
	}

	sort.Strings(mdFiles)

	var allBlocks []blocks.DocBlock
	for _, path := range mdFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		blocks, err := markdown.Parse(data)
		if err != nil {
			return err
		}
		allBlocks = append(allBlocks, blocks...)
	}

	// 4. Preprocess Blocks (Mermaid & Code)
	for i, block := range allBlocks {
		switch blk := block.(type) {
		case blocks.MermaidBlock:
			svgPath, pngPath, err := mermaid.Render(blk.Content, b.CacheDir)
			if err != nil {
				fmt.Printf("Warning: Could not render Mermaid diagram: %v\n", err)
				allBlocks[i] = blocks.ParagraphBlock{
					Content: []blocks.TextSegment{
						{Text: "[Diagramm konnte nicht gerendert werden - mmdc fehlt]", Italic: true},
					},
				}
			} else {
				// We use PNG for PDF because gofpdf has better support for it, 
				// but we generated SVG as requested.
				allBlocks[i] = blocks.ImageBlock{
					Path:  pngPath,
					Alt:   "Mermaid Diagram (SVG Source: " + svgPath + ")",
					Title: blk.Title,
				}
			}
		case blocks.CodeBlock:
			segments, bg, err := code.GetSegments(blk.Content, blk.Language, cfg.CodeTheme)
			if err != nil {
				return err
			}
			blk.Segments = segments
			blk.BgColor = bg
			allBlocks[i] = blk
		case blocks.ImageBlock:
			// Resolve relative paths
			if !filepath.IsAbs(blk.Path) {
				blk.Path = filepath.Join(b.ProjectDir, "assets", blk.Path)
				allBlocks[i] = blk
			}
		}
	}

	// 5. Generate PDF with Versioning
	baseName := cfg.Title
	if baseName == "" {
		baseName = "Documentation"
	}

	outputPath := ""
	version := 1
	for {
		fileName := fmt.Sprintf("%s_v%d.pdf", baseName, version)
		outputPath = filepath.Join(b.OutDir, fileName)
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			break
		}
		version++
	}

	gen := pdf.NewGenerator(cfg, allBlocks, fontDir)
	return gen.Generate(outputPath)
}
