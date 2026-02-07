package cache

import (
	"context"
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

func (c *Cache) Hit(cmd domain.Command) bool {
	relPath, err := c.git.ToRelative(cmd.WorkingDir)
	if err != nil {
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

	entry, exists := c.data[relPath]
	if !exists {
		return false
	}

	return entry.Hash == hash
}

func (c *Cache) RecordResult(cmd domain.Command, success bool) {
	relPath, err := c.git.ToRelative(cmd.WorkingDir)
	if err != nil {
		return
	}

	if existing, ok := c.results[relPath]; ok {
		c.results[relPath] = existing && success
	} else {
		c.results[relPath] = success
	}
}

func (c *Cache) Flush() error {
	now := time.Now()

	for path, passed := range c.results {
		if !passed {
			continue
		}

		hash, err := c.git.TreeHash(c.ctx, path)
		if err != nil {
			continue
		}

		c.data[path] = Entry{
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
