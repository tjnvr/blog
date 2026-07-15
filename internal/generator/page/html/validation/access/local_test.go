package access

import (
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLocalResolver_Resolve_ShouldResolveRelativeRefFromHTMLDirectory(t *testing.T) {
	// setup
	resolver := NewResolver(afero.NewMemMapFs(), "/build", "/build/posts/article.html")

	// test
	path := resolver.Resolve("../assets/logo.png")

	// expect
	assert.Equal(t, "/build/assets/logo.png", path)
}

func TestLocalResolver_Resolve_ShouldResolveAbsoluteRefFromBuildRoot(t *testing.T) {
	// setup
	resolver := NewResolver(afero.NewMemMapFs(), "/build", "/build/posts/article.html")

	// test
	path := resolver.Resolve("/assets/logo.png")

	// expect
	assert.Equal(t, "/build/assets/logo.png", path)
}

func TestLocalResolver_Resolve_ShouldIgnoreFragment(t *testing.T) {
	// setup
	resolver := NewResolver(afero.NewMemMapFs(), "/build", "/build/index.html")

	// test
	path := resolver.Resolve("/posts/article.html#section")

	// expect
	assert.Equal(t, "/build/posts/article.html", path)
}

func TestLocalResolver_Exists_ShouldReturnTrueWhenFileExists(t *testing.T) {
	// given
	name := uuid.New().String() + ".png"

	// setup
	fs := afero.NewMemMapFs()
	assert.NoError(t, afero.WriteFile(fs, "/build/assets/"+name, []byte("x"), 0644))
	resolver := NewResolver(fs, "/build", "/build/posts/article.html")

	// test
	ok, err := resolver.Exists("/assets/" + name)

	// expect
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestLocalResolver_Exists_ShouldReturnFalseWhenFileIsMissing(t *testing.T) {
	// given
	name := uuid.New().String() + ".png"

	// setup
	resolver := NewResolver(afero.NewMemMapFs(), "/build", "/build/posts/article.html")

	// test
	ok, err := resolver.Exists("/assets/" + name)

	// expect
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestLocalResolver_Exists_ShouldReturnTrueForSamePageFragment(t *testing.T) {
	// setup
	resolver := NewResolver(afero.NewMemMapFs(), "/build", "/build/index.html")

	// test
	ok, err := resolver.Exists("#section")

	// expect
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestLocalResolver_Exists_ShouldReturnTrueWhenTargetIsAnExistingDirectory(t *testing.T) {
	// setup
	fs := afero.NewMemMapFs()
	assert.NoError(t, afero.WriteFile(fs, "/build/posts/index.html", []byte("x"), 0644))
	resolver := NewResolver(fs, "/build", "/build/index.html")

	// test
	ok, err := resolver.Exists("/posts")

	// expect
	assert.NoError(t, err)
	assert.True(t, ok)
}
