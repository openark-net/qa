package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNotGitRepo = errors.New("not a git repository")

type GitClient struct {
	repoRoot string
}

func NewGitClient(ctx context.Context) (*GitClient, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if strings.Contains(stderr.String(), "not a git repository") {
			return nil, ErrNotGitRepo
		}
		return nil, fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}

	return &GitClient{
		repoRoot: strings.TrimSpace(stdout.String()),
	}, nil
}

func (g *GitClient) RepoRoot() string {
	return g.repoRoot
}

func (g *GitClient) TreeHash(ctx context.Context, relativePath string) (string, error) {
	var ref string
	if relativePath == "." {
		ref = "HEAD^{tree}"
	} else {
		ref = "HEAD:" + relativePath
	}

	cmd := exec.CommandContext(ctx, "git", "rev-parse", ref)
	cmd.Dir = g.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("git rev-parse %s: %s", ref, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (g *GitClient) IsDirty(ctx context.Context, relativePath string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", relativePath)
	cmd.Dir = g.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return false, fmt.Errorf("git status --porcelain %s: %w", relativePath, err)
	}

	return stdout.Len() > 0, nil
}

func (g *GitClient) ToRelative(absolutePath string) (string, error) {
	rel, err := filepath.Rel(g.repoRoot, absolutePath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %s is outside repository %s", absolutePath, g.repoRoot)
	}
	if rel == "" {
		return ".", nil
	}
	return rel, nil
}
