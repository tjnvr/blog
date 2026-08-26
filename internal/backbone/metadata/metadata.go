package metadata

import (
	"regexp"
	"strconv"
	"strings"
)

type Metadata struct {
	Title        string // Highest heading (e.g # heading).
	CreationDate string
	Seq          int // Display order. 0 means unsequenced.
	Description  string
	Image        string // Path to an illustration image, relative to the markdown file it is declared in.
}

var metadataRegexp = regexp.MustCompile(`<!--\s*(\S+):\s*(.+?)\s*-->`)

func Extract(data []byte) Metadata {
	var m Metadata
	for _, match := range metadataRegexp.FindAllSubmatch(data, -1) {
		key := string(match[1])
		value := string(match[2])
		switch key {
		case "creation-date":
			m.CreationDate = value
		case "seq":
			if seq, err := strconv.Atoi(value); err == nil {
				m.Seq = seq
			}
		case "description":
			m.Description = value
		case "image":
			m.Image = value
		}
	}
	m.Title = extractTitle(data)
	return m
}

func extractTitle(data []byte) string {
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}
