// Package config enthält die Konfigurationsmodelle und Ladelogik für godocgen.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config repräsentiert die Hauptkonfiguration für ein Dokumentationsprojekt.
// Sie definiert das Erscheinungsbild, die Schriften und das Layout des generierten PDFs.
type Config struct {
	Title       string      `yaml:"title" validate:"required"`          // Haupttitel des Dokuments
	Subtitle    string      `yaml:"subtitle"`                           // Untertitel für das Deckblatt
	Author      string      `yaml:"author"`                             // Autor des Dokuments
	Header      Header      `yaml:"header"`                             // Header-Konfiguration
	Footer      Footer      `yaml:"footer"`                             // Footer-Konfiguration
	Colors      Colors      `yaml:"colors"`                             // Farbschema
	Theme       string      `yaml:"theme"`                              // Vorbelegtes Theme (z.B. catppuccin-mocha)
	Fonts       Fonts       `yaml:"fonts" validate:"required"`          // Schriftarten-Konfiguration
	FontSize    float64     `yaml:"font_size" validate:"required,gt=0"` // Standard-Schriftgröße
	PageNumbers PageNumbers `yaml:"page_numbers"`                       // Seitennummerierungseinstellungen
	Layout      Layout      `yaml:"layout"`                             // Layout-Vorgaben (Ränder, Ausrichtung)
	Gradient    Gradient    `yaml:"gradient"`                           // Hintergrund-Farbverläufe
	CodeTheme   string      `yaml:"code_theme"`                         // Theme für Code-Highlighting
	Code        Code        `yaml:"code"`                               // Code-Block-Einstellungen
	Mermaid     Mermaid     `yaml:"mermaid"`                            // Mermaid-Diagramm-Konfiguration
	TOC         TOC         `yaml:"toc"`                                // Inhaltsverzeichnis-Einstellungen
}

// TOC definiert Einstellungen für das Inhaltsverzeichnis.
type TOC struct {
	Enabled      bool    `yaml:"enabled"`       // Inhaltsverzeichnis anzeigen
	ShowNumbers  bool    `yaml:"show_numbers"`  // Nummern im TOC anzeigen
	ShowDots     bool    `yaml:"show_dots"`     // Punkte zwischen Text und Seite anzeigen
	LineSpacing  float64 `yaml:"line_spacing"`  // Zeilenabstand im TOC (z.B. 1.0 für kompakt, 1.5 für mehr Abstand)
	BoldHeadings bool    `yaml:"bold_headings"` // Überschriften fett darstellen
	FontSize     float64 `yaml:"font_size"`     // Schriftgröße für TOC-Einträge (0 = Standard)
	Indent       float64 `yaml:"indent"`        // Einrückung pro Level in mm (Standard: 8)
}

// Save speichert die Konfiguration in eine YAML-Datei.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Gradient definiert Einstellungen für farbige Verläufe im Dokument.
type Gradient struct {
	Enabled     bool   `yaml:"enabled"`     // Aktiviert den Farbverlauf
	Start       string `yaml:"start"`       // Hex-Code der Startfarbe
	End         string `yaml:"end"`         // Hex-Code der Endfarbe
	Orientation string `yaml:"orientation"` // "vertical" oder "horizontal"
	Global      bool   `yaml:"global"`      // Wenn wahr, wird der Verlauf auf allen Seiten angewendet
}

// Header definiert den Text oder das Bild im oberen Bereich jeder Seite.
type Header struct {
	Text  string `yaml:"text"`  // Text im Header
	Image string `yaml:"image"` // Pfad zu einer Bilddatei für den Header
}

// Footer definiert den Text oder das Bild im unteren Bereich jeder Seite.
type Footer struct {
	Text   string `yaml:"text"`   // Veraltet: Text im Footer (nutze Left/Center/Right)
	Image  string `yaml:"image"`  // Pfad zu einer Bilddatei für den Footer
	Left   string `yaml:"left"`   // Inhalt linke Zone
	Center string `yaml:"center"` // Inhalt mittlere Zone
	Right  string `yaml:"right"`  // Inhalt rechte Zone
}

// Colors definiert die im Dokument verwendeten Farben.
type Colors struct {
	Title      string `yaml:"title"`      // Farbe für Überschriften
	Header     string `yaml:"header"`     // Farbe für Header-Text
	Background string `yaml:"background"` // Seitenhintergrundfarbe
	Text       string `yaml:"text"`       // Standard-Textfarbe
	Accent     string `yaml:"accent"`     // Farbe für Akzente
}

// Fonts definiert die zu verwendenden Schriftarten.
type Fonts struct {
	Zip     string `yaml:"zip"`                          // Pfad zu einem ZIP mit TTF-Dateien (optional bei System-Fonts)
	URL     string `yaml:"url" validate:"omitempty,url"` // URL zum Download eines Font-ZIPs
	Regular string `yaml:"regular" validate:"required"`  // Dateiname oder absoluter Pfad zur Schriftart
	Bold    string `yaml:"bold"`                         // Dateiname oder absoluter Pfad
	Italic  string `yaml:"italic"`                       // Dateiname oder absoluter Pfad
	Mono    string `yaml:"mono"`                         // Dateiname oder absoluter Pfad
}

// PageNumbers steuert die Anzeige von Seitenzahlen.
type PageNumbers struct {
	StartPage int `yaml:"start_page"` // Ab welcher Seite die Zählung beginnt
}

// Layout definiert die räumliche Anordnung der Elemente.
type Layout struct {
	StartPage       string  `yaml:"startpage" validate:"oneof=left center right justify"` // Ausrichtung des Deckblatts
	Body            string  `yaml:"body" validate:"oneof=left center right justify"`      // Standard-Textausrichtung
	Margins         Margins `yaml:"margins"`                                              // Seitenränder
	HeaderNumbering bool    `yaml:"header_numbering"`                                     // Automatische Nummerierung von Überschriften
	LineSpacing     float64 `yaml:"line_spacing" validate:"omitempty,gt=0"`               // Zeilenabstand (z.B. 1.5)
	FooterStyle     string  `yaml:"footer_style" validate:"omitempty,oneof=fixed inline"` // "fixed" (unten) oder "inline" (nach Content)
}

// Margins definiert die Seitenränder in Millimetern.
type Margins struct {
	Left   float64 `yaml:"left"`
	Right  float64 `yaml:"right"`
	Top    float64 `yaml:"top"`
	Bottom float64 `yaml:"bottom"`
}

// Mermaid definiert Einstellungen für die Diagramm-Generierung.
type Mermaid struct {
	Renderer string  `yaml:"renderer"` // Renderer-Typ ("mmdc" oder leer für Chrome-Fallback)
	Width    float64 `yaml:"width"`    // Breite der Diagramme in mm (0 = automatisch)
	Scale    float64 `yaml:"scale"`    // Skalierungsfaktor für Diagramme (z.B. 0.8 für 80%)
}

// Code definiert Einstellungen für Code-Blöcke.
type Code struct {
	FontSize    float64 `yaml:"font_size"`     // Standard-Schriftgröße für Code (0 = nutzt globale FontSize)
	MinFontSize float64 `yaml:"min_font_size"` // Minimale Schriftgröße bei AutoScale (Standard: 6)
	AutoScale   bool    `yaml:"auto_scale"`    // Automatische Schriftgrößenanpassung für große Code-Blöcke
	MaxLines    int     `yaml:"max_lines"`     // Ab dieser Zeilenanzahl wird skaliert (Standard: 30)
	MaxLineLen  int     `yaml:"max_line_len"`  // Ab dieser Zeilenlänge wird skaliert (Standard: 80)
}
