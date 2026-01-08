package markdown

import (
	"docgen/internal/blocks"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	astTable "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

func Parse(content []byte) ([]blocks.DocBlock, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
	)
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	var docBlocks []blocks.DocBlock

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			docBlocks = append(docBlocks, blocks.HeadingBlock{
				Level: node.Level,
				Text:  string(node.Text(content)),
			})
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph:
			docBlocks = append(docBlocks, blocks.ParagraphBlock{
				Content: parseTextSegments(node, content),
			})
			return ast.WalkSkipChildren, nil
		case *ast.FencedCodeBlock:
			lang := string(node.Language(content))
			codeContent := ""
			for i := 0; i < node.Lines().Len(); i++ {
				line := node.Lines().At(i)
				codeContent += string(line.Value(content))
			}
			if lang == "mermaid" {
				title := ""
				if node.Info != nil {
					info := string(node.Info.Text(content))
					start := strings.Index(info, "{")
					end := strings.Index(info, "}")
					if start != -1 && end != -1 && end > start {
						title = info[start+1 : end]
					}
				}
				docBlocks = append(docBlocks, blocks.MermaidBlock{
					Content: codeContent,
					Title:   title,
				})
			} else {
				docBlocks = append(docBlocks, blocks.CodeBlock{
					Language: lang,
					Content:  codeContent,
				})
			}
			return ast.WalkSkipChildren, nil
		case *ast.List:
			listBlock := blocks.ListBlock{
				Ordered: node.IsOrdered(),
			}
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				if li, ok := child.(*ast.ListItem); ok {
					listBlock.Items = append(listBlock.Items, blocks.ListItem{
						Content: parseTextSegments(li, content),
					})
				}
			}
			docBlocks = append(docBlocks, listBlock)
			return ast.WalkSkipChildren, nil
		case *ast.Image:
			docBlocks = append(docBlocks, blocks.ImageBlock{
				Path: string(node.Destination),
				Alt:  string(node.Text(content)),
			})
			return ast.WalkSkipChildren, nil
		case *ast.ThematicBreak:
			docBlocks = append(docBlocks, blocks.PageBreakBlock{})
			return ast.WalkSkipChildren, nil
		case *astTable.Table:
			table := blocks.TableBlock{}
			for row := node.FirstChild(); row != nil; row = row.NextSibling() {
				if r, ok := row.(*astTable.TableRow); ok {
					var rowData []blocks.TableRow
					for cell := r.FirstChild(); cell != nil; cell = cell.NextSibling() {
						if c, ok := cell.(*astTable.TableCell); ok {
							rowData = append(rowData, blocks.TableRow{
								Content: parseTextSegments(c, content),
								Header:  c.Alignment == astTable.AlignNone, // Simplified
							})
						}
					}
					table.Rows = append(table.Rows, rowData)
				}
			}
			docBlocks = append(docBlocks, table)
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking markdown ast: %w", err)
	}

	return docBlocks, nil
}

func parseTextSegments(n ast.Node, source []byte) []blocks.TextSegment {
	var segments []blocks.TextSegment
	isBold := false
	isItalic := false

	ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if node.Kind() == ast.KindText {
			if entering {
				txt := node.(*ast.Text)
				segments = append(segments, blocks.TextSegment{
					Text:   string(txt.Text(source)),
					Bold:   isBold,
					Italic: isItalic,
				})
			}
		} else if node.Kind() == ast.KindEmphasis {
			em := node.(*ast.Emphasis)
			if entering {
				if em.Level == 1 {
					isItalic = true
				} else if em.Level == 2 {
					isBold = true
				}
			} else {
				if em.Level == 1 {
					isItalic = false
				} else if em.Level == 2 {
					isBold = false
				}
			}
		} else if node.Kind() == ast.KindCodeSpan {
			if entering {
				cs := node.(*ast.CodeSpan)
				segments = append(segments, blocks.TextSegment{
					Text: string(cs.Text(source)),
					Code: true,
				})
				return ast.WalkSkipChildren, nil
			}
		}

		return ast.WalkContinue, nil
	})
	return segments
}
