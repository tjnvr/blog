package substitution

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tjnvr/blog/internal/generator/section"
)

// fakeSubstituter is a test double implementing Substituter
type fakeSubstituter struct {
	placeholder string
	resolveFunc func(string) (string, error)
}

func (f fakeSubstituter) Placeholder() string { return f.placeholder }
func (f fakeSubstituter) Resolve(content string) (string, error) {
	return f.resolveFunc(content)
}

func TestNewRegistry(t *testing.T) {
	sections := []section.Section{
		{DirName: "", DisplayName: "Accueil"},
		{DirName: "posts", DisplayName: "Posts"},
		{DirName: "about", DisplayName: "About"},
	}
	r := NewRegistry("output.html", "source.md", nil, nil, sections, "")
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
		return
	}
	if len(r.substitutions) != 4 {
		t.Errorf("NewRegistry() should have 4 default substituters, got %d", len(r.substitutions))
	}
}

func TestNewRegistryWithSubstituters(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		r := NewRegistryWithSubstituters()
		if r == nil {
			t.Fatal("NewRegistryWithSubstituters() returned nil")
			return
		}
		if len(r.substitutions) != 0 {
			t.Errorf("expected 0 substituters, got %d", len(r.substitutions))
		}
	})

	t.Run("custom substituters", func(t *testing.T) {
		s1 := fakeSubstituter{placeholder: "{{a}}", resolveFunc: func(string) (string, error) { return "A", nil }}
		s2 := fakeSubstituter{placeholder: "{{b}}", resolveFunc: func(string) (string, error) { return "B", nil }}
		r := NewRegistryWithSubstituters(s1, s2)
		if len(r.substitutions) != 2 {
			t.Errorf("expected 2 substituters, got %d", len(r.substitutions))
		}
	})
}

func TestRegistry_Apply(t *testing.T) {
	tests := []struct {
		name     string
		subs     []Substituer
		template string
		content  string
		want     string
		wantErr  bool
	}{
		{
			name: "applies single substitution",
			subs: []Substituer{
				fakeSubstituter{
					placeholder: "{{name}}",
					resolveFunc: func(string) (string, error) { return "World", nil },
				},
			},
			template: "Hello {{name}}!",
			content:  "ignored",
			want:     "Hello World!",
		},
		{
			name: "applies multiple substitutions",
			subs: []Substituer{
				fakeSubstituter{
					placeholder: "{{title}}",
					resolveFunc: func(string) (string, error) { return "My Title", nil },
				},
				fakeSubstituter{
					placeholder: "{{body}}",
					resolveFunc: func(string) (string, error) { return "<p>content</p>", nil },
				},
			},
			template: "<h1>{{title}}</h1><div>{{body}}</div>",
			content:  "source",
			want:     "<h1>My Title</h1><div><p>content</p></div>",
		},
		{
			name:     "no substitutions returns template as-is",
			subs:     []Substituer{},
			template: "Hello World",
			content:  "anything",
			want:     "Hello World",
		},
		{
			name: "returns error when resolve fails",
			subs: []Substituer{
				fakeSubstituter{
					placeholder: "{{fail}}",
					resolveFunc: func(string) (string, error) { return "", fmt.Errorf("resolve error") },
				},
			},
			template: "Hello {{fail}}",
			content:  "anything",
			wantErr:  true,
		},
		{
			name: "replaces all occurrences of same placeholder",
			subs: []Substituer{
				fakeSubstituter{
					placeholder: "{{x}}",
					resolveFunc: func(string) (string, error) { return "val", nil },
				},
			},
			template: "{{x}} and {{x}}",
			content:  "source",
			want:     "val and val",
		},
		{
			name: "template without matching placeholder unchanged",
			subs: []Substituer{
				fakeSubstituter{
					placeholder: "{{missing}}",
					resolveFunc: func(string) (string, error) { return "value", nil },
				},
			},
			template: "No placeholders here",
			content:  "source",
			want:     "No placeholders here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistryWithSubstituters(tt.subs...)
			got, err := r.Apply(tt.template, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Error("Apply() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Apply() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Apply() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRegistry_Apply_WithDefaultSubstituters(t *testing.T) {
	r := NewRegistry("output.html", "source.md", nil, nil, []section.Section{{DirName: "", DisplayName: "Accueil"}, {DirName: "posts", DisplayName: "Posts"}}, "")
	template := `<title>{{title}}</title><div>{{navigation}}</div><body>{{content}}</body>`
	content := `<h1>Test Title</h1><p>Hello world</p>`

	result, err := r.Apply(template, content)
	if err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}

	if !strings.Contains(result, "<title>Test Title</title>") {
		t.Errorf("expected title substitution, got %q", result)
	}
	if !strings.Contains(result, "<p>Hello world</p>") {
		t.Errorf("expected content substitution, got %q", result)
	}
	if !strings.Contains(result, "<nav") {
		t.Errorf("expected navigation substitution, got %q", result)
	}
}

func TestRegistry_Apply_SummarySubstitution(t *testing.T) {
	r := NewRegistry("output.html", "source.md", nil, nil, []section.Section{{DirName: "", DisplayName: "Accueil"}}, "")
	template := `<body>{{content}}</body>`
	content := `<h1 id="title">Title</h1>` +
		`<p>{{summary}}</p>` +
		`<h2 id="intro"><a href="#intro" class="heading-anchor">#</a>Introduction</h2>` +
		`<h3 id="details"><a href="#details" class="heading-anchor">#</a>Details</h3>`

	result, err := r.Apply(template, content)
	if err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}

	if strings.Contains(result, "{{summary}}") {
		t.Error("{{summary}} placeholder should have been replaced")
	}
	if !strings.Contains(result, `href="#intro"`) {
		t.Errorf("expected summary link to #intro, got %q", result)
	}
	if !strings.Contains(result, `href="#details"`) {
		t.Errorf("expected summary link to #details, got %q", result)
	}
}
