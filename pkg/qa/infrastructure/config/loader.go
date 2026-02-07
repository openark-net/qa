package config

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/openark-net/qa/pkg/qa/domain"
	"gopkg.in/yaml.v3"
)

type qaFile struct {
	Includes []string `yaml:"includes"`
	Format   []string `yaml:"format"`
	Checks   []string `yaml:"checks"`
}

type Loader struct {
	fsys fs.FS
}

func New(fsys fs.FS) *Loader {
	return &Loader{fsys: fsys}
}

func (l *Loader) Load(rootPath string) (domain.ConfigSet, error) {
	configPath := path.Join(rootPath, ".qa.yml")
	visited := make(map[string]bool)
	return l.loadFile(configPath, visited)
}

func (l *Loader) loadFile(filePath string, visited map[string]bool) (domain.ConfigSet, error) {
	cleanPath := path.Clean(filePath)

	if visited[cleanPath] {
		return domain.ConfigSet{}, fmt.Errorf("circular include detected: %s", cleanPath)
	}
	visited[cleanPath] = true

	data, err := fs.ReadFile(l.fsys, cleanPath)
	if err != nil {
		return domain.ConfigSet{}, fmt.Errorf("reading %s: %w", cleanPath, err)
	}

	var file qaFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return domain.ConfigSet{}, fmt.Errorf("parsing %s: %w", cleanPath, err)
	}

	dir := path.Dir(cleanPath)
	result := domain.ConfigSet{
		Format: make(map[string][]domain.Command),
	}

	for _, cmdStr := range file.Format {
		result.Format[dir] = append(result.Format[dir], domain.Command{
			Cmd:        cmdStr,
			WorkingDir: dir,
		})
	}

	for _, cmdStr := range file.Checks {
		result.Checks = append(result.Checks, domain.Command{
			Cmd:        cmdStr,
			WorkingDir: dir,
		})
	}

	for _, include := range file.Includes {
		includePath := path.Join(dir, include)
		included, err := l.loadFile(includePath, visited)
		if err != nil {
			return domain.ConfigSet{}, err
		}
		result = merge(result, included)
	}

	return result, nil
}

func merge(a, b domain.ConfigSet) domain.ConfigSet {
	for dir, cmds := range b.Format {
		a.Format[dir] = append(a.Format[dir], cmds...)
	}
	a.Checks = append(a.Checks, b.Checks...)
	return a
}
