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

func init() {
	// IHK Theme: Weißer Hintergrund mit blauer Syntax-Hervorhebung
	// Professionelles, sauberes Design passend für IHK-Dokumentationen
	ihkStyle := styles.Register(chroma.MustNewStyle("ihk", chroma.StyleEntries{
		chroma.Background:         "bg:#ffffff",
		chroma.Text:               "#1a1a1a",
		chroma.Keyword:            "#0066cc bold",
		chroma.KeywordConstant:    "#0066cc bold",
		chroma.KeywordDeclaration: "#0066cc bold",
		chroma.KeywordNamespace:   "#0066cc bold",
		chroma.KeywordPseudo:      "#0066cc",
		chroma.KeywordReserved:    "#0066cc bold",
		chroma.KeywordType:        "#0077aa bold",
		chroma.Name:               "#1a1a1a",
		chroma.NameBuiltin:        "#0077aa",
		chroma.NameBuiltinPseudo:  "#0077aa",
		chroma.NameClass:          "#0055aa bold",
		chroma.NameConstant:       "#003366",
		chroma.NameDecorator:      "#0066cc",
		chroma.NameEntity:         "#0066cc",
		chroma.NameException:      "#cc0000",
		chroma.NameFunction:       "#0055aa",
		chroma.NameLabel:          "#003366",
		chroma.NameNamespace:      "#0055aa",
		chroma.NameTag:            "#0066cc bold",
		chroma.NameVariable:       "#003366",
		chroma.NameVariableClass:  "#003366",
		chroma.NameVariableGlobal: "#003366",
		chroma.Literal:            "#0066cc",
		chroma.LiteralDate:        "#0066cc",
		chroma.String:             "#008844",
		chroma.StringAffix:        "#008844",
		chroma.StringBacktick:     "#008844",
		chroma.StringChar:         "#008844",
		chroma.StringDelimiter:    "#008844",
		chroma.StringDoc:          "#666666 italic",
		chroma.StringDouble:       "#008844",
		chroma.StringEscape:       "#0066cc",
		chroma.StringHeredoc:      "#008844",
		chroma.StringInterpol:     "#0066cc",
		chroma.StringOther:        "#008844",
		chroma.StringRegex:        "#0066cc",
		chroma.StringSingle:       "#008844",
		chroma.StringSymbol:       "#008844",
		chroma.Number:             "#cc6600",
		chroma.NumberBin:          "#cc6600",
		chroma.NumberFloat:        "#cc6600",
		chroma.NumberHex:          "#cc6600",
		chroma.NumberInteger:      "#cc6600",
		chroma.NumberIntegerLong:  "#cc6600",
		chroma.NumberOct:          "#cc6600",
		chroma.Operator:           "#0066cc",
		chroma.OperatorWord:       "#0066cc bold",
		chroma.Punctuation:        "#1a1a1a",
		chroma.Comment:            "#888888 italic",
		chroma.CommentHashbang:    "#888888 italic",
		chroma.CommentMultiline:   "#888888 italic",
		chroma.CommentPreproc:     "#0066cc",
		chroma.CommentPreprocFile: "#008844",
		chroma.CommentSingle:      "#888888 italic",
		chroma.CommentSpecial:     "#0066cc italic",
		chroma.Generic:            "#1a1a1a",
		chroma.GenericDeleted:     "#cc0000",
		chroma.GenericEmph:        "italic",
		chroma.GenericError:       "#cc0000",
		chroma.GenericHeading:     "#0055aa bold",
		chroma.GenericInserted:    "#008844",
		chroma.GenericOutput:      "#666666",
		chroma.GenericPrompt:      "#0055aa bold",
		chroma.GenericStrong:      "bold",
		chroma.GenericSubheading:  "#0055aa",
		chroma.GenericTraceback:   "#cc0000",
	}))
	_ = ihkStyle // Verhindert "unused variable" Warnung
}

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
