package converter

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type MermaidExtension struct{}

func (e *MermaidExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewMermaidRenderer(), 100),
	))
}

type MermaidRenderer struct{}

func NewMermaidRenderer() renderer.NodeRenderer {
	return &MermaidRenderer{}
}

func (r *MermaidRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *MermaidRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	language := string(n.Language(source))
	if language != "mermaid" {
		return ast.WalkContinue, nil
	}

	if entering {
		_, _ = w.WriteString("<div class=\"mermaid\">")
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			_, _ = w.Write(line.Value(source))
		}
	} else {
		_, _ = w.WriteString("</div>")
	}
	return ast.WalkContinue, nil
}
