package site

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjnvr/blog/internal/generator/section"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContent(t *testing.T, files map[string]string) (afero.Fs, string, string) {
	t.Helper()
	fs := afero.NewMemMapFs()
	contentDir := "/content"
	buildDir := "/build"

	for path, content := range files {
		fullPath := filepath.Join(contentDir, path)
		if err := fs.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
		if err := afero.WriteFile(fs, fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return fs, contentDir, buildDir
}

func TestNewGenerator(t *testing.T) {
	g, err := NewGenerator(afero.NewMemMapFs())
	assert.Nil(t, err)
	assert.Equal(t, "./content/markdown", g.contentDir)
	assert.Equal(t, "./target/build", g.buildDir)
	assert.Equal(t, "./content/assets", g.assetsDir)
	assert.Equal(t, "./scripts", g.scriptsDir)
	assert.NotNil(t, g.pageGeneratorFactory)
}

func TestExtractSection(t *testing.T) {
	tests := []struct {
		name       string
		contentDir string
		filePath   string
		want       string
		wantErr    bool
	}{
		{
			name:       "file at root of content dir",
			contentDir: "content/markdown",
			filePath:   "content/markdown/index.md",
			want:       "",
		},
		{
			name:       "file in section subdirectory",
			contentDir: "content/markdown",
			filePath:   "content/markdown/posts/hello.md",
			want:       "posts",
		},
		{
			name:       "file in nested section",
			contentDir: "content/markdown",
			filePath:   "content/markdown/blog/2024/post.md",
			want:       "blog/2024",
		},
		{
			name:       "file in about section",
			contentDir: "content/markdown",
			filePath:   "content/markdown/about/index.md",
			want:       "about",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := section.ExtractSection(tt.contentDir, tt.filePath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakePageGenerator struct {
	generateErr error
	validateErr error
	generated   bool
	validated   bool
}

func (f *fakePageGenerator) Generate() error {
	f.generated = true
	return f.generateErr
}

func (f *fakePageGenerator) Validate() error {
	f.validated = true
	return f.validateErr
}

func TestListSections_ReadsDisplayNameFromIndexMd(t *testing.T) {
	fs, contentDir, _ := setupTestContent(t, map[string]string{
		"index.md":       "# Accueil\n",
		"posts/index.md": "# All Articles\n\nContent.\n",
		"about/index.md": "# About Me\n\nInfo.\n",
	})

	g, _ := NewGenerator(fs)
	g.contentDir = contentDir

	require.NoError(t, g.listSections())
	// home + posts + about
	require.Len(t, g.sections, 3)

	for _, s := range g.sections {
		switch s.DirName {
		case "":
			assert.Equal(t, "Accueil", s.DisplayName)
		case "posts":
			assert.Equal(t, "All Articles", s.DisplayName)
		case "about":
			assert.Equal(t, "About Me", s.DisplayName)
		default:
			assert.Fail(t, "unexpected section", s.DirName)
		}
	}
}

func TestListSections_FallbackToDirectoryName(t *testing.T) {
	fs, contentDir, _ := setupTestContent(t, map[string]string{
		"index.md": "# Accueil\n",
		// posts/ has no index.md
		"posts/hello.md": "# Hello\n",
	})

	g, _ := NewGenerator(fs)
	g.contentDir = contentDir

	require.NoError(t, g.listSections())
	// home + posts
	require.Len(t, g.sections, 2)

	assert.Equal(t, "", g.sections[0].DirName)
	assert.Equal(t, "posts", g.sections[1].DirName)
	assert.Equal(t, "Posts", g.sections[1].DisplayName)
}

func TestGenerate_NonExistentContentDirError(t *testing.T) {
	g, _ := NewGenerator(afero.NewMemMapFs())
	g.contentDir = "/nonexistent/content"
	g.buildDir = "/build"
	g.assetsDir = "/nonexistent/assets"
	g.scriptsDir = "/nonexistent/scripts"

	assert.Error(t, g.Generate())
}

func TestGenerate_NonMdFilesReportError(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":  "# Title",
		"extra.txt": "not markdown",
	})

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen, _ := NewGenerator(fs)
	gen.contentDir = contentDir
	gen.buildDir = buildDir
	gen.assetsOutDir = filepath.Join(buildDir, "assets")
	gen.scriptsOutDir = filepath.Join(buildDir, "scripts")
	gen.assetsDir = "/assets"
	gen.scriptsDir = "/scripts"

	err := gen.Generate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrong extension")
}

func TestGenerate_PageGeneratorError(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	factory := func(markdownPath, htmlOutPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		return &fakePageGenerator{generateErr: fmt.Errorf("page generation failed")}
	}

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen, _ := NewGenerator(fs)
	gen.contentDir = contentDir
	gen.buildDir = buildDir
	gen.assetsOutDir = filepath.Join(buildDir, "assets")
	gen.scriptsOutDir = filepath.Join(buildDir, "scripts")
	gen.assetsDir = "/assets"
	gen.scriptsDir = "/scripts"
	gen.pageGeneratorFactory = factory

	err := gen.Generate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "page generation failed")
}

func TestValidate_ValidationError(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	factory := func(markdownPath, htmlOutPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		return &fakePageGenerator{validateErr: fmt.Errorf("validation failed")}
	}

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen, _ := NewGenerator(fs)
	gen.contentDir = contentDir
	gen.buildDir = buildDir
	gen.assetsOutDir = filepath.Join(buildDir, "assets")
	gen.scriptsOutDir = filepath.Join(buildDir, "scripts")
	gen.assetsDir = "/assets"
	gen.scriptsDir = "/scripts"
	gen.pageGeneratorFactory = factory

	require.NoError(t, gen.Generate())

	err := gen.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidate_NoPages(t *testing.T) {
	g, _ := NewGenerator(afero.NewMemMapFs())
	assert.NoError(t, g.Validate())
}

func TestValidate_AllPass(t *testing.T) {
	g, _ := NewGenerator(afero.NewMemMapFs())
	pg1 := &fakePageGenerator{}
	pg2 := &fakePageGenerator{}
	pg3 := &fakePageGenerator{}
	g.pagesGenerators = []PageGenerator{pg1, pg2, pg3}

	assert.NoError(t, g.Validate())

	for i, pg := range []*fakePageGenerator{pg1, pg2, pg3} {
		assert.Truef(t, pg.validated, "Validate() was not called on page generator %d", i)
	}
}

func TestValidate_WithErrors(t *testing.T) {
	g, _ := NewGenerator(afero.NewMemMapFs())
	g.pagesGenerators = []PageGenerator{
		&fakePageGenerator{},
		&fakePageGenerator{validateErr: fmt.Errorf("broken link")},
		&fakePageGenerator{validateErr: fmt.Errorf("missing image")},
	}
	err := g.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broken link")
	assert.Contains(t, err.Error(), "missing image")
}

func newIntegrationGenerator(t *testing.T, fs afero.Fs, contentDir, buildDir, assetsDir, scriptsDir string) *Generator {
	t.Helper()
	gen, _ := NewGenerator(fs)
	gen.contentDir = contentDir
	gen.buildDir = buildDir
	gen.assetsOutDir = filepath.Join(buildDir, "assets")
	gen.scriptsOutDir = filepath.Join(buildDir, "scripts")
	gen.assetsDir = assetsDir
	gen.scriptsDir = scriptsDir
	return gen
}

func TestIntegration_BasicSiteGeneration(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":        "# Welcome\n\nThis is the homepage.\n\n[Go to posts](posts/index.md)\n",
		"posts/index.md":  "# All Posts\n\n- [First Post](first.md)\n- [Second Post](second.md)\n",
		"posts/first.md":  "# First Post\n\nContent of the first post.\n\n[Back to posts](index.md)\n",
		"posts/second.md": "# Second Post\n\nContent of the second post.\n",
	})

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", "/scripts")
	require.NoError(t, gen.Generate())

	for _, file := range []string{"index.html", "posts/index.html", "posts/first.html", "posts/second.html"} {
		_, err := fs.Stat(filepath.Join(buildDir, file))
		assert.NoError(t, err, "expected file %s to exist", file)
	}
}

func TestIntegration_ScriptsCopied(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Page",
	})

	scriptsDir := "/scripts"
	scriptContent := `(function() { console.log("test"); })();`
	_ = fs.MkdirAll(scriptsDir, 0755)
	_ = afero.WriteFile(fs, filepath.Join(scriptsDir, "test.js"), []byte(scriptContent), 0644)

	_ = fs.MkdirAll("/assets", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", scriptsDir)
	require.NoError(t, gen.Generate())

	copiedScript, err := afero.ReadFile(fs, filepath.Join(buildDir, "scripts/test.js"))
	require.NoError(t, err, "script should be copied")
	assert.Equal(t, scriptContent, string(copiedScript))
}

func TestIntegration_DarkModeScriptCopied(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Page",
	})

	scriptsDir := "/scripts"
	darkModeScript := `(function() {
  const STORAGE_KEY = 'theme';
  function toggleTheme() { /* toggle */ }
  window.toggleTheme = toggleTheme;
})();`
	_ = fs.MkdirAll(scriptsDir, 0755)
	_ = afero.WriteFile(fs, filepath.Join(scriptsDir, "dark-mode.js"), []byte(darkModeScript), 0644)

	_ = fs.MkdirAll("/assets", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", scriptsDir)
	require.NoError(t, gen.Generate())

	copiedScript, err := afero.ReadFile(fs, filepath.Join(buildDir, "scripts/dark-mode.js"))
	require.NoError(t, err, "dark-mode.js should be copied")
	assert.Contains(t, string(copiedScript), "toggleTheme")
}

func TestIntegration_LinkConversion(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":         "# Home\n\n[Go to posts](posts/index.md)\n",
		"posts/index.md":   "# Posts\n\n[Back to home](../index.md)\n[Read article](article.md)\n",
		"posts/article.md": "# Article\n\n[Back to index](index.md)\n",
	})

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", "/scripts")
	require.NoError(t, gen.Generate())

	homeContent, _ := afero.ReadFile(fs, filepath.Join(buildDir, "index.html"))
	assert.Contains(t, string(homeContent), `href="posts/index.html"`)

	postsContent, _ := afero.ReadFile(fs, filepath.Join(buildDir, "posts/index.html"))
	assert.True(t,
		strings.Contains(string(postsContent), `href="../index.html"`) || strings.Contains(string(postsContent), `href="index.html"`),
		"posts/index should link to index.html",
	)
	assert.Contains(t, string(postsContent), `href="article.html"`)
}

func TestIntegration_StaticAssetsCopied(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Page",
	})

	assetsDir := "/assets"
	_ = fs.MkdirAll(filepath.Join(assetsDir, "images"), 0755)
	_ = afero.WriteFile(fs, filepath.Join(assetsDir, "style.css"), []byte("body { color: red; }"), 0644)
	_ = afero.WriteFile(fs, filepath.Join(assetsDir, "images/bg.png"), []byte("fake png data"), 0644)

	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, assetsDir, "/scripts")
	require.NoError(t, gen.Generate())

	cssContent, err := afero.ReadFile(fs, filepath.Join(buildDir, "assets/style.css"))
	assert.NoError(t, err, "CSS file should be copied")
	assert.Equal(t, "body { color: red; }", string(cssContent))

	_, err = fs.Stat(filepath.Join(buildDir, "assets/images/bg.png"))
	assert.NoError(t, err, "PNG file should be copied")
}

func TestIntegration_TitleExtraction(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# My Page Title\n\nSome content.\n",
	})

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", "/scripts")
	require.NoError(t, gen.Generate())

	content, err := afero.ReadFile(fs, filepath.Join(buildDir, "page.html"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "<title>My Page Title</title>")
}

func TestIntegration_InlineAttributesWithoutConfig(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Title {.my-class #my-id}\n\n## Section {data-testid=section-1}\n",
	})

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", "/scripts")
	require.NoError(t, gen.Generate())

	content, err := afero.ReadFile(fs, filepath.Join(buildDir, "page.html"))
	require.NoError(t, err)

	html := string(content)
	assert.Contains(t, html, `class="my-class"`)
	assert.Contains(t, html, `id="my-id"`)
	assert.Contains(t, html, `data-testid="section-1"`)
}

func TestIntegration_NavigationBar(t *testing.T) {
	fs, contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":            "# Accueil\n\nWelcome.\n",
		"posts/index.md":      "# Posts\n\nArticle list.\n",
		"posts/first-post.md": "# First Post\n\nContent.\n",
		"about/index.md":      "# About\n\nAbout page.\n",
	})

	_ = fs.MkdirAll("/assets", 0755)
	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, "/assets", "/scripts")
	require.NoError(t, gen.Generate())

	t.Run("root page has nav with all section links using non-prefixed paths", func(t *testing.T) {
		homeContent, err := afero.ReadFile(fs, filepath.Join(buildDir, "index.html"))
		require.NoError(t, err)
		homeHTML := string(homeContent)

		assert.Contains(t, homeHTML, "<nav")
		assert.Contains(t, homeHTML, `href="index.html"`)
		assert.Contains(t, homeHTML, "Accueil")
		assert.Contains(t, homeHTML, "Posts")
		assert.Contains(t, homeHTML, "About")
		assert.NotContains(t, homeHTML, `href="../`)
	})

	t.Run("section index page has nav with ../ prefixed paths", func(t *testing.T) {
		postsContent, err := afero.ReadFile(fs, filepath.Join(buildDir, "posts/index.html"))
		require.NoError(t, err)
		postsHTML := string(postsContent)

		assert.Contains(t, postsHTML, "<nav")
		assert.Contains(t, postsHTML, `href="../index.html"`)
		assert.Contains(t, postsHTML, `href="../posts/index.html"`)
		assert.Contains(t, postsHTML, `href="../about/index.html"`)
		assert.Contains(t, postsHTML, "Accueil")
	})

	t.Run("non-index page in section also has nav with ../ prefixed paths", func(t *testing.T) {
		postContent, err := afero.ReadFile(fs, filepath.Join(buildDir, "posts/first-post.html"))
		require.NoError(t, err)
		postHTML := string(postContent)

		assert.Contains(t, postHTML, "<nav")
		assert.Contains(t, postHTML, `href="../index.html"`)
		assert.Contains(t, postHTML, `href="../posts/index.html"`)
		assert.Contains(t, postHTML, `href="../about/index.html"`)
	})

	t.Run("about section page has nav with ../ prefixed paths", func(t *testing.T) {
		aboutContent, err := afero.ReadFile(fs, filepath.Join(buildDir, "about/index.html"))
		require.NoError(t, err)
		aboutHTML := string(aboutContent)

		assert.Contains(t, aboutHTML, "<nav")
		assert.Contains(t, aboutHTML, `href="../index.html"`)
		assert.Contains(t, aboutHTML, `href="../posts/index.html"`)
		assert.Contains(t, aboutHTML, "Posts")
	})

	t.Run("all pages have consistent navigation links", func(t *testing.T) {
		for _, page := range []string{"index.html", "posts/index.html", "posts/first-post.html", "about/index.html"} {
			content, err := afero.ReadFile(fs, filepath.Join(buildDir, page))
			require.NoError(t, err, "failed to read %s", page)
			html := string(content)
			assert.Contains(t, html, "Accueil", "%s should contain Accueil nav link", page)
			assert.Contains(t, html, "Posts", "%s should contain Posts nav link", page)
			assert.Contains(t, html, "About", "%s should contain About nav link", page)
		}
	})
}

func TestIntegration_AssetPathConversion(t *testing.T) {
	// Use a shared root so that relative paths from markdown to assets resolve correctly
	contentDir := "/root/content/markdown"
	assetsDir := "/root/content/assets"
	buildDir := "/root/target/build"

	fs := afero.NewMemMapFs()

	for path, content := range map[string]string{
		"index.md":             "# Home\n\nWelcome.\n",
		"posts/index.md":       "# Posts\n\nArticle list.\n",
		"posts/second-post.md": "# Second Post\n\nContent.\n\n![Image](../../assets/images/photo.png)\n",
	} {
		fullPath := filepath.Join(contentDir, path)
		_ = fs.MkdirAll(filepath.Dir(fullPath), 0755)
		_ = afero.WriteFile(fs, fullPath, []byte(content), 0644)
	}

	_ = fs.MkdirAll(filepath.Join(assetsDir, "images"), 0755)
	_ = afero.WriteFile(fs, filepath.Join(assetsDir, "images/photo.png"), []byte("fake png data"), 0644)

	_ = fs.MkdirAll("/scripts", 0755)

	gen := newIntegrationGenerator(t, fs, contentDir, buildDir, assetsDir, "/scripts")
	require.NoError(t, gen.Generate())

	postContent, err := afero.ReadFile(fs, filepath.Join(buildDir, "posts/second-post.html"))
	require.NoError(t, err)

	// The image is at /root/target/build/assets/images/photo.png
	// The HTML is at /root/target/build/posts/second-post.html
	// So the correct relative path is ../assets/images/photo.png
	assert.Contains(t, string(postContent), `src="../assets/images/photo.png"`)
}
