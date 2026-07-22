package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// ImageAttributeTransformer attaches a trailing {...} attribute list
// (e.g. {.small-centered}) immediately following an image to that image's
// node, so the attribute renders on the <img> tag. The brace text itself is
// removed from the surrounding paragraph.
type ImageAttributeTransformer struct{}

// Transform implements parser.ASTTransformer.
func (t *ImageAttributeTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		for c := n.FirstChild(); c != nil; {
			next := c.NextSibling()
			if img, ok := c.(*ast.Image); ok {
				attachTrailingAttributes(img, reader)
			}
			walk(c)
			c = next
		}
	}
	walk(doc)
}

// attachTrailingAttributes parses a {...} attribute list from the text node
// immediately following img, if any, attaches the parsed attributes to img,
// and trims or removes the consumed portion of that text node.
func attachTrailingAttributes(img *ast.Image, reader text.Reader) {
	txt, ok := img.NextSibling().(*ast.Text)
	if !ok {
		return
	}

	seg := txt.Segment
	reader.SetPosition(0, text.NewSegment(seg.Start, seg.Stop))
	attrs, ok := parser.ParseAttributes(reader)
	if !ok {
		return
	}

	for _, a := range attrs {
		img.SetAttribute(a.Name, a.Value)
	}

	_, newSeg := reader.Position()
	if newSeg.Start >= seg.Stop {
		txt.Parent().RemoveChild(txt.Parent(), txt)
	} else {
		txt.Segment = text.NewSegment(newSeg.Start, seg.Stop)
	}
}
