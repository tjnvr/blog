package page

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tjnvr/blog/internal/generator/page/filesystem"
	htmlsubstitution "github.com/tjnvr/blog/internal/generator/page/html/substitution"
	"github.com/tjnvr/blog/internal/generator/page/html/validation"
	mdsubstitution "github.com/tjnvr/blog/internal/generator/page/markdown/substitution"
	"github.com/tjnvr/blog/internal/generator/section"
)

func newTestGenerator(t *testing.T, markdownPath, htmlOutputPath, buildDir, sectionName string, fs *filesystem.MemoryFileSystem) *Generator {
	t.Helper()
	mdSubs := mdsubstitution.NewRegistry(markdownPath)
	subs := htmlsubstitution.NewRegistry(htmlOutputPath, markdownPath, nil, nil, nil, sectionName)
	vals := validation.NewRegistry(nil, false)
	return NewGenerator(markdownPath, htmlOutputPath, buildDir, sectionName, fs, mdSubs, subs, vals)
}

func TestNewGenerator(t *testing.T) {
	t.Run("sets output path without section", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/index.md", "/build/index.html", "/build", "", fs)

		if g.destinationHTMLPath != "/build/index.html" {
			t.Errorf("htmlOutputPath = %q, want %q", g.destinationHTMLPath, "/build/index.html")
		}
	})

	t.Run("sets output path with section", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/posts/hello.md", "/build/posts/hello.html", "/build", "posts", fs)

		if g.destinationHTMLPath != "/build/posts/hello.html" {
			t.Errorf("htmlOutputPath = %q, want %q", g.destinationHTMLPath, "/build/posts/hello.html")
		}
	})

	t.Run("stores all fields", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/page.md", "/out/page.html", "/out", "blog", fs)

		if g.sourceMDPath != "/content/page.md" {
			t.Errorf("markdownPath = %q, want %q", g.sourceMDPath, "/content/page.md")
		}
		if g.buildDir != "/out" {
			t.Errorf("buildDir = %q, want %q", g.buildDir, "/out")
		}
		if g.sectionName != "blog" {
			t.Errorf("sectionName = %q, want %q", g.sectionName, "blog")
		}
	})
}

func TestGenerator_Generate(t *testing.T) {
	t.Run("generates html from markdown", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/index.md", []byte("# Hello World\n\nSome content here."))

		g := newTestGenerator(t, "/content/index.md", "/build/index.html", "/build", "", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		output, ok := fs.GetFile("/build/index.html")
		if !ok {
			t.Fatal("Generate() did not write output file")
		}

		html := string(output)
		if !strings.Contains(html, "Hello World") {
			t.Errorf("output should contain title, got %q", html)
		}
		if !strings.Contains(html, "Some content here.") {
			t.Errorf("output should contain content, got %q", html)
		}
		if !strings.Contains(html, "<title>Hello World</title>") {
			t.Errorf("output should have title tag, got %q", html)
		}
		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Errorf("output should contain DOCTYPE, got %q", html)
		}
	})

	t.Run("generates html with section name", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/posts/article.md", []byte("# My Article\n\nArticle body."))

		g := newTestGenerator(t, "/content/posts/article.md", "/build/posts/article.html", "/build", "posts", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		_, ok := fs.GetFile("/build/posts/article.html")
		if !ok {
			t.Fatal("Generate() did not write output file at section path")
		}
	})

	t.Run("returns error when markdown file not found", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		g := newTestGenerator(t, "/content/missing.md", "/build/missing.html", "/build", "", fs)

		err := g.Generate()
		if err == nil {
			t.Fatal("Generate() expected error for missing file, got nil")
		}
		if !strings.Contains(err.Error(), "reading") {
			t.Errorf("error should mention reading, got %q", err.Error())
		}
	})

	t.Run("stores html content bytes after generation", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/page.md", []byte("# Title\n\nBody text."))

		g := newTestGenerator(t, "/content/page.md", "/build/page.html", "/build", "", fs)
		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		if len(g.htmlContentBytes) == 0 {
			t.Error("htmlContentBytes should be populated after Generate()")
		}
	})
}

func TestGenerator_Generate_MkdirAllError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.AddFile("/content/page.md", []byte("# Title\n\nContent."))
	fs.MkdirAllErr = fmt.Errorf("permission denied")

	g := newTestGenerator(t, "/content/page.md", "/build/page.html", "/build", "", fs)

	err := g.Generate()
	if err == nil {
		t.Fatal("Generate() expected error when MkdirAll fails, got nil")
	}
	if !strings.Contains(err.Error(), "output directory") {
		t.Errorf("error should mention output directory, got %q", err.Error())
	}
}

func TestGenerator_Generate_WriteFileError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.AddFile("/content/page.md", []byte("# Title\n\nContent."))
	fs.WriteFileErr = fmt.Errorf("disk full")

	g := newTestGenerator(t, "/content/page.md", "/build/page.html", "/build", "", fs)

	err := g.Generate()
	if err == nil {
		t.Fatal("Generate() expected error when WriteFile fails, got nil")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Errorf("error should mention write, got %q", err.Error())
	}
}

type htmlSubsMock struct {
	err error
}

func (sm htmlSubsMock) Placeholder() string {
	return ""
}

func (sm htmlSubsMock) Resolve(content string) (string, error) {
	return "", sm.err
}

func TestGenerator_Generate_HTMLSubstitutionError(t *testing.T) {
	// given
	expectedErr := fmt.Errorf("this is an html substitution error")
	sm := htmlSubsMock{err: expectedErr}

	// setup
	htmlSubs := htmlsubstitution.NewRegistryWithSubstituters(sm)
	fs := filesystem.NewMemoryFileSystem()
	fs.AddFile("/content/notitle.md", []byte("No heading here, just text."))
	g := NewGenerator(
		"/content/notitle.md",
		"/build/notitle.html",
		"/build",
		"",
		fs,
		mdsubstitution.NewRegistryWithSubstituters(),
		htmlSubs,
		validation.NewRegistryWithValidators(),
	)

	err := g.Generate()

	// expect
	assert.ErrorIs(t, err, expectedErr)
}

func TestGenerator_Generate_NavigationBar(t *testing.T) {
	t.Run("generates navigation with sections from root", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/index.md", []byte("# Home\n\nWelcome."))

		subs := htmlsubstitution.NewRegistry("/build/index.html", "/content/index.md", nil, nil, []section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}, {DirName: "about", DisplayName: "About"}}, "")
		vals := validation.NewRegistry(nil, false)
		g := NewGenerator("/content/index.md", "/build/index.html", "/build", "", fs, mdsubstitution.NewRegistry("/content/index.md"), subs, vals)

		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		output, ok := fs.GetFile("/build/index.html")
		if !ok {
			t.Fatal("Generate() did not write output file")
		}

		html := string(output)
		if !strings.Contains(html, "<nav") {
			t.Error("output should contain nav element")
		}
		if !strings.Contains(html, `href="index.html"`) {
			t.Errorf("output should contain home link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="posts/index.html"`) {
			t.Errorf("output should contain posts link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="about/index.html"`) {
			t.Errorf("output should contain about link, got:\n%s", html)
		}
	})

	t.Run("generates navigation with relative paths from section", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/posts/hello.md", []byte("# Hello\n\nPost content."))

		subs := htmlsubstitution.NewRegistry("/build/posts/hello.html", "/content/posts/hello.md", nil, nil, []section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}, {DirName: "about", DisplayName: "About"}}, "posts")
		vals := validation.NewRegistry(nil, false)
		g := NewGenerator("/content/posts/hello.md", "/build/posts/hello.html", "/build", "posts", fs, mdsubstitution.NewRegistry("/content/posts/hello.md"), subs, vals)

		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		output, ok := fs.GetFile("/build/posts/hello.html")
		if !ok {
			t.Fatal("Generate() did not write output file")
		}

		html := string(output)
		if !strings.Contains(html, `href="../index.html"`) {
			t.Errorf("output should contain relative home link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="../posts/index.html"`) {
			t.Errorf("output should contain relative posts link, got:\n%s", html)
		}
		if !strings.Contains(html, `href="../about/index.html"`) {
			t.Errorf("output should contain relative about link, got:\n%s", html)
		}
	})
}

// fakeMarkdownSubstituter is a test double for the markdown substitution.Substituer interface
type fakeMarkdownSubstituter struct {
	placeholder string
	err         error
}

func (f fakeMarkdownSubstituter) Placeholder() string      { return f.placeholder }
func (f fakeMarkdownSubstituter) Resolve() (string, error) { return "", f.err }

func TestGenerator_Generate_MarkdownSubstitutionError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.AddFile("/content/page.md", []byte("# Title\n\n{{my-placeholder}}\n\nContent."))

	mdSubs := mdsubstitution.NewRegistryWithSubstituters(
		fakeMarkdownSubstituter{placeholder: "{{my-placeholder}}", err: fmt.Errorf("substitution failed")},
	)
	subs := htmlsubstitution.NewRegistry("/build/page.html", "/content/page.md", nil, nil, nil, "")
	vals := validation.NewRegistry(nil, false)
	g := NewGenerator("/content/page.md", "/build/page.html", "/build", "", fs, mdSubs, subs, vals)

	err := g.Generate()
	if err == nil {
		t.Fatal("Generate() expected error when markdown substitution fails, got nil")
	}
	if !strings.Contains(err.Error(), "substitution failed") {
		t.Errorf("error should mention substitution failure, got %q", err.Error())
	}
}

func TestGenerator_Validate(t *testing.T) {
	t.Run("validate returns nil with empty registry", func(t *testing.T) {
		fs := filesystem.NewMemoryFileSystem()
		fs.AddFile("/content/page.md", []byte("# Page\n\nContent."))

		subs := htmlsubstitution.NewRegistry("/build/page.html", "/content/page.md", nil, nil, nil, "")
		vals := validation.NewRegistryWithValidators()
		g := NewGenerator("/content/page.md", "/build/page.html", "/build", "", fs, mdsubstitution.NewRegistry("/content/page.md"), subs, vals)

		err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		err = g.Validate()
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})
}
