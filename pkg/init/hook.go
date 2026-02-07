package setup

import (
	"errors"
	"os"
	"path/filepath"
)

const hookScript = `#!/bin/bash
qa
`

var (
	ErrNotGitRepo        = errors.New(".git directory not found")
	ErrHookAlreadyExists = errors.New("pre-commit hook already exists")
)

func InstallHook() error {
	gitDir := ".git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return ErrNotGitRepo
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		return ErrHookAlreadyExists
	}

	return os.WriteFile(hookPath, []byte(hookScript), 0755)
}
