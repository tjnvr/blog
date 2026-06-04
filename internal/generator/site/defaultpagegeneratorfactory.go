package site

import (
	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/page"
)

type defaultPageGeneratorFactory struct{}

func (f defaultPageGeneratorFactory) Create(
	fs afero.Fs,
	markdownSubstituer page.MarkdownSubstituer,
	markdownToHTMLConverter page.MarkdownConverter,
	HTMLSubstituer page.HTMLSubstituer,
	HTMLValidator page.PageValidator,
) PageGenerator {
	return page.NewGenerator(
		fs,
		markdownSubstituer,
		markdownToHTMLConverter,
		HTMLSubstituer,
		HTMLValidator,
	)
}
