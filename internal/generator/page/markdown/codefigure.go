package markdown

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindCodeFigure is the node kind of CodeFigure.
var KindCodeFigure = ast.NewNodeKind("CodeFigure")

// CodeFigure wraps a fenced code block that declares a title.
type CodeFigure struct {
	ast.BaseBlock
	Title []byte
}

// Kind implements ast.Node.
func (n *CodeFigure) Kind() ast.NodeKind {
	return KindCodeFigure
}

// Dump implements ast.Node.
func (n *CodeFigure) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Title": string(n.Title)}, nil)
}

// CodeFigureTransformer wraps fenced code blocks whose info string carries a
// {title="..."} attribute list in a CodeFigure, and removes that attribute
// list from the info string so the language is left untouched.
type CodeFigureTransformer struct{}

// Transform implements parser.ASTTransformer.
func (t *CodeFigureTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	var blocks []*ast.FencedCodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if fc, ok := n.(*ast.FencedCodeBlock); ok && entering {
			blocks = append(blocks, fc)
		}
		return ast.WalkContinue, nil
	})

	for _, fc := range blocks {
		if title := extractTitle(fc, reader); title != nil {
			fig := &CodeFigure{Title: title}
			parent := fc.Parent()
			parent.ReplaceChild(parent, fc, fig)
			fig.AppendChild(fig, fc)
		}
	}
}

// extractTitle returns the title attribute of fc's info string, if any, and
// trims the attribute list from the info string.
func extractTitle(fc *ast.FencedCodeBlock, reader text.Reader) []byte {
	if fc.Info == nil {
		return nil
	}

	seg := fc.Info.Segment
	info := seg.Value(reader.Source())
	i := bytes.IndexByte(info, '{')
	if i < 0 {
		return nil
	}

	reader.SetPosition(0, text.NewSegment(seg.Start+i, seg.Stop))
	attrs, ok := parser.ParseAttributes(reader)
	if !ok {
		return nil
	}
	title, ok := attrs.Find([]byte("title"))
	if !ok {
		return nil
	}
	value, ok := title.([]byte)
	if !ok {
		return nil
	}

	lang := util.TrimRightSpace(info[:i])
	if len(lang) == 0 {
		fc.Info = nil
	} else {
		fc.Info = ast.NewTextSegment(text.NewSegment(seg.Start, seg.Start+len(lang)))
	}
	return value
}

// CodeFigureRenderer renders a CodeFigure as a <figure> holding the code
// block followed by a <figcaption>.
type CodeFigureRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *CodeFigureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCodeFigure, r.renderCodeFigure)
}

func (r *CodeFigureRenderer) renderCodeFigure(
	w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<figure>\n")
	} else {
		_, _ = w.WriteString("<figcaption>")
		_, _ = w.Write(util.EscapeHTML(node.(*CodeFigure).Title))
		_, _ = w.WriteString("</figcaption>\n</figure>\n")
	}
	return ast.WalkContinue, nil
}
