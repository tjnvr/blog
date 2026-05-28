package substitution

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSubstituter struct {
	placeholder string
	resolution  string
	err         error
}

func (f fakeSubstituter) Placeholder() string      { return f.placeholder }
func (f fakeSubstituter) Resolve() (string, error) { return f.resolution, f.err }

func TestNewRegistry(t *testing.T) {
	r := NewRegistry("/content/posts/index.md", afero.NewMemMapFs())
	require.NotNil(t, r)
	assert.Len(t, r.substitutions, 1)
}

func TestNewRegistryWithSubstituters(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		r := NewRegistryWithSubstituters()
		require.NotNil(t, r)
		assert.Len(t, r.substitutions, 0)
	})

	t.Run("custom substituters", func(t *testing.T) {
		s1 := fakeSubstituter{placeholder: "{{a}}", resolution: "A"}
		s2 := fakeSubstituter{placeholder: "{{b}}", resolution: "B"}
		r := NewRegistryWithSubstituters(s1, s2)
		assert.Len(t, r.substitutions, 2)
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
			r := NewRegistryWithSubstituters(tt.subs...)
			got, err := r.Apply(tt.content)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
