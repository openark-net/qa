package presenter

import (
	"fmt"
	"path/filepath"

	"github.com/openark-net/qa/pkg/qa/domain"
)

type DirColumn struct {
	prefixes map[string]string
}

func NewDirColumn(cfg domain.ConfigSet, root string) DirColumn {
	labels := make(map[string]string)
	for _, dir := range dirsOf(cfg) {
		labels[dir] = relativeLabel(root, dir)
	}

	width := 0
	for _, label := range labels {
		if len(label) > width {
			width = len(label)
		}
	}

	prefixes := make(map[string]string, len(labels))
	for dir, label := range labels {
		prefixes[dir] = fmt.Sprintf("%-*s ", width+1, label+":")
	}
	return DirColumn{prefixes: prefixes}
}

func (c DirColumn) Prefix(workingDir string) string {
	return c.prefixes[workingDir]
}

func dirsOf(cfg domain.ConfigSet) []string {
	seen := make(map[string]struct{})
	var dirs []string
	add := func(dir string) {
		if _, ok := seen[dir]; ok {
			return
		}
		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}

	for dir := range cfg.Format {
		add(dir)
	}
	for _, cmd := range cfg.Checks {
		add(cmd.WorkingDir)
	}
	return dirs
}

func relativeLabel(root, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return "."
	}
	return "./" + rel
}
