// Package pdf implementiert die PDF-Generierung unter Verwendung von gofpdf.
package pdf

import (
	"fmt"
	"godocgen/internal/blocks"
	"godocgen/internal/config"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

// Generator ist die zentrale Komponente zur Erzeugung des PDF-Dokuments.
type Generator struct {
	pdf             *gofpdf.Fpdf      // Die zugrunde liegende PDF-Bibliothek
	cfg             *config.Config    // Die Projektkonfiguration
	blocks          []blocks.DocBlock // Die zu rendernden Inhaltsblöcke
	toc             []TOCEntry        // Gesammelte Inhaltsverzeichniseinträge
	fontDir         string            // Verzeichnis der extrahierten Schriftarten
	totalPages      int               // Gesamtanzahl der Seiten (nach Pass 1)
	headingCounts   []int             // Zähler für die Nummerierung von Überschriften
	registeredFonts map[string]bool   // Verfolgt bereits registrierte Schriftarten
	inTOC           bool              // Status, ob gerade das Inhaltsverzeichnis gerendert wird
}

// TOCEntry repräsentiert einen Eintrag im Inhaltsverzeichnis.
type TOCEntry struct {
	Level  int    // Ebene der Überschrift (1-6)
	Number string // Hierarchische Nummer (z.B. 1.2.3)
	Text   string // Text der Überschrift
	Page   int    // Seitenzahl
	Link   int    // Interner PDF-Link zur Zielseite
}

// NewGenerator erstellt einen neuen PDF-Generator.
func NewGenerator(cfg *config.Config, blocks []blocks.DocBlock, fontDir string) *Generator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(true)
	pdf.SetMargins(cfg.Layout.Margins.Left, cfg.Layout.Margins.Top, cfg.Layout.Margins.Right)
	pdf.SetAutoPageBreak(true, cfg.Layout.Margins.Bottom)

	g := &Generator{
		pdf:             pdf,
		cfg:             cfg,
		blocks:          blocks,
		fontDir:         fontDir,
		headingCounts:   make([]int, 6),
		registeredFonts: make(map[string]bool),
	}

	// Schriften beim Initialisieren registrieren
	g.registerFonts(fontDir)

	return g
}

// resolveFontPath versucht den Pfad einer Schriftart aufzulösen (absolut oder relativ).
func (g *Generator) resolveFontPath(filename string) string {
	if filename == "" {
		return ""
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	// Zuerst im fontDir suchen (Extraktionsverzeichnis oder Projektverzeichnis)
	path := filepath.Join(g.fontDir, filename)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	// Fallback für Windows: Suche im System-Font Verzeichnis
	if os.Getenv("OS") == "Windows_NT" {
		winFont := filepath.Join(os.Getenv("WINDIR"), "Fonts", filename)
		if _, err := os.Stat(winFont); err == nil {
			return winFont
		}
	}
	return path
}

// registerFonts registriert die im Projekt definierten Schriften in der PDF-Bibliothek.
func (g *Generator) registerFonts(fontDir string) {
	regularPath := g.resolveFontPath(g.cfg.Fonts.Regular)
	if _, err := os.Stat(regularPath); err == nil {
		g.pdf.AddUTF8Font("main", "", regularPath)
		g.registeredFonts["main"] = true
	} else {
		fmt.Printf("Warnung: Normale Schriftart nicht unter %s gefunden.\n", regularPath)
	}

	if g.cfg.Fonts.Bold != "" {
		boldPath := g.resolveFontPath(g.cfg.Fonts.Bold)
		if _, err := os.Stat(boldPath); err == nil {
			g.pdf.AddUTF8Font("main", "B", boldPath)
			g.registeredFonts["mainB"] = true
		} else if g.registeredFonts["main"] {
			g.pdf.AddUTF8Font("main", "B", regularPath)
			g.registeredFonts["mainB"] = true
		}
	}
	if g.cfg.Fonts.Italic != "" {
		italicPath := g.resolveFontPath(g.cfg.Fonts.Italic)
		if _, err := os.Stat(italicPath); err == nil {
			g.pdf.AddUTF8Font("main", "I", italicPath)
			g.registeredFonts["mainI"] = true
		} else if g.registeredFonts["main"] {
			g.pdf.AddUTF8Font("main", "I", regularPath)
			g.registeredFonts["mainI"] = true
		}
	}
	if g.registeredFonts["mainB"] && g.registeredFonts["mainI"] {
		boldPath := g.resolveFontPath(g.cfg.Fonts.Bold)
		italicPath := g.resolveFontPath(g.cfg.Fonts.Italic)
		if _, err := os.Stat(boldPath); err == nil {
			g.pdf.AddUTF8Font("main", "BI", boldPath)
			g.registeredFonts["mainBI"] = true
		} else if _, err := os.Stat(italicPath); err == nil {
			g.pdf.AddUTF8Font("main", "BI", italicPath)
			g.registeredFonts["mainBI"] = true
		} else if g.registeredFonts["main"] {
			g.pdf.AddUTF8Font("main", "BI", regularPath)
			g.registeredFonts["mainBI"] = true
		}
	}

	if g.cfg.Fonts.Mono != "" {
		monoPath := g.resolveFontPath(g.cfg.Fonts.Mono)
		if _, err := os.Stat(monoPath); err == nil {
			g.pdf.AddUTF8Font("mono", "", monoPath)
			g.pdf.AddUTF8Font("mono", "I", monoPath)
			g.pdf.AddUTF8Font("mono", "B", monoPath)
			g.pdf.AddUTF8Font("mono", "BI", monoPath)
			g.registeredFonts["mono"] = true
			g.registeredFonts["monoI"] = true
			g.registeredFonts["monoB"] = true
			g.registeredFonts["monoBI"] = true
		} else if g.registeredFonts["main"] {
			g.pdf.AddUTF8Font("mono", "", regularPath)
			g.pdf.AddUTF8Font("mono", "I", regularPath)
			g.pdf.AddUTF8Font("mono", "B", regularPath)
			g.pdf.AddUTF8Font("mono", "BI", regularPath)
			g.registeredFonts["mono"] = true
			g.registeredFonts["monoI"] = true
			g.registeredFonts["monoB"] = true
			g.registeredFonts["monoBI"] = true
		}
	}
}

// Generate führt den zweistufigen Rendering-Prozess aus und speichert das Ergebnis.
func (g *Generator) Generate(outputPath string) error {
	// Durchgang 1: Messen und Sammeln des Inhaltsverzeichnisses
	g.headingCounts = make([]int, 6)
	g.renderAll(true)
	g.totalPages = g.pdf.PageNo()

	// Zurücksetzen für Durchgang 2
	g.pdf = gofpdf.New("P", "mm", "A4", "")
	g.pdf.SetCompression(true)
	g.pdf.SetMargins(g.cfg.Layout.Margins.Left, g.cfg.Layout.Margins.Top, g.cfg.Layout.Margins.Right)
	g.pdf.SetAutoPageBreak(true, g.cfg.Layout.Margins.Bottom)
	g.registeredFonts = make(map[string]bool)
	g.registerFonts(g.fontDir)
	g.headingCounts = make([]int, 6)

	// Durchgang 2: Finales Rendern
	g.renderAll(false)

	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return err
	}

	return g.pdf.OutputFileAndClose(outputPath)
}

// renderAll steuert das Rendern aller Dokumententeile.
func (g *Generator) renderAll(isMeasurement bool) {
	g.setupHeaderFooter()

	// Titelseite
	g.renderFrontPage()

	// Inhaltsverzeichnis
	if g.cfg.TOC.Enabled {
		g.renderTOC(isMeasurement)
	}

	// Inhalt (Blöcke)
	for _, block := range g.blocks {
		g.renderBlock(block, isMeasurement)
	}
}
