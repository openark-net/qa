package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Hash     string    `yaml:"hash"`
	LastPass time.Time `yaml:"last_pass"`
}

type Storage struct{}

func (Storage) Load(cacheDir, repoRoot string) (map[string]Entry, error) {
	path := cachePath(cacheDir, repoRoot)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]Entry), nil
		}
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	var entries map[string]Entry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing cache file: %w", err)
	}

	if entries == nil {
		return make(map[string]Entry), nil
	}

	return entries, nil
}

func (Storage) Save(cacheDir, repoRoot string, data map[string]Entry) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling cache data: %w", err)
	}

	path := cachePath(cacheDir, repoRoot)
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, content, 0644); err != nil {
		return fmt.Errorf("writing temp cache file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming cache file: %w", err)
	}

	return nil
}

func cachePath(cacheDir, repoRoot string) string {
	sanitized := strings.ReplaceAll(repoRoot, string(filepath.Separator), "_")
	return filepath.Join(cacheDir, sanitized+".yml")
}
