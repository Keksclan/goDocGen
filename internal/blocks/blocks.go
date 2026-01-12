// Package blocks definiert die verschiedenen Inhaltselemente eines Dokuments.
package blocks

import "godocgen/internal/engine/code"

// DocBlock ist das Interface für alle renderbaren Elemente im Dokument.
type DocBlock interface {
	IsBlock()
}

// HeadingBlock repräsentiert eine Überschrift.
type HeadingBlock struct {
	Level           int    // Ebene der Überschrift (1-6)
	Text            string // Textinhalt
	ParentNumbering string // Basis-Nummerierung aus der Ordnerstruktur (z.B. "1.1_")
	AnchorID        string // Eindeutige ID für Anchor-Links (z.B. "einführung-in-docgen")
	ExcludeFromTOC  bool   // Wenn true, wird die Überschrift nicht im TOC angezeigt und nicht nummeriert (Syntax: !#! oder !##!)
}

func (h HeadingBlock) IsBlock() {}

// ParagraphBlock repräsentiert einen Textabsatz.
type ParagraphBlock struct {
	Content []TextSegment // Liste der formatierten Textsegmente
}

func (p ParagraphBlock) IsBlock() {}

// TextSegment repräsentiert einen Teil eines Textes mit Formatierung.
type TextSegment struct {
	Text          string // Textinhalt
	Italic        bool   // Kursiv
	Bold          bool   // Fett
	Strikethrough bool   // Durchgestrichen
	Code          bool   // Inline-Code
	Link          string // URL oder lokaler Pfad
}

// ImageBlock repräsentiert ein Bild.
type ImageBlock struct {
	Path  string  // Dateipfad zum Bild
	Alt   string  // Alternativtext
	Title string  // Optionaler Bildtitel
	Width float64 // Optionale Breite in mm (0 = automatisch)
	Scale float64 // Optionaler Skalierungsfaktor (z.B. 0.8 für 80%)
}

func (i ImageBlock) IsBlock() {}

// MermaidBlock repräsentiert ein Mermaid-Diagramm.
type MermaidBlock struct {
	Content string // Mermaid-Syntax Quellcode
	Title   string // Optionaler Titel des Diagramms
}

func (m MermaidBlock) IsBlock() {}

// CodeBlock repräsentiert einen mehrzeiligen Code-Abschnitt.
type CodeBlock struct {
	Language string         // Programmiersprache (für Highlighting)
	Content  string         // Quellcode als Text
	Segments []code.Segment // Farbig formatierte Segmente (nach Highlighting)
	BgColor  string         // Hintergrundfarbe des Blocks
}

func (c CodeBlock) IsBlock() {}

// ListBlock repräsentiert eine ungeordnete oder geordnete Liste.
type ListBlock struct {
	Items   []ListItem // Einträge der Liste
	Ordered bool       // Wahr, wenn die Liste nummeriert ist
}

func (l ListBlock) IsBlock() {}

// ListItem repräsentiert einen einzelnen Listeneintrag.
type ListItem struct {
	Content []TextSegment
	SubList *ListBlock // Optionale verschachtelte Liste
}

// PageBreakBlock erzwingt einen Seitenumbruch im Dokument.
type PageBreakBlock struct{}

func (p PageBreakBlock) IsBlock() {}

// BlockquoteBlock repräsentiert ein Zitat (Blockquote).
type BlockquoteBlock struct {
	Content []DocBlock // Inhalt des Zitats (kann Paragraphen, Listen etc. enthalten)
}

func (b BlockquoteBlock) IsBlock() {}

// TableBlock repräsentiert eine Tabelle.
type TableBlock struct {
	Rows [][]TableRow // Zweidimensionale Liste der Tabellenzellen
}

// TableRow repräsentiert eine Zelle in einer Tabellenzeile.
type TableRow struct {
	Content []TextSegment
	Header  bool // Wahr, wenn die Zelle als Kopfzeile formatiert werden soll
}

func (t TableBlock) IsBlock() {}
