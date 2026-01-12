// Package markdown bietet Funktionen zum Parsen von Markdown-Dateien in interne Dokumentblöcke.
package markdown

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"godocgen/internal/blocks"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extAst "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// excludeFromTOCMarker ist ein spezieller Marker, der anzeigt, dass eine Überschrift
// nicht im Inhaltsverzeichnis erscheinen und nicht nummeriert werden soll.
const excludeFromTOCMarker = "\x00EXCLUDE_TOC\x00"

// preprocessExcludedHeadings wandelt die !#! Syntax in normale Headings um und fügt einen Marker hinzu.
// Beispiel: "!#! Überschrift" -> "# \x00EXCLUDE_TOC\x00Überschrift"
// Beispiel: "!###!Blackboxtesting" -> "### \x00EXCLUDE_TOC\x00Blackboxtesting"
// Unterstützt !#!, !##!, !###!, etc. mit oder ohne Leerzeichen nach dem letzten !
func preprocessExcludedHeadings(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	var result []string

	// Regex für !#! bis !######! Syntax (Leerzeichen nach ! ist optional)
	excludeHeadingRegex := regexp.MustCompile(`^(!)(#{1,6})(!) ?(.*)$`)

	for _, line := range lines {
		if matches := excludeHeadingRegex.FindStringSubmatch(line); matches != nil {
			// matches[2] = die # Zeichen, matches[4] = der Text
			hashes := matches[2]
			text := matches[4]
			// Umwandeln in normales Heading mit Marker
			result = append(result, hashes+" "+excludeFromTOCMarker+text)
		} else {
			result = append(result, line)
		}
	}

	return []byte(strings.Join(result, "\n"))
}

// generateAnchorID erstellt eine URL-freundliche ID aus einem Überschriftentext.
// Beispiel: "Einführung in DocGen" -> "einführung-in-docgen"
func generateAnchorID(text string) string {
	// Zu Kleinbuchstaben konvertieren
	text = strings.ToLower(text)

	// Umlaute und Sonderzeichen normalisieren
	replacer := strings.NewReplacer(
		"ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss",
		"Ä", "ae", "Ö", "oe", "Ü", "ue",
	)
	text = replacer.Replace(text)

	// Nur alphanumerische Zeichen und Bindestriche behalten
	var result strings.Builder
	lastWasHyphen := false
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			lastWasHyphen = false
		} else if r == ' ' || r == '-' || r == '_' {
			if !lastWasHyphen {
				result.WriteRune('-')
				lastWasHyphen = true
			}
		}
	}

	// Führende und nachfolgende Bindestriche entfernen
	id := strings.Trim(result.String(), "-")

	// Mehrfache Bindestriche durch einzelne ersetzen
	re := regexp.MustCompile(`-+`)
	id = re.ReplaceAllString(id, "-")

	return id
}

// Parse analysiert den Markdown-Inhalt und wandelt ihn in eine Liste von DocBlocks um.
func Parse(content []byte, parentNumbering string) ([]blocks.DocBlock, error) {
	// Vorverarbeitung: !#! Syntax in normale Headings mit Marker umwandeln
	processedContent := preprocessExcludedHeadings(content)

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
		),
	)
	reader := text.NewReader(processedContent)
	doc := md.Parser().Parse(reader)

	var docBlocks []blocks.DocBlock

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			headingText := string(node.Text(processedContent))
			excludeFromTOC := false

			// Prüfen ob der Marker vorhanden ist
			if strings.Contains(headingText, excludeFromTOCMarker) {
				excludeFromTOC = true
				// Marker aus dem Text entfernen
				headingText = strings.ReplaceAll(headingText, excludeFromTOCMarker, "")
			}

			docBlocks = append(docBlocks, blocks.HeadingBlock{
				Level:           node.Level,
				Text:            headingText,
				ParentNumbering: parentNumbering,
				AnchorID:        generateAnchorID(headingText),
				ExcludeFromTOC:  excludeFromTOC,
			})
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph:
			docBlocks = append(docBlocks, blocks.ParagraphBlock{
				Content: parseTextSegments(node, processedContent),
			})
			return ast.WalkSkipChildren, nil
		case *ast.FencedCodeBlock:
			lang := string(node.Language(processedContent))
			codeContent := ""
			for i := 0; i < node.Lines().Len(); i++ {
				line := node.Lines().At(i)
				codeContent += string(line.Value(processedContent))
			}
			if lang == "mermaid" {
				title := ""
				if node.Info != nil {
					info := string(node.Info.Text(processedContent))
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
			listBlock := parseList(node, processedContent)
			docBlocks = append(docBlocks, listBlock)
			return ast.WalkSkipChildren, nil
		case *ast.Image:
			docBlocks = append(docBlocks, blocks.ImageBlock{
				Path: string(node.Destination),
				Alt:  string(node.Text(processedContent)),
			})
			return ast.WalkSkipChildren, nil
		case *ast.ThematicBreak:
			docBlocks = append(docBlocks, blocks.PageBreakBlock{})
			return ast.WalkSkipChildren, nil
		case *extAst.Table:
			table := blocks.TableBlock{}
			for row := node.FirstChild(); row != nil; row = row.NextSibling() {
				if r, ok := row.(*extAst.TableRow); ok {
					var rowData []blocks.TableRow
					for cell := r.FirstChild(); cell != nil; cell = cell.NextSibling() {
						if c, ok := cell.(*extAst.TableCell); ok {
							rowData = append(rowData, blocks.TableRow{
								Content: parseTextSegments(c, processedContent),
								Header:  c.Alignment == extAst.AlignNone,
							})
						}
					}
					table.Rows = append(table.Rows, rowData)
				}
			}
			docBlocks = append(docBlocks, table)
			return ast.WalkSkipChildren, nil
		case *ast.Blockquote:
			// Blockquote-Inhalt rekursiv parsen
			var quoteContent []blocks.DocBlock
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				switch c := child.(type) {
				case *ast.Paragraph:
					quoteContent = append(quoteContent, blocks.ParagraphBlock{
						Content: parseTextSegments(c, processedContent),
					})
				case *ast.List:
					listBlock := parseList(c, processedContent)
					quoteContent = append(quoteContent, listBlock)
				}
			}
			docBlocks = append(docBlocks, blocks.BlockquoteBlock{
				Content: quoteContent,
			})
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Fehler beim Traversieren des Markdown AST: %w", err)
	}

	return docBlocks, nil
}

// parseTextSegments extrahiert Textsegmente mit Formatierungen (fett, kursiv, durchgestrichen, code) aus einem AST-Knoten.
func parseTextSegments(n ast.Node, source []byte) []blocks.TextSegment {
	var segments []blocks.TextSegment
	isBold := false
	isItalic := false
	isStrikethrough := false
	currentLink := ""

	ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if node.Kind() == ast.KindText {
			if entering {
				txt := node.(*ast.Text)
				segments = append(segments, blocks.TextSegment{
					Text:          string(txt.Text(source)),
					Bold:          isBold,
					Italic:        isItalic,
					Strikethrough: isStrikethrough,
					Link:          currentLink,
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
		} else if node.Kind() == extAst.KindStrikethrough {
			if entering {
				isStrikethrough = true
			} else {
				isStrikethrough = false
			}
		} else if node.Kind() == ast.KindLink {
			lnk := node.(*ast.Link)
			if entering {
				currentLink = string(lnk.Destination)
			} else {
				currentLink = ""
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

// parseList parst eine Liste rekursiv mit Unterstützung für verschachtelte Listen.
func parseList(node *ast.List, source []byte) blocks.ListBlock {
	listBlock := blocks.ListBlock{
		Ordered: node.IsOrdered(),
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if li, ok := child.(*ast.ListItem); ok {
			item := blocks.ListItem{
				Content: parseListItemText(li, source),
			}

			// Prüfe auf verschachtelte Listen
			for subChild := li.FirstChild(); subChild != nil; subChild = subChild.NextSibling() {
				if subList, ok := subChild.(*ast.List); ok {
					subListBlock := parseList(subList, source)
					item.SubList = &subListBlock
					break // Nur eine verschachtelte Liste pro Item
				}
			}

			listBlock.Items = append(listBlock.Items, item)
		}
	}

	return listBlock
}

// parseListItemText extrahiert nur den Text eines ListItems (ohne verschachtelte Listen).
func parseListItemText(li *ast.ListItem, source []byte) []blocks.TextSegment {
	var segments []blocks.TextSegment

	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		// Nur Paragraphen und TextBlocks verarbeiten, keine verschachtelten Listen
		if _, ok := child.(*ast.List); ok {
			continue
		}
		if p, ok := child.(*ast.Paragraph); ok {
			segments = append(segments, parseTextSegments(p, source)...)
		} else if child.Kind() == ast.KindTextBlock {
			segments = append(segments, parseTextSegments(child, source)...)
		}
	}

	return segments
}
