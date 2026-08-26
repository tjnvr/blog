package articlelisting

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/tjnvr/blog/internal/hrefpath"
	"github.com/tjnvr/blog/internal/relpath"
)

func TestSubstituter_Placeholder_ShouldReturnListChildPagesPlaceholder(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	assetPathsResolver := relpath.NewResolver("content/assets", "assets")
	pagePathsResolver := hrefpath.NewResolver("content/markdown")

	// setup
	s := NewSubstituer(fs, "public/posts/index.html", "content/markdown/posts/index.md", assetPathsResolver, pagePathsResolver)

	// test
	got := s.Placeholder()

	// expect
	assert.Equal(t, `<p>{{list-child-pages}}</p>`, got)
}

func TestSubstituter_Resolve_ShouldRenderSiblingArticles(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "content/markdown/posts/index.md",
		[]byte("# Articles\n\n{{list-child-pages}}"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	if err := afero.WriteFile(fs, "content/markdown/posts/fixed-article.md",
		[]byte("<!-- creation-date: 2025-12-21 -->\n<!-- description: A fixed article summary. -->\n# CloudFront et S3"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	if err := afero.WriteFile(fs, "content/markdown/posts/other-article.md",
		[]byte("<!-- creation-date: 2025-11-02 -->\n<!-- description: Another article summary. -->\n<!-- image: ../../assets/images/cover.png -->\n# Ma stack de veille"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	assetPathsResolver := relpath.NewResolver("content/assets", "assets")
	pagePathsResolver := hrefpath.NewResolver("content/markdown")

	// setup
	s := NewSubstituer(fs, "public/posts/index.html", "content/markdown/posts/index.md", assetPathsResolver, pagePathsResolver)

	// test
	got, err := s.Resolve("")

	// expect
	assert.NoError(t, err)
	assert.Contains(t, got, "CloudFront et S3")
	assert.Contains(t, got, "21 décembre 2025")
	assert.Contains(t, got, "A fixed article summary.")
	assert.Contains(t, got, `href="/posts/fixed-article"`)
	assert.Contains(t, got, "Ma stack de veille")
	assert.Contains(t, got, "2 novembre 2025")
	assert.Contains(t, got, "Another article summary.")
	assert.Contains(t, got, `href="/posts/other-article"`)
	assert.Contains(t, got, `src="../../assets/images/cover.png"`)
}

func TestSubstituter_Resolve_ShouldReturnEmptyWhenNoSiblingArticles(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "content/markdown/posts/index.md",
		[]byte("# Articles\n\n{{list-child-pages}}"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	assetPathsResolver := relpath.NewResolver("content/assets", "assets")
	pagePathsResolver := hrefpath.NewResolver("content/markdown")

	// setup
	s := NewSubstituer(fs, "public/posts/index.html", "content/markdown/posts/index.md", assetPathsResolver, pagePathsResolver)

	// test
	got, err := s.Resolve("")

	// expect
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestSubstituter_Resolve_ShouldOmitThumbnailWhenArticleHasNoImage(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "content/markdown/posts/index.md",
		[]byte("# Articles\n\n{{list-child-pages}}"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	if err := afero.WriteFile(fs, "content/markdown/posts/fixed-article.md",
		[]byte("<!-- creation-date: 2026-01-03 -->\n<!-- description: No thumbnail here. -->\n# Go et afero pour les tests"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	assetPathsResolver := relpath.NewResolver("content/assets", "assets")
	pagePathsResolver := hrefpath.NewResolver("content/markdown")

	// setup
	s := NewSubstituer(fs, "public/posts/index.html", "content/markdown/posts/index.md", assetPathsResolver, pagePathsResolver)

	// test
	got, err := s.Resolve("")

	// expect
	assert.NoError(t, err)
	assert.Contains(t, got, "Go et afero pour les tests")
	assert.NotContains(t, got, "<img")
}

func TestSubstituter_Resolve_ShouldReturnErrorOnMalformedCreationDate(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "content/markdown/posts/index.md",
		[]byte("# Articles\n\n{{list-child-pages}}"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	if err := afero.WriteFile(fs, "content/markdown/posts/bad-article.md",
		[]byte("<!-- creation-date: not-a-date -->\n<!-- description: Broken date. -->\n# Broken"), 0644); err != nil {
		t.Fatalf("WriteFile err: %v", err)
	}
	assetPathsResolver := relpath.NewResolver("content/assets", "assets")
	pagePathsResolver := hrefpath.NewResolver("content/markdown")

	// setup
	s := NewSubstituer(fs, "public/posts/index.html", "content/markdown/posts/index.md", assetPathsResolver, pagePathsResolver)

	// test
	_, err := s.Resolve("")

	// expect
	assert.Error(t, err)
}
