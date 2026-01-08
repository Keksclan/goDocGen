package code

import (
	"docgen/internal/util"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type CustomTheme struct {
	Name    string            `json:"name"`
	Entries map[string]string `json:"entries"`
}

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
			// Try with Background, etc.
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

func Highlight(code, lang, theme, cacheDir string) (string, error) {
	hash := util.HashString(code + lang + theme)
	cachePath := filepath.Join(cacheDir, "code", hash+".png")

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	os.MkdirAll(filepath.Dir(cachePath), 0755)

	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get(theme)
	if style == nil {
		style = styles.Get("github")
	}

	// We'll use a simple approach: since we can't easily render Chroma to PDF directly with all styles,
	// we use a workaround or a simplified text rendering in the PDF engine later.
	// But the requirement says "robuste Lösung".
	// Chroma doesn't have a built-in PNG formatter in v2 anymore (it was in a sub-package).
	// Let's use the text-based segments and let the PDF engine handle it, 
	// or use a simple black & white fallback for now if it gets too complex.

	// Actually, I will implement a "Block" structure where CodeBlock contains Segments (Text + Color).
	return "", nil
}

type Segment struct {
	Text  string
	Color string // hex
	Bold  bool
}

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
			// Fallback if file not found or invalid
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

		// Force comments to be grey if requested or if it's a comment
		if token.Type.Category() == chroma.Comment {
			color = "#888888" // Medium grey
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
