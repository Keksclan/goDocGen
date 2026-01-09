// Package code bietet Funktionen zum Syntax-Highlighting von Quellcode.
package code

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// CustomTheme repräsentiert ein benutzerdefiniertes Chroma-Theme, das aus einer JSON-Datei geladen werden kann.
type CustomTheme struct {
	Name    string            `json:"name"`
	Entries map[string]string `json:"entries"`
}

// loadCustomTheme lädt ein Theme aus einer JSON-Datei.
func loadCustomTheme(path string) (*chroma.Style, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ct CustomTheme
	if err := json.Unmarshal(data, &ct); err != nil {
		return nil, err
	}

	entries := chroma.StyleEntries{}
	for k, v := range ct.Entries {
		tt, err := chroma.TokenTypeString(k)
		if err != nil {
			if strings.EqualFold(k, "Background") {
				tt = chroma.Background
			} else if strings.EqualFold(k, "Text") {
				tt = chroma.Text
			} else {
				continue
			}
		}
		entries[tt] = v
	}

	style, err := chroma.NewStyle(ct.Name, entries)
	if err != nil {
		return nil, err
	}

	return style, nil
}

// Segment repräsentiert einen Teil des Codes mit spezifischer Formatierung.
type Segment struct {
	Text  string // Der Textinhalt
	Color string // Hex-Farbcode
	Bold  bool   // Fettgedruckt
}

// GetSegments zerlegt den Code in farbige Segmente basierend auf der Sprache und dem gewählten Theme.
// Es gibt auch die Hintergrundfarbe des Themes zurück.
func GetSegments(code, lang, theme string) ([]Segment, string, error) {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	var style *chroma.Style
	if strings.HasSuffix(theme, ".json") {
		var err error
		style, err = loadCustomTheme(theme)
		if err != nil {
			style = styles.Get("catppuccin-latte")
		}
	} else {
		style = styles.Get(theme)
	}

	if style == nil {
		style = styles.Get("catppuccin-latte")
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, "", err
	}

	var segments []Segment
	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)
		color := ""
		if entry.Colour.IsSet() {
			color = entry.Colour.String()
		}

		if token.Type.Category() == chroma.Comment {
			color = "#888888"
		}

		segments = append(segments, Segment{
			Text:  token.Value,
			Color: color,
			Bold:  entry.Bold == chroma.Yes,
		})
	}

	bg := style.Get(chroma.Background).Background.String()

	return segments, bg, nil
}
