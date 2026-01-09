package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// LoadConfig lädt die Konfiguration aus einer YAML-Datei, wendet Themes an und setzt Standardwerte.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Lesen der Konfigurationsdatei: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("Fehler beim Parsen von YAML: %w", err)
	}

	if cfg.Theme == "" {
		cfg.Theme = "catppuccin-mocha"
	}
	ApplyTheme(cfg)
	setDefaults(cfg)

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("Validierungsfehler: %w", err)
	}

	return cfg, nil
}

// ApplyTheme wendet vordefinierte Farbpaletten (Catppuccin) auf die Konfiguration an,
// sofern keine expliziten Farben definiert wurden.
func ApplyTheme(cfg *Config) {
	colors := GetThemeColors(cfg.Theme)
	if cfg.Colors.Title == "" {
		cfg.Colors.Title = colors.Title
	}
	if cfg.Colors.Header == "" {
		cfg.Colors.Header = colors.Header
	}
	if cfg.Colors.Background == "" {
		cfg.Colors.Background = colors.Background
	}
	if cfg.Colors.Text == "" {
		cfg.Colors.Text = colors.Text
	}
	if cfg.Colors.Accent == "" {
		cfg.Colors.Accent = colors.Accent
	}
	if cfg.CodeTheme == "" {
		cfg.CodeTheme = cfg.Theme
	}
}

// GetThemeColors gibt die Standardfarben für ein gegebenes Theme zurück.
func GetThemeColors(theme string) Colors {
	switch theme {
	case "catppuccin-latte":
		return Colors{
			Title:      "#8839ef",
			Header:     "#1e66f5",
			Background: "#eff1f5",
			Text:       "#4c4f69",
			Accent:     "#ea76cb",
		}
	case "catppuccin-frappe":
		return Colors{
			Title:      "#ca9ee6",
			Header:     "#8caaee",
			Background: "#303446",
			Text:       "#c6d0f5",
			Accent:     "#f4b8e4",
		}
	case "catppuccin-macchiato":
		return Colors{
			Title:      "#c6a0f6",
			Header:     "#8aadf4",
			Background: "#24273a",
			Text:       "#cad3f5",
			Accent:     "#f5bde6",
		}
	case "catppuccin-mocha":
		return Colors{
			Title:      "#cba6f7",
			Header:     "#89b4fa",
			Background: "#1e1e2e",
			Text:       "#cdd6f4",
			Accent:     "#f5c2e7",
		}
	case "github-light":
		return Colors{
			Title:      "#0550ae",
			Header:     "#24292f",
			Background: "#ffffff",
			Text:       "#24292f",
			Accent:     "#cf222e",
		}
	case "github-dark":
		return Colors{
			Title:      "#79c0ff",
			Header:     "#c9d1d9",
			Background: "#0d1117",
			Text:       "#c9d1d9",
			Accent:     "#ff7b72",
		}
	case "solarized-light":
		return Colors{
			Title:      "#268bd2",
			Header:     "#586e75",
			Background: "#fdf6e3",
			Text:       "#657b83",
			Accent:     "#d33682",
		}
	case "nord":
		return Colors{
			Title:      "#88c0d0",
			Header:     "#81a1c1",
			Background: "#2e3440",
			Text:       "#d8dee9",
			Accent:     "#b48ead",
		}
	case "dracula":
		return Colors{
			Title:      "#bd93f9",
			Header:     "#6272a4",
			Background: "#282a36",
			Text:       "#f8f8f2",
			Accent:     "#ff79c6",
		}
	case "ayu-light":
		return Colors{
			Title:      "#f5222d",
			Header:     "#5c6166",
			Background: "#fafafa",
			Text:       "#5c6166",
			Accent:     "#f29718",
		}
	case "tango-light":
		return Colors{
			Title:      "#204a87",
			Header:     "#2e3436",
			Background: "#eeeeec",
			Text:       "#2e3436",
			Accent:     "#ce5c00",
		}
	case "gruvbox-light":
		return Colors{
			Title:      "#af3a03",
			Header:     "#3c3836",
			Background: "#fbf1c7",
			Text:       "#3c3836",
			Accent:     "#427b58",
		}
	case "one-light":
		return Colors{
			Title:      "#4078f2",
			Header:     "#383a42",
			Background: "#fafafa",
			Text:       "#383a42",
			Accent:     "#e45649",
		}
	case "nord-light":
		return Colors{
			Title:      "#5e81ac",
			Header:     "#4c566a",
			Background: "#eceff4",
			Text:       "#2e3440",
			Accent:     "#88c0d0",
		}
	default:
		return Colors{
			Title:  "#1e66f5",
			Header: "#1e66f5",
		}
	}
}

// GetAvailableThemes gibt eine Liste aller eingebauten Themes zurück.
func GetAvailableThemes() []string {
	return []string{
		"catppuccin-mocha",
		"catppuccin-latte",
		"catppuccin-frappe",
		"catppuccin-macchiato",
		"github-light",
		"github-dark",
		"solarized-light",
		"nord",
		"dracula",
		"ayu-light",
		"tango-light",
		"gruvbox-light",
		"one-light",
		"nord-light",
	}
}

// setDefaults setzt Standardwerte für Layout, Ränder und Schriften, falls diese fehlen.
func setDefaults(cfg *Config) {
	if cfg.Colors.Title == "" {
		cfg.Colors.Title = "#1e66f5" // Default Blue (statt E.ON Red)
	}
	if cfg.Colors.Header == "" {
		cfg.Colors.Header = "#1e66f5" // Default Blue
	}
	if cfg.Mermaid.Renderer == "" {
		cfg.Mermaid.Renderer = "mmdc"
	}
	if cfg.Layout.StartPage == "" {
		cfg.Layout.StartPage = "center"
	}
	if cfg.Layout.Body == "" {
		cfg.Layout.Body = "justify"
	}
	// Default: Nummerierung an
	cfg.Layout.HeaderNumbering = true

	if cfg.Layout.Margins.Left == 0 {
		cfg.Layout.Margins.Left = 10
	}
	if cfg.Layout.Margins.Right == 0 {
		cfg.Layout.Margins.Right = 10
	}
	if cfg.Layout.Margins.Top == 0 {
		cfg.Layout.Margins.Top = 10
	}
	if cfg.Layout.Margins.Bottom == 0 {
		cfg.Layout.Margins.Bottom = 10
	}
	if cfg.Gradient.Orientation == "" {
		cfg.Gradient.Orientation = "vertical"
	}
	if cfg.FontSize == 0 {
		cfg.FontSize = 12.0
	}
	if cfg.CodeTheme == "" {
		cfg.CodeTheme = "catppuccin-latte"
	}

	// TOC Defaults
	if !cfg.TOC.Enabled {
		// Standardmäßig an, wenn nicht explizit aus
		cfg.TOC.Enabled = true
	}
	// Wir können hier nicht einfach auf false prüfen, da bool default false ist.
	// In Go ist es schwer zu unterscheiden zwischen "nicht gesetzt" und "false gesetzt" bei bools ohne Zeiger.
	// Aber für TOC.Enabled wollen wir meistens true.
	// Eigentlich sollten wir in der docgen.yml schauen, ob es da ist.
	// Für jetzt setzen wir vernünftige Standards.
	cfg.TOC.ShowNumbers = true
	cfg.TOC.ShowDots = true

	// Footer Defaults
	if cfg.Footer.Left == "" && cfg.Footer.Center == "" && cfg.Footer.Right == "" {
		if cfg.Footer.Text != "" {
			cfg.Footer.Left = cfg.Footer.Text
		}
		cfg.Footer.Right = "{page} / {total}"
	}
}
