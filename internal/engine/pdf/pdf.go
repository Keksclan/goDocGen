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

// registerFonts registriert die im Projekt definierten Schriften in der PDF-Bibliothek.
func (g *Generator) registerFonts(fontDir string) {
	regularPath := filepath.Join(fontDir, g.cfg.Fonts.Regular)
	if _, err := os.Stat(regularPath); err == nil {
		g.pdf.AddUTF8Font("Main", "", regularPath)
		g.registeredFonts["Main"] = true
	} else {
		fmt.Printf("Warnung: Normale Schriftart nicht unter %s gefunden.\n", regularPath)
	}

	if g.cfg.Fonts.Bold != "" {
		boldPath := filepath.Join(fontDir, g.cfg.Fonts.Bold)
		if _, err := os.Stat(boldPath); err == nil {
			g.pdf.AddUTF8Font("Main", "B", boldPath)
			g.registeredFonts["MainB"] = true
		} else if g.registeredFonts["Main"] {
			g.pdf.AddUTF8Font("Main", "B", regularPath)
			g.registeredFonts["MainB"] = true
		}
	}
	if g.cfg.Fonts.Italic != "" {
		italicPath := filepath.Join(fontDir, g.cfg.Fonts.Italic)
		if _, err := os.Stat(italicPath); err == nil {
			g.pdf.AddUTF8Font("Main", "I", italicPath)
			g.registeredFonts["MainI"] = true
		} else if g.registeredFonts["Main"] {
			g.pdf.AddUTF8Font("Main", "I", regularPath)
			g.registeredFonts["MainI"] = true
		}
	}
	if g.registeredFonts["MainB"] && g.registeredFonts["MainI"] {
		boldPath := filepath.Join(fontDir, g.cfg.Fonts.Bold)
		italicPath := filepath.Join(fontDir, g.cfg.Fonts.Italic)
		if _, err := os.Stat(boldPath); err == nil {
			g.pdf.AddUTF8Font("Main", "BI", boldPath)
			g.registeredFonts["MainBI"] = true
		} else if _, err := os.Stat(italicPath); err == nil {
			g.pdf.AddUTF8Font("Main", "BI", italicPath)
			g.registeredFonts["MainBI"] = true
		} else if g.registeredFonts["Main"] {
			g.pdf.AddUTF8Font("Main", "BI", regularPath)
			g.registeredFonts["MainBI"] = true
		}
	}

	if g.cfg.Fonts.Mono != "" {
		monoPath := filepath.Join(fontDir, g.cfg.Fonts.Mono)
		if _, err := os.Stat(monoPath); err == nil {
			g.pdf.AddUTF8Font("Mono", "", monoPath)
			g.pdf.AddUTF8Font("Mono", "I", monoPath)
			g.pdf.AddUTF8Font("Mono", "B", monoPath)
			g.pdf.AddUTF8Font("Mono", "BI", monoPath)
			g.registeredFonts["Mono"] = true
			g.registeredFonts["MonoI"] = true
			g.registeredFonts["MonoB"] = true
			g.registeredFonts["MonoBI"] = true
		} else if g.registeredFonts["Main"] {
			g.pdf.AddUTF8Font("Mono", "", regularPath)
			g.pdf.AddUTF8Font("Mono", "I", regularPath)
			g.pdf.AddUTF8Font("Mono", "B", regularPath)
			g.pdf.AddUTF8Font("Mono", "BI", regularPath)
			g.registeredFonts["Mono"] = true
			g.registeredFonts["MonoI"] = true
			g.registeredFonts["MonoB"] = true
			g.registeredFonts["MonoBI"] = true
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
