package site

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	robotsFileName = "robots.txt"
	robotsContent  = "User-agent: *\nDisallow:\nSitemap: " + sitemapBaseURLPlaceholder + "/" + sitemapFileName
)

func (g *Generator) generateRobotsTxt() error {
	if err := afero.WriteFile(g.fs, filepath.Join(g.PublicDir, robotsFileName), []byte(robotsContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", robotsFileName, err)
	}
	return nil
}
