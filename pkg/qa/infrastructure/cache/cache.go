package cache

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/openark-net/qa/pkg/qa/domain"
)

const ttl = 7 * 24 * time.Hour

type Cache struct {
	ctx      context.Context
	git      *GitClient
	storage  Storage
	cacheDir string
	repoRoot string
	data     map[string]Entry
	mu       sync.Mutex
	results  map[string]bool
}

func New(ctx context.Context, cacheDir string) (*Cache, error) {
	git, err := NewGitClient(ctx)
	if err != nil {
		return nil, err
	}

	storage := Storage{}
	data, err := storage.Load(cacheDir, git.RepoRoot())
	if err != nil {
		data = make(map[string]Entry)
	}

	pruned := prune(data, time.Now(), ttl)

	return &Cache{
		ctx:      ctx,
		git:      git,
		storage:  storage,
		cacheDir: cacheDir,
		repoRoot: git.RepoRoot(),
		data:     pruned,
		results:  make(map[string]bool),
	}, nil
}

func (c *Cache) resolvePath(workingDir string) string {
	if filepath.IsAbs(workingDir) {
		rel, err := c.git.ToRelative(workingDir)
		if err != nil {
			return ""
		}
		return rel
	}
	return filepath.Clean(workingDir)
}

func cacheKey(relPath, cmd string) string {
	return relPath + "::" + cmd
}

func (c *Cache) Hit(cmd domain.Command) bool {
	relPath := c.resolvePath(cmd.WorkingDir)
	if relPath == "" {
		return false
	}

	dirty, err := c.git.IsDirty(c.ctx, relPath)
	if err != nil || dirty {
		return false
	}

	hash, err := c.git.TreeHash(c.ctx, relPath)
	if err != nil {
		return false
	}

	key := cacheKey(relPath, cmd.Cmd)
	entry, exists := c.data[key]
	if !exists {
		return false
	}

	return entry.Hash == hash
}

func (c *Cache) RecordResult(cmd domain.Command, success bool) {
	relPath := c.resolvePath(cmd.WorkingDir)
	if relPath == "" {
		return
	}

	key := cacheKey(relPath, cmd.Cmd)
	c.mu.Lock()
	c.results[key] = success
	c.mu.Unlock()
}

func (c *Cache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	for key, passed := range c.results {
		if !passed {
			continue
		}

		parts := strings.SplitN(key, "::", 2)
		path := parts[0]

		hash, err := c.git.TreeHash(c.ctx, path)
		if err != nil {
			continue
		}

		c.data[key] = Entry{
			Hash:     hash,
			LastPass: now,
		}
	}

	return c.storage.Save(c.cacheDir, c.repoRoot, c.data)
}

func prune(data map[string]Entry, now time.Time, maxAge time.Duration) map[string]Entry {
	result := make(map[string]Entry, len(data))
	cutoff := now.Add(-maxAge)

	for path, entry := range data {
		if entry.LastPass.After(cutoff) {
			result[path] = entry
		}
	}

	return result
}
