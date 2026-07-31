package site

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tjnvr/blog/internal/abspath"
	"github.com/tjnvr/blog/internal/backbone/section"
	"github.com/tjnvr/blog/internal/hrefpath"
	mfs "github.com/tjnvr/blog/internal/io/fs"
	"github.com/tjnvr/blog/internal/relpath"
)

type fakePageGenerator struct{}

func (fakePageGenerator) Generate() error { return nil }
func (fakePageGenerator) Validate() error { return nil }

// fakeRelResolver is a relpath.Resolver double. Its resolved value is a
// fixed id rather than a real computation, so tests can assert which
// resolver *instance* reached a given call site.
type fakeRelResolver struct{ id string }

func (f fakeRelResolver) Resolve(string, string) (string, error) { return f.id, nil }

// fakeHrefResolver is an hrefpath.Resolver double, same purpose as above.
type fakeHrefResolver struct{ id string }

func (f fakeHrefResolver) Resolve(string) (string, error) { return f.id, nil }

// TestGeneratePages_RoutesResolversToTheirCorrectRole guards against
// generatePages ever routing the file-writing resolver and the href
// resolver to the wrong place: the file resolver must drive HTMLPath (the
// on-disk destination), and the href resolver — not the file resolver —
// must be the one threaded into the page factory for link generation.
func TestGeneratePages_RoutesResolversToTheirCorrectRole(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "content/index.md", []byte("# Home"), 0644))

	assetsFake := fakeRelResolver{id: "assets-resolver"}
	linksFake := fakeRelResolver{id: "links-resolver"}
	hrefFake := fakeHrefResolver{id: "href-resolver"}

	var capturedHTMLPath string
	var capturedHrefResolver hrefpath.Resolver
	var capturedAssetResolver relpath.Resolver

	g := &Generator{
		Config:          Config{ContentDir: "content"},
		pagesFinder:     mfs.NewFilesFinder(fs, mfs.WithExtension(".md")),
		pagesGenerators: make([]PageGenerator, 0),
		fs:              fs,
		pageGeneratorFactory: func(_ afero.Fs, _, HTMLPath string, _ abspath.ResolverFactory, _ section.Resolver, hrefPathsResolver hrefpath.Resolver, assetPathsResolver relpath.Resolver, _ bool) PageGenerator {
			capturedHTMLPath = HTMLPath
			capturedHrefResolver = hrefPathsResolver
			capturedAssetResolver = assetPathsResolver
			return fakePageGenerator{}
		},
	}

	// test
	err := g.generatePages(assetsFake, linksFake, hrefFake)

	// expect
	require.NoError(t, err)
	assert.Equal(t, "links-resolver", capturedHTMLPath, "HTMLPath must come from the file-writing resolver")
	assert.Equal(t, hrefFake, capturedHrefResolver, "the factory's href slot must receive the href resolver, not the file resolver")
	assert.Equal(t, assetsFake, capturedAssetResolver)
}
