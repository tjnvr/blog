package article

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"github.com/tjnvr/blog/internal/generator/page/markdown/metadata"
	"github.com/tjnvr/blog/internal/io/fs"
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
	fs              afero.Fs
	filePathsLister fs.FilesFinder
	indexFilePath   string
}

func NewPageArticlesLister(fs afero.Fs, indexFilePath string, filePathsLister fs.FilesFinder) ListPageArticles {
	return ListPageArticles{
		fs:              fs,
		indexFilePath:   indexFilePath,
		filePathsLister: filePathsLister,
	}
}

func (la ListPageArticles) ListPrinters() ([]Article, error) {
	articles := make([]Article, 0)
	dir := filepath.Dir(la.indexFilePath)
	filePaths, err := la.filePathsLister.FindFiles(dir)
	if err != nil {
		return articles, fmt.Errorf("error on ListFilePaths: %s", err)
	}

	for _, path := range filePaths {
		// only first-level .md files, excluding the index file itself
		if filepath.Dir(path) != dir || path == la.indexFilePath {
			continue
		}

		data, err := afero.ReadFile(la.fs, path)
		if err != nil {
			return nil, err
		}
		name := extractTitle(data)
		if name == "" {
			continue
		}

		articles = append(articles, Article{
			name:      name,
			filePath:  filepath.Base(path),
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
