package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tjnvr/blog/internal/generator/section"
)

// Generator test helpers
func (g *Generator) withPageGeneratorFactory(factory pageGeneratorFactory) *Generator {
	g.pageGeneratorFactory = factory
	return g
}

func (g *Generator) withContentDir(dir string) *Generator {
	g.contentDir = dir
	return g
}

func (g *Generator) withBuildDir(dir string) *Generator {
	g.buildDir = dir
	return g
}

func (g *Generator) withAssetsDir(dir string) *Generator {
	g.assetsDir = dir
	return g
}

func (g *Generator) withScriptsDir(dir string) *Generator {
	g.scriptsDir = dir
	return g
}

func setupTestContent(t *testing.T, files map[string]string) (contentDir, buildDir string) {
	t.Helper()
	contentDir = t.TempDir()
	buildDir = t.TempDir()

	for path, content := range files {
		fullPath := filepath.Join(contentDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return contentDir, buildDir
}

func createTestGenerator(contentDir, buildDir string) *Generator {
	g, _ := NewGenerator()
	g.withContentDir(contentDir).
		withBuildDir(buildDir)
	return g
}

func TestNewGenerator(t *testing.T) {
	g, err := NewGenerator()
	assert.Nil(t, err)
	if g.contentDir != "./content/markdown" {
		t.Errorf("contentDir = %q, want %q", g.contentDir, "./content/markdown")
	}
	if g.buildDir != "./target/build" {
		t.Errorf("buildDir = %q, want %q", g.buildDir, "./target/build")
	}
	if g.assetsDir != "./content/assets" {
		t.Errorf("assetsDir = %q, want %q", g.assetsDir, "./content/assets")
	}
	if g.scriptsDir != "./scripts" {
		t.Errorf("scriptsDir = %q, want %q", g.scriptsDir, "./scripts")
	}
	if g.pageGeneratorFactory == nil {
		t.Error("pageGeneratorFactory should not be nil")
	}
}

func TestWithBuilders(t *testing.T) {
	g, _ := NewGenerator()
	g.withContentDir("/content").
		withBuildDir("/build").
		withAssetsDir("/assets").
		withScriptsDir("/scripts")

	if g.contentDir != "/content" {
		t.Errorf("contentDir = %q, want %q", g.contentDir, "/content")
	}
	if g.buildDir != "/build" {
		t.Errorf("buildDir = %q, want %q", g.buildDir, "/build")
	}
	if g.assetsDir != "/assets" {
		t.Errorf("assetsDir = %q, want %q", g.assetsDir, "/assets")
	}
	if g.scriptsDir != "/scripts" {
		t.Errorf("scriptsDir = %q, want %q", g.scriptsDir, "/scripts")
	}
}

func TestWithPageGeneratorFactory(t *testing.T) {
	called := false
	factory := func(markdownPath, htmlOutPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		called = true
		return &fakePageGenerator{}
	}

	g, _ := NewGenerator()
	g.withPageGeneratorFactory(factory)
	g.pageGeneratorFactory("test.md", "test.html", "/build", "", newPathResolver{}, newPathResolver{}, nil, false)

	if !called {
		t.Error("custom factory should have been called")
	}
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
			got, err := extractSection(tt.contentDir, tt.filePath)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("extractSection(%q, %q) = %q, want %q", tt.contentDir, tt.filePath, got, tt.want)
			}
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
	contentDir, _ := setupTestContent(t, map[string]string{
		"index.md":       "# Accueil\n",
		"posts/index.md": "# All Articles\n\nContent.\n",
		"about/index.md": "# About Me\n\nInfo.\n",
	})

	g, _ := NewGenerator()
	g.withContentDir(contentDir)

	if err := g.listSections(); err != nil {
		t.Fatalf("listSections() error = %v", err)
	}

	// home + posts + about
	if len(g.sections) != 3 {
		t.Fatalf("expected 3 sections (home + 2), got %d", len(g.sections))
	}

	for _, s := range g.sections {
		switch s.DirName {
		case "":
			if s.DisplayName != "Accueil" {
				t.Errorf("home section DisplayName = %q, want %q", s.DisplayName, "Accueil")
			}
		case "posts":
			if s.DisplayName != "All Articles" {
				t.Errorf("posts section DisplayName = %q, want %q", s.DisplayName, "All Articles")
			}
		case "about":
			if s.DisplayName != "About Me" {
				t.Errorf("about section DisplayName = %q, want %q", s.DisplayName, "About Me")
			}
		default:
			t.Errorf("unexpected section %q", s.DirName)
		}
	}
}

func TestListSections_FallbackToDirectoryName(t *testing.T) {
	contentDir, _ := setupTestContent(t, map[string]string{
		"index.md": "# Accueil\n",
		// posts/ has no index.md
		"posts/hello.md": "# Hello\n",
	})

	g, _ := NewGenerator()
	g.withContentDir(contentDir)

	if err := g.listSections(); err != nil {
		t.Fatalf("listSections() error = %v", err)
	}

	// home + posts
	if len(g.sections) != 2 {
		t.Fatalf("expected 2 sections (home + posts), got %d", len(g.sections))
	}

	// First section is home
	if g.sections[0].DirName != "" {
		t.Errorf("first section DirName = %q, want empty string (home)", g.sections[0].DirName)
	}

	// Second section is posts with fallback display name
	if g.sections[1].DirName != "posts" {
		t.Errorf("DirName = %q, want %q", g.sections[1].DirName, "posts")
	}
	if g.sections[1].DisplayName != "Posts" {
		t.Errorf("DisplayName = %q, want %q (fallback to capitalized dir name)", g.sections[1].DisplayName, "Posts")
	}
}

func TestGenerate_NonExistentContentDirError(t *testing.T) {
	buildDir := t.TempDir()
	gen := createTestGenerator("/nonexistent/content", buildDir).
		withAssetsDir("/nonexistent/assets").
		withScriptsDir("/nonexistent/scripts")

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error for non-existent content dir")
	}
}

func TestGenerate_NonMdFilesReportError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":  "# Title",
		"extra.txt": "not markdown",
	})

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))

	// Create empty assets/scripts dirs so copyDir doesn't fail
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error for non-md file in content dir")
	}
	if !strings.Contains(err.Error(), "wrong extension") {
		t.Errorf("error should mention wrong extension, got %q", err.Error())
	}
}

func TestGenerate_PageGeneratorError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	factory := func(markdownPath, htmlOutPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		return &fakePageGenerator{generateErr: fmt.Errorf("page generation failed")}
	}

	gen := createTestGenerator(contentDir, buildDir).
		withPageGeneratorFactory(factory).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err == nil {
		t.Fatal("expected error when page generator fails")
	}
	if !strings.Contains(err.Error(), "page generation failed") {
		t.Errorf("error should contain page generator message, got %q", err.Error())
	}
}

func TestValidate_ValidationError(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Title",
	})

	factory := func(markdownPath, htmlOutPath, buildDir, pageSection string, assetsPathTranslater, linksPathTranslater newPathResolver, sections []section.Section, skipURLValidation bool) PageGenerator {
		return &fakePageGenerator{validateErr: fmt.Errorf("validation failed")}
	}

	gen := createTestGenerator(contentDir, buildDir).
		withPageGeneratorFactory(factory).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	if err := gen.Generate(); err != nil {
		t.Fatalf("unexpected error during generation: %v", err)
	}

	err := gen.Validate()
	if err == nil {
		t.Fatal("expected error when validation fails")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error should contain validation message, got %q", err.Error())
	}
}

func TestValidate_NoPages(t *testing.T) {
	g, _ := NewGenerator()
	err := g.Validate()
	if err != nil {
		t.Errorf("Validate() with no pages should return nil, got %v", err)
	}
}

func TestValidate_AllPass(t *testing.T) {
	g, _ := NewGenerator()
	pg1 := &fakePageGenerator{}
	pg2 := &fakePageGenerator{}
	pg3 := &fakePageGenerator{}
	g.pagesGenerators = []PageGenerator{pg1, pg2, pg3}

	err := g.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	for i, pg := range []*fakePageGenerator{pg1, pg2, pg3} {
		if !pg.validated {
			t.Errorf("Validate() was not called on page generator %d", i)
		}
	}
}

func TestValidate_WithErrors(t *testing.T) {
	g, _ := NewGenerator()
	g.pagesGenerators = []PageGenerator{
		&fakePageGenerator{},
		&fakePageGenerator{validateErr: fmt.Errorf("broken link")},
		&fakePageGenerator{validateErr: fmt.Errorf("missing image")},
	}
	err := g.Validate()
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "broken link") {
		t.Errorf("error should contain 'broken link', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "missing image") {
		t.Errorf("error should contain 'missing image', got %q", err.Error())
	}
}

func TestIntegration_BasicSiteGeneration(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":        "# Welcome\n\nThis is the homepage.\n\n[Go to posts](posts/index.md)\n",
		"posts/index.md":  "# All Posts\n\n- [First Post](first.md)\n- [Second Post](second.md)\n",
		"posts/first.md":  "# First Post\n\nContent of the first post.\n\n[Back to posts](index.md)\n",
		"posts/second.md": "# Second Post\n\nContent of the second post.\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expectedFiles := []string{
		"index.html",
		"posts/index.html",
		"posts/first.html",
		"posts/second.html",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(buildDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestIntegration_ScriptsCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Page",
	})

	scriptsDir := t.TempDir()
	scriptContent := `(function() { console.log("test"); })();`
	_ = os.WriteFile(filepath.Join(scriptsDir, "test.js"), []byte(scriptContent), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		withScriptsDir(scriptsDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets"))
	_ = os.MkdirAll(gen.assetsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	copiedScript, err := os.ReadFile(filepath.Join(buildDir, "scripts/test.js"))
	if err != nil {
		t.Fatalf("script should be copied: %v", err)
	}
	if string(copiedScript) != scriptContent {
		t.Errorf("script content mismatch, got: %s", copiedScript)
	}
}

func TestIntegration_DarkModeScriptCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Page",
	})

	scriptsDir := t.TempDir()
	darkModeScript := `(function() {
  const STORAGE_KEY = 'theme';
  function toggleTheme() { /* toggle */ }
  window.toggleTheme = toggleTheme;
})();`
	_ = os.WriteFile(filepath.Join(scriptsDir, "dark-mode.js"), []byte(darkModeScript), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		withScriptsDir(scriptsDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets"))
	_ = os.MkdirAll(gen.assetsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	copiedScript, err := os.ReadFile(filepath.Join(buildDir, "scripts/dark-mode.js"))
	if err != nil {
		t.Fatalf("dark-mode.js should be copied: %v", err)
	}
	if !strings.Contains(string(copiedScript), "toggleTheme") {
		t.Errorf("dark-mode.js should contain toggleTheme function")
	}
}

func TestIntegration_LinkConversion(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":         "# Home\n\n[Go to posts](posts/index.md)\n",
		"posts/index.md":   "# Posts\n\n[Back to home](../index.md)\n[Read article](article.md)\n",
		"posts/article.md": "# Article\n\n[Back to index](index.md)\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	homeContent, _ := os.ReadFile(filepath.Join(buildDir, "index.html"))
	if !strings.Contains(string(homeContent), `href="posts/index.html"`) {
		t.Errorf("index should link to posts/index.html, got:\n%s", homeContent)
	}

	postsContent, _ := os.ReadFile(filepath.Join(buildDir, "posts/index.html"))
	if !strings.Contains(string(postsContent), `href="../index.html"`) && !strings.Contains(string(postsContent), `href="index.html"`) {
		t.Errorf("posts/index should link to index.html, got:\n%s", postsContent)
	}
	if !strings.Contains(string(postsContent), `href="article.html"`) {
		t.Errorf("posts/index should link to article.html, got:\n%s", postsContent)
	}
}

func TestIntegration_StaticAssetsCopied(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Page",
	})

	assetsDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(assetsDir, "images"), 0755)
	_ = os.WriteFile(filepath.Join(assetsDir, "style.css"), []byte("body { color: red; }"), 0644)
	_ = os.WriteFile(filepath.Join(assetsDir, "images/bg.png"), []byte("fake png data"), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(assetsDir).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	cssContent, err := os.ReadFile(filepath.Join(buildDir, "assets/style.css"))
	if err != nil {
		t.Errorf("CSS file should be copied: %v", err)
	} else if string(cssContent) != "body { color: red; }" {
		t.Errorf("CSS content mismatch")
	}

	_, err = os.ReadFile(filepath.Join(buildDir, "assets/images/bg.png"))
	if err != nil {
		t.Errorf("PNG file should be copied: %v", err)
	}
}

func TestIntegration_TitleExtraction(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# My Page Title\n\nSome content.\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(content), "<title>My Page Title</title>") {
		t.Errorf("title tag should contain page title, got:\n%s", content)
	}
}

func TestIntegration_InlineAttributesWithoutConfig(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md": "# Home\n",
		"page.md":  "# Title {.my-class #my-id}\n\n## Section {data-testid=section-1}\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(buildDir, "page.html"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	html := string(content)
	if !strings.Contains(html, `class="my-class"`) {
		t.Errorf("should have inline class")
	}
	if !strings.Contains(html, `id="my-id"`) {
		t.Errorf("should have inline id")
	}
	if !strings.Contains(html, `data-testid="section-1"`) {
		t.Errorf("should have data attribute")
	}
}

func TestIntegration_NavigationBar(t *testing.T) {
	contentDir, buildDir := setupTestContent(t, map[string]string{
		"index.md":            "# Accueil\n\nWelcome.\n",
		"posts/index.md":      "# Posts\n\nArticle list.\n",
		"posts/first-post.md": "# First Post\n\nContent.\n",
		"about/index.md":      "# About\n\nAbout page.\n",
	})

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(filepath.Join(t.TempDir(), "empty-assets")).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.assetsDir, 0755)
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	t.Run("root page has nav with all section links using non-prefixed paths", func(t *testing.T) {
		homeContent, err := os.ReadFile(filepath.Join(buildDir, "index.html"))
		if err != nil {
			t.Fatalf("failed to read index.html: %v", err)
		}
		homeHTML := string(homeContent)

		if !strings.Contains(homeHTML, "<nav") {
			t.Error("should contain nav element")
		}
		if !strings.Contains(homeHTML, `href="index.html"`) {
			t.Errorf("should have home link with non-prefixed path, got:\n%s", homeHTML)
		}
		if !strings.Contains(homeHTML, "Accueil") {
			t.Error("should contain Accueil display name")
		}
		if !strings.Contains(homeHTML, "Posts") {
			t.Error("should contain Posts display name")
		}
		if !strings.Contains(homeHTML, "About") {
			t.Error("should contain About display name")
		}
		// Root page should NOT have ../ prefixed links
		if strings.Contains(homeHTML, `href="../`) {
			t.Errorf("root page should not have ../ prefixed nav links, got:\n%s", homeHTML)
		}
	})

	t.Run("section index page has nav with ../ prefixed paths", func(t *testing.T) {
		postsContent, err := os.ReadFile(filepath.Join(buildDir, "posts/index.html"))
		if err != nil {
			t.Fatalf("failed to read posts/index.html: %v", err)
		}
		postsHTML := string(postsContent)

		if !strings.Contains(postsHTML, "<nav") {
			t.Error("should contain nav element")
		}
		if !strings.Contains(postsHTML, `href="../index.html"`) {
			t.Errorf("should have ../ prefixed home link, got:\n%s", postsHTML)
		}
		if !strings.Contains(postsHTML, `href="../posts/index.html"`) {
			t.Errorf("should have ../ prefixed posts link, got:\n%s", postsHTML)
		}
		if !strings.Contains(postsHTML, `href="../about/index.html"`) {
			t.Errorf("should have ../ prefixed about link, got:\n%s", postsHTML)
		}
		if !strings.Contains(postsHTML, "Accueil") {
			t.Error("should contain Accueil display name")
		}
	})

	t.Run("non-index page in section also has nav with ../ prefixed paths", func(t *testing.T) {
		postContent, err := os.ReadFile(filepath.Join(buildDir, "posts/first-post.html"))
		if err != nil {
			t.Fatalf("failed to read posts/first-post.html: %v", err)
		}
		postHTML := string(postContent)

		if !strings.Contains(postHTML, "<nav") {
			t.Error("should contain nav element")
		}
		if !strings.Contains(postHTML, `href="../index.html"`) {
			t.Errorf("should have ../ prefixed home link, got:\n%s", postHTML)
		}
		if !strings.Contains(postHTML, `href="../posts/index.html"`) {
			t.Errorf("should have ../ prefixed posts link, got:\n%s", postHTML)
		}
		if !strings.Contains(postHTML, `href="../about/index.html"`) {
			t.Errorf("should have ../ prefixed about link, got:\n%s", postHTML)
		}
	})

	t.Run("about section page has nav with ../ prefixed paths", func(t *testing.T) {
		aboutContent, err := os.ReadFile(filepath.Join(buildDir, "about/index.html"))
		if err != nil {
			t.Fatalf("failed to read about/index.html: %v", err)
		}
		aboutHTML := string(aboutContent)

		if !strings.Contains(aboutHTML, "<nav") {
			t.Error("should contain nav element")
		}
		if !strings.Contains(aboutHTML, `href="../index.html"`) {
			t.Errorf("should have ../ prefixed home link, got:\n%s", aboutHTML)
		}
		if !strings.Contains(aboutHTML, `href="../posts/index.html"`) {
			t.Errorf("should have ../ prefixed posts link, got:\n%s", aboutHTML)
		}
		if !strings.Contains(aboutHTML, "Posts") {
			t.Error("should contain Posts display name")
		}
	})

	t.Run("all pages have consistent navigation links", func(t *testing.T) {
		pages := []string{
			"index.html",
			"posts/index.html",
			"posts/first-post.html",
			"about/index.html",
		}
		for _, page := range pages {
			content, err := os.ReadFile(filepath.Join(buildDir, page))
			if err != nil {
				t.Fatalf("failed to read %s: %v", page, err)
			}
			html := string(content)
			if !strings.Contains(html, "Accueil") {
				t.Errorf("%s should contain Accueil nav link", page)
			}
			if !strings.Contains(html, "Posts") {
				t.Errorf("%s should contain Posts nav link", page)
			}
			if !strings.Contains(html, "About") {
				t.Errorf("%s should contain About nav link", page)
			}
		}
	})
}

func TestIntegration_AssetPathConversion(t *testing.T) {
	// Use a shared root so that relative paths from markdown to assets resolve correctly
	rootDir := t.TempDir()
	contentDir := filepath.Join(rootDir, "content", "markdown")
	assetsDir := filepath.Join(rootDir, "content", "assets")
	buildDir := filepath.Join(rootDir, "target", "build")

	// Create content files
	for path, content := range map[string]string{
		"index.md":             "# Home\n\nWelcome.\n",
		"posts/index.md":       "# Posts\n\nArticle list.\n",
		"posts/second-post.md": "# Second Post\n\nContent.\n\n![Image](../../assets/images/photo.png)\n",
	} {
		fullPath := filepath.Join(contentDir, path)
		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
		_ = os.WriteFile(fullPath, []byte(content), 0644)
	}

	// Create asset files
	_ = os.MkdirAll(filepath.Join(assetsDir, "images"), 0755)
	_ = os.WriteFile(filepath.Join(assetsDir, "images/photo.png"), []byte("fake png data"), 0644)

	gen := createTestGenerator(contentDir, buildDir).
		withAssetsDir(assetsDir).
		withScriptsDir(filepath.Join(t.TempDir(), "empty-scripts"))
	_ = os.MkdirAll(gen.scriptsDir, 0755)

	err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	postContent, err := os.ReadFile(filepath.Join(buildDir, "posts/second-post.html"))
	if err != nil {
		t.Fatalf("failed to read posts/second-post.html: %v", err)
	}
	postHTML := string(postContent)

	// The image is at target/build/assets/images/photo.png
	// The HTML is at target/build/posts/second-post.html
	// So the correct relative path is ../assets/images/photo.png
	if !strings.Contains(postHTML, `src="../assets/images/photo.png"`) {
		t.Errorf("expected src=\"../assets/images/photo.png\" in output, got:\n%s", postHTML)
	}
}
