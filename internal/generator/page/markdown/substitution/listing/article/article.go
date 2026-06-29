package article

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/page/markdown/metadata"
	mfs "github.com/tjnvr/blog/internal/io/fs"
)

type Article struct {
	name      string
	filePath  string
	createdAt string
}

func (a Article) Print() string {
	if a.createdAt != "" {
		return fmt.Sprintf("- [%s](%s) · *%s*", a.name, a.filePath, a.createdAt)
	}
	return fmt.Sprintf("- [%s](%s)", a.name, a.filePath)
}

type ListPageArticles struct {
	fs                     afero.Fs
	articleFilePathsFinder mfs.FilesFinder
	articlesHomeFilePath   string
}

func NewPageArticlesLister(fs afero.Fs, articlesHomeFilePath string) ListPageArticles {
	return ListPageArticles{
		fs:                   fs,
		articlesHomeFilePath: articlesHomeFilePath,
		articleFilePathsFinder: mfs.NewFilesFinder(fs,
			mfs.WithLevel(1),
			mfs.WithExtension(".md"),
		),
	}
}

func (la ListPageArticles) ListPrinters() ([]Article, error) {
	articles := make([]Article, 0)
	dir := filepath.Dir(la.articlesHomeFilePath)
	filePaths, err := la.articleFilePathsFinder.FindFiles(dir)
	if err != nil {
		return articles, fmt.Errorf("error on ListFilePaths: %s", err)
	}

	for _, relPath := range filePaths {
		// only first-level .md files, excluding the index file itself
		if relPath == filepath.Base(la.articlesHomeFilePath) {
			continue
		}

		fullPath := filepath.Join(dir, relPath)
		data, err := afero.ReadFile(la.fs, fullPath)
		if err != nil {
			return nil, err
		}
		name := extractTitle(data)
		if name == "" {
			continue
		}

		articles = append(articles, Article{
			name:      name,
			filePath:  relPath,
			createdAt: metadata.Extract(data).CreationDate,
		})
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].createdAt > articles[j].createdAt
	})

	return articles, nil
}

func extractTitle(data []byte) string {
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return after
		}
	}
	return ""
}
