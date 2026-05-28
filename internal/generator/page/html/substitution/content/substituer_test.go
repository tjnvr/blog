package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPathTranslater struct {
	newPath string
}

func (m mockPathTranslater) GetNewPath(oldPath, fromPath string) (string, error) {
	return m.newPath, nil
}

func TestConvertAssetsPath(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		filePath string
		newPath  string
		want     string
	}{
		{
			name:     "replaces relative image path",
			html:     `<img src="images/photo.png">`,
			filePath: "content/page.md",
			newPath:  "assets/photo.png",
			want:     `<img src="assets/photo.png">`,
		},
		{
			name:     "preserves attributes after src",
			html:     `<img src="images/photo.png" alt="a photo">`,
			filePath: "content/page.md",
			newPath:  "assets/photo.png",
			want:     `<img src="assets/photo.png" alt="a photo">`,
		},
		{
			name:     "preserves attributes before src",
			html:     `<img alt="a photo" src="images/photo.png">`,
			filePath: "content/page.md",
			newPath:  "assets/photo.png",
			want:     `<img alt="a photo" src="assets/photo.png">`,
		},
		{
			name:     "skips http URL",
			html:     `<img src="http://example.com/photo.png">`,
			filePath: "content/page.md",
			newPath:  "should-not-be-used",
			want:     `<img src="http://example.com/photo.png">`,
		},
		{
			name:     "skips https URL",
			html:     `<img src="https://example.com/photo.png">`,
			filePath: "content/page.md",
			newPath:  "should-not-be-used",
			want:     `<img src="https://example.com/photo.png">`,
		},
		{
			name:     "skips absolute path",
			html:     `<img src="/images/photo.png">`,
			filePath: "content/page.md",
			newPath:  "should-not-be-used",
			want:     `<img src="/images/photo.png">`,
		},
		{
			name:     "replaces multiple images",
			html:     `<img src="a.png"><img src="b.png">`,
			filePath: "content/page.md",
			newPath:  "new.png",
			want:     `<img src="new.png"><img src="new.png">`,
		},
		{
			name:     "no img tags returns html unchanged",
			html:     `<p>no images here</p>`,
			filePath: "content/page.md",
			newPath:  "unused",
			want:     `<p>no images here</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Substituter{
				markdownSourcePath:    tt.filePath,
				assetsPathsTranslater: mockPathTranslater{newPath: tt.newPath},
			}
			got, err := s.convertAssetsPath(tt.html, tt.filePath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertMdLinksPath(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		filePath string
		newPath  string
		want     string
	}{
		{
			name:     "replaces relative md link",
			html:     `<a href="posts/hello.md">Hello</a>`,
			filePath: "index.md",
			newPath:  "posts/hello.html",
			want:     `<a href="posts/hello.html">Hello</a>`,
		},
		{
			name:     "replaces parent relative md link",
			html:     `<a href="../index.md">Home</a>`,
			filePath: "posts/hello.md",
			newPath:  "../index.html",
			want:     `<a href="../index.html">Home</a>`,
		},
		{
			name:     "skips http URL",
			html:     `<a href="http://example.com/page.md">External</a>`,
			filePath: "index.md",
			newPath:  "should-not-be-used",
			want:     `<a href="http://example.com/page.md">External</a>`,
		},
		{
			name:     "skips https URL",
			html:     `<a href="https://example.com/page.md">External</a>`,
			filePath: "index.md",
			newPath:  "should-not-be-used",
			want:     `<a href="https://example.com/page.md">External</a>`,
		},
		{
			name:     "skips absolute path",
			html:     `<a href="/pages/about.md">About</a>`,
			filePath: "index.md",
			newPath:  "should-not-be-used",
			want:     `<a href="/pages/about.md">About</a>`,
		},
		{
			name:     "ignores non-md links",
			html:     `<a href="style.css">CSS</a>`,
			filePath: "index.md",
			newPath:  "should-not-be-used",
			want:     `<a href="style.css">CSS</a>`,
		},
		{
			name:     "replaces multiple md links",
			html:     `<a href="a.md">A</a> and <a href="b.md">B</a>`,
			filePath: "index.md",
			newPath:  "new.html",
			want:     `<a href="new.html">A</a> and <a href="new.html">B</a>`,
		},
		{
			name:     "preserves other attributes",
			html:     `<a class="nav" href="page.md" title="Page">Link</a>`,
			filePath: "index.md",
			newPath:  "page.html",
			want:     `<a class="nav" href="page.html" title="Page">Link</a>`,
		},
		{
			name:     "no links returns html unchanged",
			html:     `<p>no links here</p>`,
			filePath: "index.md",
			newPath:  "unused",
			want:     `<p>no links here</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Substituter{
				markdownSourcePath:  tt.filePath,
				linksPathTranslater: mockPathTranslater{newPath: tt.newPath},
			}
			got, err := s.convertMdLinksPath(tt.html, tt.filePath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
