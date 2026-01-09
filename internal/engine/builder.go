// Package engine enthält die Kernlogik zur Verarbeitung von Markdown und Erzeugung von PDFs.
package engine

import (
	"godocgen/internal/blocks"
	"godocgen/internal/config"
	"godocgen/internal/engine/code"
	"godocgen/internal/engine/fonts"
	"godocgen/internal/engine/markdown"
	"godocgen/internal/engine/mermaid"
	"godocgen/internal/engine/pdf"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Builder koordiniert den gesamten Build-Prozess eines Dokumentationsprojekts.
type Builder struct {
	ProjectDir string // Wurzelverzeichnis des Projekts
	OutDir     string // Verzeichnis, in dem das PDF gespeichert wird
	CacheDir   string // Verzeichnis für temporäre Dateien (Fonts, Diagramme)
	ConfigName string // Name der Konfigurationsdatei (Standard: docgen.yml)
}

// NewBuilder erstellt eine neue Builder-Instanz mit Standardwerten.
func NewBuilder(projectDir, outDir string) *Builder {
	return &Builder{
		ProjectDir: projectDir,
		OutDir:     outDir,
		CacheDir:   ".cache",
		ConfigName: "docgen.yml",
	}
}

// Build führt den Build-Prozess aus und gibt den Pfad zur generierten PDF-Datei zurück.
func (b *Builder) Build() (string, error) {
	// 1. Konfiguration laden
	cfgPath := filepath.Join(b.ProjectDir, b.ConfigName)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return "", fmt.Errorf("Konfigurationsfehler: %w", err)
	}

	// 2. Schriften extrahieren/herunterladen
	var fontDir string
	if cfg.Fonts.Zip != "" || cfg.Fonts.URL != "" {
		var fontZip string
		if cfg.Fonts.URL != "" {
			var err error
			fontZip, err = fonts.DownloadFonts(cfg.Fonts.URL, b.CacheDir)
			if err != nil {
				return "", fmt.Errorf("Schriften konnten nicht heruntergeladen werden: %w", err)
			}
		} else {
			fontZip = filepath.Join(b.ProjectDir, cfg.Fonts.Zip)
		}

		var err error
		fontDir, err = fonts.ExtractFonts(fontZip, b.CacheDir)
		if err != nil {
			return "", fmt.Errorf("Schriftartenfehler: %w", err)
		}
	} else {
		// System-Fonts Modus: fontDir bleibt leer, Pfade werden absolut oder relativ zum Projekt aufgelöst
		fontDir = b.ProjectDir
	}

	// 3. Markdown-Dateien rekursiv laden
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
		return "", fmt.Errorf("Content-Verzeichnis konnte nicht gelesen werden: %w", err)
	}

	sort.Slice(mdFiles, func(i, j int) bool {
		// Sortiere primär nach Dateiname, ignoriere Ordnerstruktur für die Reihenfolge
		// falls Dateinamen gleich sind, nimm den vollen Pfad als Fallback
		baseI := filepath.Base(mdFiles[i])
		baseJ := filepath.Base(mdFiles[j])
		if baseI != baseJ {
			return baseI < baseJ
		}
		return mdFiles[i] < mdFiles[j]
	})

	var allBlocks []blocks.DocBlock
	for _, path := range mdFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		blocks, err := markdown.Parse(data)
		if err != nil {
			return "", err
		}
		allBlocks = append(allBlocks, blocks...)
	}

	// 4. Blöcke vorverarbeiten (Mermaid & Code-Highlighting)
	for i, block := range allBlocks {
		switch blk := block.(type) {
		case blocks.MermaidBlock:
			svgPath, pngPath, err := mermaid.Render(blk.Content, b.CacheDir)
			if err != nil {
				fmt.Printf("Warnung: Mermaid-Diagramm konnte nicht gerendert werden: %v\n", err)
				allBlocks[i] = blocks.ParagraphBlock{
					Content: []blocks.TextSegment{
						{Text: "[Diagramm konnte nicht gerendert werden - mmdc fehlt]", Italic: true},
					},
				}
			} else {
				allBlocks[i] = blocks.ImageBlock{
					Path:  pngPath,
					Alt:   "Mermaid Diagram (SVG Quelle: " + svgPath + ")",
					Title: blk.Title,
				}
			}
		case blocks.CodeBlock:
			segments, bg, err := code.GetSegments(blk.Content, blk.Language, cfg.CodeTheme)
			if err != nil {
				return "", err
			}
			blk.Segments = segments
			blk.BgColor = bg
			allBlocks[i] = blk
		case blocks.ImageBlock:
			// Relative Pfade auflösen
			if !filepath.IsAbs(blk.Path) {
				blk.Path = filepath.Join(b.ProjectDir, "assets", blk.Path)
				allBlocks[i] = blk
			}
		}
	}

	// 5. PDF mit Versionierung generieren
	baseName := cfg.Title
	if baseName == "" {
		baseName = "Dokumentation"
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

	// Falls das Ausgabeverzeichnis nicht existiert, erstellen
	if err := os.MkdirAll(b.OutDir, 0755); err != nil {
		return "", fmt.Errorf("Ausgabeverzeichnis konnte nicht erstellt werden: %w", err)
	}

	gen := pdf.NewGenerator(cfg, allBlocks, fontDir)
	return outputPath, gen.Generate(outputPath)
}
