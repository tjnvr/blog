package articlelisting

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"path/filepath"
	"sort"

	"github.com/spf13/afero"

	"github.com/tjnvr/blog/internal/backbone/metadata"
	"github.com/tjnvr/blog/internal/hrefpath"
	mfs "github.com/tjnvr/blog/internal/io/fs"
	"github.com/tjnvr/blog/internal/relpath"
)

//go:embed template.html
var listingTemplate string

// substituter resolves the {{list-child-pages}} placeholder with a
// generated listing of the first-level sibling article pages.
type substituter struct {
	fs                 afero.Fs
	template           *template.Template
	HTMLPath           string
	markdownPath       string
	assetPathsResolver relpath.Resolver
	pagePathsResolver  hrefpath.Resolver
}

func NewSubstituer(fs afero.Fs, HTMLPath, markdownPath string, assetPathsResolver relpath.Resolver, pagePathsResolver hrefpath.Resolver) substituter {
	return substituter{
		fs:                 fs,
		template:           template.Must(template.New("listing").Parse(listingTemplate)),
		HTMLPath:           HTMLPath,
		markdownPath:       markdownPath,
		assetPathsResolver: assetPathsResolver,
		pagePathsResolver:  pagePathsResolver,
	}
}

func (s substituter) Placeholder() string {
	return "<p>{{list-child-pages}}</p>"
}

// articleView is the data one article contributes to the listing template.
type articleView struct {
	HRef        string
	Title       string
	Description string
	ImageSrc    string
	PublishedOn string

	creationDate string // raw ISO date, kept only to sort articles chronologically
}

func (s substituter) Resolve(_ string) (string, error) {
	articles, err := s.listArticles()
	if err != nil {
		return "", err
	}
	if len(articles) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	if err := s.template.ExecuteTemplate(&buf, "listing", articles); err != nil {
		return "", fmt.Errorf("execute err: %v", err)
	}

	return buf.String(), nil
}

func (s substituter) listArticles() ([]articleView, error) {
	dir := filepath.Dir(s.markdownPath)
	finder := mfs.NewFilesFinder(s.fs, mfs.WithLevel(1), mfs.WithExtension(".md"))
	relPaths, err := finder.FindFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("FindFiles err: %v", err)
	}

	views := make([]articleView, 0, len(relPaths))
	for _, relPath := range relPaths {
		// exclude the index file itself
		if relPath == filepath.Base(s.markdownPath) {
			continue
		}

		fullPath := filepath.Join(dir, relPath)
		data, err := afero.ReadFile(s.fs, fullPath)
		if err != nil {
			return nil, err
		}

		m := metadata.Extract(data)
		if m.Title == "" {
			continue
		}

		href, err := s.pagePathsResolver.Resolve(fullPath)
		if err != nil {
			return nil, fmt.Errorf("pagePathsResolver.Resolve(%s) err: %v", fullPath, err)
		}

		publishedOn, err := formatLong(m.CreationDate)
		if err != nil {
			return nil, fmt.Errorf("article %q: %w", fullPath, err)
		}

		view := articleView{
			HRef:         href,
			Title:        m.Title,
			Description:  m.Description,
			PublishedOn:  publishedOn,
			creationDate: m.CreationDate,
		}

		if m.Image != "" {
			imagePath := filepath.Join(filepath.Dir(fullPath), m.Image)
			imageSrc, err := s.assetPathsResolver.Resolve(imagePath, s.HTMLPath)
			if err != nil {
				return nil, fmt.Errorf("assetPathsResolver.Resolve(%s) err: %v", imagePath, err)
			}
			view.ImageSrc = imageSrc
		}

		views = append(views, view)
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].creationDate > views[j].creationDate
	})

	return views, nil
}
