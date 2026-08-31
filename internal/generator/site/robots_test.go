package site

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRobotsTxt_WritesExpectedContent(t *testing.T) {
	// given
	fs := afero.NewMemMapFs()
	g := &Generator{Config: Config{PublicDir: "target"}, fs: fs}

	// test
	err := g.generateRobotsTxt()

	// expect
	require.NoError(t, err)
	content, err := afero.ReadFile(fs, "target/robots.txt")
	require.NoError(t, err)
	assert.Equal(t, "User-agent: *\nDisallow:\nSitemap: __BASE_URL__/sitemap.xml", string(content))
}
