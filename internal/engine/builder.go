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
	"regexp"
	"sort"
	"strings"
)

type numberedFile struct {
	path      string
	numbering string
	sortKey   string
}

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

	// 3. Markdown-Dateien rekursiv laden und nach Header-Nummern sortieren
	contentDir := filepath.Join(b.ProjectDir, "content")
	numberedFiles, err := b.scanAndSortContent(contentDir)
	if err != nil {
		return "", fmt.Errorf("Content-Verzeichnis konnte nicht gelesen werden: %w", err)
	}

	var allBlocks []blocks.DocBlock
	for _, nf := range numberedFiles {
		data, err := os.ReadFile(nf.path)
		if err != nil {
			return "", err
		}
		blks, err := markdown.Parse(data, nf.numbering)
		if err != nil {
			return "", err
		}
		allBlocks = append(allBlocks, blks...)
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
				// Mermaid-Konfiguration für Größe anwenden
				allBlocks[i] = blocks.ImageBlock{
					Path:  pngPath,
					Alt:   "Mermaid Diagram (SVG Quelle: " + svgPath + ")",
					Title: blk.Title,
					Width: cfg.Mermaid.Width, // Konfigurierbare Breite
					Scale: cfg.Mermaid.Scale, // Konfigurierbare Skalierung
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

// scanAndSortContent durchläuft das Verzeichnis, extrahiert Header-Nummern und sortiert danach.
func (b *Builder) scanAndSortContent(dir string) ([]numberedFile, error) {
	var files []numberedFile

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		sortKey, numbering := b.extractHeaderInfo(path)
		files = append(files, numberedFile{
			path:      path,
			sortKey:   sortKey,
			numbering: numbering,
		})
		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Sortieren basierend auf dem sortKey (numerisch)
	sort.Slice(files, func(i, j int) bool {
		return compareVersions(files[i].sortKey, files[j].sortKey)
	})

	return files, nil
}

// extractHeaderInfo liest die erste Header-Zeile und extrahiert die Nummerierung.
func (b *Builder) extractHeaderInfo(path string) (string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "999", ""
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Suche nach dem ersten Header (# Text oder ## Text)
	re := regexp.MustCompile(`^#+\s+(([0-9]+\.)+[0-9]*|[0-9]+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				num := matches[1]
				// SortKey ist die reine Nummer ohne Punkt am Ende für den Vergleich
				sortKey := strings.TrimSuffix(num, ".")
				return sortKey, num
			}
			// Wenn ein Header ohne Nummer gefunden wird, ans Ende sortieren
			return "999", ""
		}
	}

	return "999", ""
}

// compareVersions vergleicht zwei Versionsnummern wie "1.1.1" und "1.2".
func compareVersions(v1, v2 string) bool {
	if v1 == "999" {
		return false
	}
	if v2 == "999" {
		return true
	}

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		var n1, n2 int
		fmt.Sscanf(parts1[i], "%d", &n1)
		fmt.Sscanf(parts2[i], "%d", &n2)

		if n1 != n2 {
			return n1 < n2
		}
	}

	return len(parts1) < len(parts2)
}
