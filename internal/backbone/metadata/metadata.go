package metadata

import (
	"regexp"
	"strconv"
)

type Metadata struct {
	CreationDate string
	Seq          int // Display order. 0 means unsequenced.
	Description  string
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
		}
	}
	return m
}
