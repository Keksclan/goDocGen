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
	// IHK Theme: Weißer Hintergrund mit professioneller Syntax-Hervorhebung
	// Ausgewogene Farbpalette für gute Lesbarkeit und Unterscheidbarkeit
	// Optimiert für Go-Code
	ihkStyle := styles.Register(chroma.MustNewStyle("ihk", chroma.StyleEntries{
		// Hintergrund und Basis-Text
		chroma.Background: "bg:#ffffff",
		chroma.Text:       "#24292e", // Dunkles Grau für normalen Text

		// Keywords - Blau für Go-Keywords (func, package, import, var, const, type, etc.)
		chroma.Keyword:            "#0000ff bold", // Klassisches Blau für Keywords
		chroma.KeywordConstant:    "#0000ff bold", // true, false, nil, iota
		chroma.KeywordDeclaration: "#0000ff bold", // func, type, var, const, struct, interface
		chroma.KeywordNamespace:   "#0000ff bold", // package, import
		chroma.KeywordPseudo:      "#0000ff",      // Pseudo-Keywords
		chroma.KeywordReserved:    "#0000ff bold", // Reservierte Wörter
		chroma.KeywordType:        "#267f99 bold", // Go-Typen: int, string, bool - Türkis/Teal

		// Namen - Verschiedene Farben für gute Unterscheidung
		chroma.Name:                 "#24292e",      // Normale Bezeichner - dunkelgrau
		chroma.NameBuiltin:          "#795e26",      // Go-Builtins: make, len, cap - Braun/Gold
		chroma.NameBuiltinPseudo:    "#795e26",      // Pseudo-Builtins
		chroma.NameClass:            "#267f99 bold", // Struct/Interface-Namen - Türkis
		chroma.NameConstant:         "#0070c1",      // Konstanten - Hellblau
		chroma.NameDecorator:        "#795e26",      // Decorators
		chroma.NameEntity:           "#24292e",      // Entities
		chroma.NameException:        "#d73a49",      // Errors/Exceptions - Rot
		chroma.NameFunction:         "#795e26",      // Funktionsnamen - Braun/Gold
		chroma.NameLabel:            "#6f42c1",      // Labels - Lila
		chroma.NameNamespace:        "#24292e",      // Package-Namen - dunkelgrau
		chroma.NameOther:            "#24292e",      // Andere Namen
		chroma.NameTag:              "#22863a",      // Tags - Grün
		chroma.NameVariable:         "#24292e",      // Variablen - dunkelgrau
		chroma.NameVariableClass:    "#24292e",      // Klassen-Variablen
		chroma.NameVariableGlobal:   "#24292e",      // Globale Variablen
		chroma.NameVariableInstance: "#24292e",      // Instanz-Variablen

		// Literale
		chroma.Literal:     "#098658", // Grün für Literale
		chroma.LiteralDate: "#098658",

		// Strings - Dunkelrot/Braun für gute Lesbarkeit
		chroma.String:          "#a31515", // Strings in Dunkelrot
		chroma.StringAffix:     "#a31515",
		chroma.StringBacktick:  "#a31515", // Raw strings in Go
		chroma.StringChar:      "#a31515", // Rune literals
		chroma.StringDelimiter: "#a31515",
		chroma.StringDoc:       "#6a737d italic", // Doc-Strings - Grau
		chroma.StringDouble:    "#a31515",
		chroma.StringEscape:    "#0000ff bold", // Escape-Sequenzen - Blau
		chroma.StringHeredoc:   "#a31515",
		chroma.StringInterpol:  "#a31515",
		chroma.StringOther:     "#a31515",
		chroma.StringRegex:     "#811f3f", // Regex - Dunkelrot
		chroma.StringSingle:    "#a31515",
		chroma.StringSymbol:    "#a31515",

		// Zahlen - Grün für guten Kontrast
		chroma.Number:            "#098658", // Zahlen in Grün
		chroma.NumberBin:         "#098658",
		chroma.NumberFloat:       "#098658",
		chroma.NumberHex:         "#098658",
		chroma.NumberInteger:     "#098658",
		chroma.NumberIntegerLong: "#098658",
		chroma.NumberOct:         "#098658",

		// Operatoren - Dunkelgrau
		chroma.Operator:     "#24292e", // Operatoren: +, -, *, /, :=, ==, etc.
		chroma.OperatorWord: "#0000ff", // Wort-Operatoren - Blau

		// Interpunktion - Dunkelgrau für Struktur
		chroma.Punctuation: "#24292e", // Klammern, Kommas, etc.

		// Kommentare - Grün und kursiv (wie in vielen IDEs)
		chroma.Comment:            "#008000 italic", // Einzeilige Kommentare - Grün
		chroma.CommentHashbang:    "#008000 italic",
		chroma.CommentMultiline:   "#008000 italic", // Mehrzeilige Kommentare
		chroma.CommentPreproc:     "#0000ff",        // Präprozessor - Blau
		chroma.CommentPreprocFile: "#a31515",        // Präprozessor-Datei - Dunkelrot
		chroma.CommentSingle:      "#008000 italic",
		chroma.CommentSpecial:     "#008000 bold italic", // Spezielle Kommentare (TODO, FIXME)

		// Generische Styles
		chroma.Generic:           "#24292e",
		chroma.GenericDeleted:    "#d73a49 bg:#ffeef0", // Gelöschte Zeilen - Rot
		chroma.GenericEmph:       "italic",
		chroma.GenericError:      "#d73a49 bold",       // Fehler - Rot
		chroma.GenericHeading:    "#0000ff bold",       // Überschriften - Blau
		chroma.GenericInserted:   "#22863a bg:#e6ffed", // Eingefügte Zeilen - Grün
		chroma.GenericOutput:     "#6a737d",            // Output - Grau
		chroma.GenericPrompt:     "#0000ff bold",       // Prompt - Blau
		chroma.GenericStrong:     "bold",
		chroma.GenericSubheading: "#0000ff", // Unterüberschriften - Blau
		chroma.GenericTraceback:  "#d73a49", // Traceback - Rot
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
