package substitution

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeSubstituter struct {
	placeholder string
	resolution  string
	err         error
}

func (f fakeSubstituter) Placeholder() string      { return f.placeholder }
func (f fakeSubstituter) Resolve() (string, error) { return f.resolution, f.err }

func TestRegistry_NewRegistry(t *testing.T) {
	// given
	indexFile := "/content/posts/index.md"

	// test
	r := NewRegistry(indexFile)

	// expect
	assert.NotNil(t, r)
	assert.Equal(t, 1, len(r.substitutions))
}

func TestRegistry_NewRegistryWithSubstituters(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		// test
		r := NewRegistryWithSubstituters()

		// expect
		assert.NotNil(t, r)
		assert.Equal(t, 0, len(r.substitutions))
	})

	t.Run("custom substituters", func(t *testing.T) {
		// given
		s1 := fakeSubstituter{placeholder: "{{a}}", resolution: "A"}
		s2 := fakeSubstituter{placeholder: "{{b}}", resolution: "B"}

		// test
		r := NewRegistryWithSubstituters(s1, s2)

		// expect
		assert.Equal(t, 2, len(r.substitutions))
	})
}

func TestRegistry_Apply(t *testing.T) {
	tests := []struct {
		name    string
		subs    []Substituer
		content string
		want    string
		wantErr bool
	}{
		{
			name: "placeholder not present returns content unchanged",
			subs: []Substituer{
				fakeSubstituter{placeholder: "{{missing}}", resolution: "value"},
			},
			content: "no placeholders here",
			want:    "no placeholders here",
		},
		{
			name: "single substitution applied",
			subs: []Substituer{
				fakeSubstituter{placeholder: "{{list-child-articles}}", resolution: "- [Post](post.md)"},
			},
			content: "## Articles\n{{list-child-articles}}",
			want:    "## Articles\n- [Post](post.md)",
		},
		{
			name: "multiple substituters all applied",
			subs: []Substituer{
				fakeSubstituter{placeholder: "{{a}}", resolution: "AAA"},
				fakeSubstituter{placeholder: "{{b}}", resolution: "BBB"},
			},
			content: "{{a}} and {{b}}",
			want:    "AAA and BBB",
		},
		{
			name: "substituter error propagated",
			subs: []Substituer{
				fakeSubstituter{placeholder: "{{fail}}", err: fmt.Errorf("resolve error")},
			},
			content: "{{fail}}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			r := NewRegistryWithSubstituters(tt.subs...)

			// test
			got, err := r.Apply(tt.content)

			// expect
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
