package blocks

import "docgen/internal/engine/code"

type DocBlock interface {
	IsBlock()
}

type HeadingBlock struct {
	Level int
	Text  string
}

func (h HeadingBlock) IsBlock() {}

type ParagraphBlock struct {
	Content []TextSegment
}

func (p ParagraphBlock) IsBlock() {}

type TextSegment struct {
	Text   string
	Italic bool
	Bold   bool
	Code   bool
}

type ImageBlock struct {
	Path  string
	Alt   string
	Title string
}

func (i ImageBlock) IsBlock() {}

type MermaidBlock struct {
	Content string
	Title   string
}

func (m MermaidBlock) IsBlock() {}

type CodeBlock struct {
	Language string
	Content  string
	Segments []code.Segment
	BgColor  string
}

func (c CodeBlock) IsBlock() {}

type ListBlock struct {
	Items   []ListItem
	Ordered bool
}

func (l ListBlock) IsBlock() {}

type ListItem struct {
	Content []TextSegment
}

type PageBreakBlock struct{}

func (p PageBreakBlock) IsBlock() {}

type TableBlock struct {
	Rows [][]TableRow
}

type TableRow struct {
	Content []TextSegment
	Header  bool
}

func (t TableBlock) IsBlock() {}
