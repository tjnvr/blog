package article

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tjnvr/blog/internal/generator/page/markdown/metadata"
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
	indexFilePath string
}

func NewPageArticlesLister(indexFilePath string) ListPageArticles {
	return ListPageArticles{
		indexFilePath: indexFilePath,
	}
}

func (la ListPageArticles) ListPrinters() ([]Article, error) {
	articles := make([]Article, 0)
	dir := filepath.Dir(la.indexFilePath)
	if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// skip subdirectories
		if info.IsDir() && path != dir {
			return filepath.SkipDir
		}
		// only first-level .md files, excluding the index file itself
		if filepath.Dir(path) != dir || filepath.Ext(path) != ".md" || path == la.indexFilePath {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := extractTitle(data)
		if name == "" {
			return nil
		}
		articles = append(articles, Article{
			name:      name,
			filePath:  filepath.Base(path),
			createdAt: metadata.Extract(data).CreationDate,
		})
		return nil
	}); err != nil {
		return []Article{}, err
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
