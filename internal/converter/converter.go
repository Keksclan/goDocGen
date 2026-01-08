package converter

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	ghml "github.com/yuin/goldmark/renderer/html"
)

type MarkdownConverter struct {
	gm goldmark.Markdown
}

func NewMarkdownConverter() *MarkdownConverter {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.DefinitionList,
			extension.Footnote,
			&MermaidExtension{},
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			ghml.WithUnsafe(),
		),
	)
	return &MarkdownConverter{gm: gm}
}

func (c *MarkdownConverter) Convert(md []byte) (string, error) {
	var buf bytes.Buffer
	if err := c.gm.Convert(md, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c *MarkdownConverter) ConvertChapters(chapters []string) (string, error) {
	var fullHTML bytes.Buffer
	for i, content := range chapters {
		html, err := c.Convert([]byte(content))
		if err != nil {
			return "", fmt.Errorf("error converting chapter %d: %w", i, err)
		}
		fullHTML.WriteString("<div class='chapter'>")
		fullHTML.WriteString(html)
		fullHTML.WriteString("</div>")
		if i < len(chapters)-1 {
			fullHTML.WriteString("<div class='page-break'></div>")
		}
	}
	return fullHTML.String(), nil
}
