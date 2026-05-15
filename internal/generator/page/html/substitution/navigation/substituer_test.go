package navigation

import (
	"strings"
	"testing"

	"github.com/tjnvr/blog/internal/generator/section"
)

func TestSubstituer_Placeholder(t *testing.T) {
	s := NewSubstituer(nil, "")
	if got := s.Placeholder(); got != "{{navigation}}" {
		t.Errorf("Placeholder() = %q, want %q", got, "{{navigation}}")
	}
}

func TestSubstituer_Resolve(t *testing.T) {
	tests := []struct {
		name           string
		sections       []section.Section
		currentSection string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "from root with no sections",
			sections:       []section.Section{{DirName: "", DisplayName: "Accueil"}},
			currentSection: "",
			wantContains:   []string{`href="index.html"`, "Accueil", "<nav"},
		},
		{
			name: "from root with sections",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
				{DirName: "about", DisplayName: "About"},
			},
			currentSection: "",
			wantContains: []string{
				`href="index.html"`,
				`href="posts/index.html"`,
				`href="about/index.html"`,
				"Accueil",
				"Posts",
				"About",
			},
		},
		{
			name: "from section with sections",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
				{DirName: "about", DisplayName: "About"},
			},
			currentSection: "posts",
			wantContains: []string{
				`href="../index.html"`,
				`href="../posts/index.html"`,
				`href="../about/index.html"`,
				"Accueil",
				"Posts",
				"About",
			},
			wantNotContain: []string{
				`href="index.html"`,
			},
		},
		{
			name: "from about section",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
				{DirName: "about", DisplayName: "About"},
			},
			currentSection: "about",
			wantContains: []string{
				`href="../index.html"`,
				`href="../posts/index.html"`,
				`href="../about/index.html"`,
			},
		},
		{
			name: "from nested section",
			sections: []section.Section{
				{DirName: "", DisplayName: "Accueil"},
				{DirName: "posts", DisplayName: "Posts"},
			},
			currentSection: "blog/2024",
			wantContains: []string{
				`href="../../index.html"`,
				`href="../../posts/index.html"`,
			},
		},
		{
			name: "home display name comes from root index.md title",
			sections: []section.Section{
				{DirName: "", DisplayName: "Bienvenue sur mon blog"},
				{DirName: "posts", DisplayName: "All Articles"},
			},
			currentSection: "",
			wantContains: []string{
				`href="index.html"`,
				"Bienvenue sur mon blog",
				`href="posts/index.html"`,
				"All Articles",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubstituer(tt.sections, tt.currentSection)
			got, err := s.Resolve("")
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			for _, substr := range tt.wantContains {
				if !strings.Contains(got, substr) {
					t.Errorf("Resolve() should contain %q, got:\n%s", substr, got)
				}
			}
			for _, substr := range tt.wantNotContain {
				if strings.Contains(got, substr) {
					t.Errorf("Resolve() should not contain %q, got:\n%s", substr, got)
				}
			}
		})
	}
}

func TestRelativePrefix(t *testing.T) {
	tests := []struct {
		section string
		want    string
	}{
		{"", ""},
		{"posts", "../"},
		{"about", "../"},
		{"blog/2024", "../../"},
		{"a/b/c", "../../../"},
	}
	for _, tt := range tests {
		t.Run(tt.section, func(t *testing.T) {
			if got := relativePrefix(tt.section); got != tt.want {
				t.Errorf("relativePrefix(%q) = %q, want %q", tt.section, got, tt.want)
			}
		})
	}
}

func TestSubstituer_Resolve_DisplayNameFromSection(t *testing.T) {
	s := NewSubstituer([]section.Section{
		{DirName: "", DisplayName: "Accueil"},
		{DirName: "posts", DisplayName: "My Blog Posts"},
	}, "")
	got, err := s.Resolve("")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if !strings.Contains(got, "My Blog Posts") {
		t.Errorf("display name should come from Section.DisplayName, got:\n%s", got)
	}
	if !strings.Contains(got, `href="posts/index.html"`) {
		t.Errorf("href should use Section.DirName, got:\n%s", got)
	}
}

func TestSubstituer_Resolve_ActiveSection(t *testing.T) {
	sections := []section.Section{
		{DirName: "", DisplayName: "Accueil"},
		{DirName: "posts", DisplayName: "Posts"},
		{DirName: "about", DisplayName: "About"},
	}

	tests := []struct {
		name           string
		currentSection string
		activeDirName  string
	}{
		{"home is active", "", ""},
		{"posts is active", "posts", "posts"},
		{"about is active", "about", "about"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubstituer(sections, tt.currentSection)
			got, err := s.Resolve("")
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			for _, sec := range sections {
				if sec.DirName == tt.activeDirName {
					if !strings.Contains(got, `class="font-semibold underline">`+sec.DisplayName) {
						t.Errorf("active section %q should have font-semibold class, got:\n%s", sec.DisplayName, got)
					}
				} else {
					if strings.Contains(got, `class="font-semibold underline">`+sec.DisplayName) {
						t.Errorf("inactive section %q should not have font-semibold class, got:\n%s", sec.DisplayName, got)
					}
				}
			}
		})
	}
}
