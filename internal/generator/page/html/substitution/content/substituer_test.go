package content

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockPathTranslater struct {
	resolveFn func(oldPath, fromPath string) (string, error)
}

func (m mockPathTranslater) Resolve(oldPath, fromPath string) (string, error) {
	return m.resolveFn(oldPath, fromPath)
}

func TestSubstituter_Resolve(t *testing.T) {
	linkErr := uuid.New().String()
	assetErr := uuid.New().String()
	resLink := uuid.New().String() + ".html"
	resAsset := uuid.New().String() + ".png"

	tests := []struct {
		name        string
		linkFn      func(string, string) (string, error)
		assetFn     func(string, string) (string, error)
		wantStr     string
		wantErrMsgs []string
	}{
		{
			name: "Successful resolution of both links and assets",
			linkFn: func(o, f string) (string, error) {
				return resLink, nil
			},
			assetFn: func(o, f string) (string, error) {
				return resAsset, nil
			},
			wantStr: fmt.Sprintf(`<a href="%s">Link</a> <img src="%s">`, resLink, resAsset),
		},
		{
			name: "Joins errors when both translating sub-systems fail",
			linkFn: func(o, f string) (string, error) {
				return "", errors.New(linkErr)
			},
			assetFn: func(o, f string) (string, error) {
				return "", errors.New(assetErr)
			},
			wantErrMsgs: []string{linkErr, assetErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			html := `<a href="link.md">Link</a> <img src="img.png">`

			// setup
			s := Substituter{
				markdownSourcePath:    "index.md",
				linksPathTranslater:   mockPathTranslater{resolveFn: tt.linkFn},
				assetsPathsTranslater: mockPathTranslater{resolveFn: tt.assetFn},
			}

			// test
			got, err := s.Resolve(html)

			// expect
			if len(tt.wantErrMsgs) > 0 {
				assert.Error(t, err)
				for _, msg := range tt.wantErrMsgs {
					assert.Contains(t, err.Error(), msg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStr, got)
			}
		})
	}
}

func TestSubstituter_ReplacePaths(t *testing.T) {
	resPath := uuid.New().String()
	errReason := uuid.New().String()
	imgRe := regexp.MustCompile(`(src=")([^"]+)(")`)
	hrefRe := regexp.MustCompile(`(href=")([^"]+)(")`)

	tests := []struct {
		name       string
		html       string
		re         *regexp.Regexp
		resolveFn  func(string, string) (string, error)
		modifyFn   func(string) string
		wantStr    string
		wantErrMsg string
	}{
		{
			name: "Substitutes relative asset path smoothly",
			html: `<img src="sub/image.png">`,
			re:   imgRe,
			resolveFn: func(o, f string) (string, error) {
				return resPath, nil
			},
			wantStr: fmt.Sprintf(`<img src="%s">`, resPath),
		},
		{
			name: "Applies path modifier callback before calling translater",
			html: `<a href="about">About</a>`,
			re:   hrefRe,
			modifyFn: func(p string) string {
				return p + ".processed"
			},
			resolveFn: func(o, f string) (string, error) {
				if strings.HasSuffix(o, ".processed") {
					return resPath, nil
				}
				return "", errors.New("missing suffix")
			},
			wantStr: fmt.Sprintf(`<a href="%s">About</a>`, resPath),
		},
		{
			name: "Skips external http addresses",
			html: `<img src="http://example.com/avatar.jpg">`,
			re:   imgRe,
			resolveFn: func(o, f string) (string, error) {
				return uuid.New().String(), nil
			},
			wantStr: `<img src="http://example.com/avatar.jpg">`,
		},
		{
			name: "Skips external https addresses",
			html: `<img src="https://example.com/avatar.jpg">`,
			re:   imgRe,
			resolveFn: func(o, f string) (string, error) {
				return uuid.New().String(), nil
			},
			wantStr: `<img src="https://example.com/avatar.jpg">`,
		},
		{
			name: "Skips local absolute routes",
			html: `<img src="/global-assets/avatar.jpg">`,
			re:   imgRe,
			resolveFn: func(o, f string) (string, error) {
				return uuid.New().String(), nil
			},
			wantStr: `<img src="/global-assets/avatar.jpg">`,
		},
		{
			name: "Accumulates multiple sequential processing errors",
			html: `<img src="err1.png"><img src="err2.png">`,
			re:   imgRe,
			resolveFn: func(o, f string) (string, error) {
				return "", errors.New(errReason)
			},
			wantErrMsg: fmt.Sprintf("%s\n%s", errReason, errReason),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			targetFile := uuid.New().String()

			// setup
			s := Substituter{markdownSourcePath: "docs/index.md"}
			mock := mockPathTranslater{resolveFn: tt.resolveFn}

			// test
			got, err := s.replacePaths(tt.html, targetFile, tt.re, mock, tt.modifyFn)

			// expect
			if tt.wantErrMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStr, got)
			}
		})
	}
}

func TestSubstituter_ConvertMdLinksPath(t *testing.T) {
	resLink := uuid.New().String() + ".html"

	tests := []struct {
		name    string
		html    string
		wantStr string
	}{
		{
			name:    "Converts typical relative markdown link extension",
			html:    `<a href="posts/hello.md">Hello</a>`,
			wantStr: fmt.Sprintf(`<a href="%s">Hello</a>`, resLink),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			targetFile := uuid.New().String()

			// setup
			s := Substituter{
				markdownSourcePath: "index.md",
				linksPathTranslater: mockPathTranslater{
					resolveFn: func(oldPath, fromPath string) (string, error) {
						return resLink, nil
					},
				},
			}

			// test
			got, err := s.convertMdLinksPath(tt.html, targetFile)

			// expect
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStr, got)
		})
	}
}

func TestSubstituter_ConvertAssetsPath(t *testing.T) {
	resAsset := uuid.New().String() + ".png"

	tests := []struct {
		name    string
		html    string
		wantStr string
	}{
		{
			name:    "Swaps basic image node relative source attribute",
			html:    `<img src="images/photo.png">`,
			wantStr: fmt.Sprintf(`<img src="%s">`, resAsset),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			targetFile := uuid.New().String()

			// setup
			s := Substituter{
				markdownSourcePath: "index.md",
				assetsPathsTranslater: mockPathTranslater{
					resolveFn: func(oldPath, fromPath string) (string, error) {
						return resAsset, nil
					},
				},
			}

			// test
			got, err := s.convertAssetsPath(tt.html, targetFile)

			// expect
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStr, got)
		})
	}
}
