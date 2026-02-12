package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfig_InCurrentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	createFile(t, filepath.Join(tmpDir, ".qa.yml"))

	found, err := FindConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != tmpDir {
		t.Errorf("expected %q, got %q", tmpDir, found)
	}
}

func TestFindConfig_InParentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	createFile(t, filepath.Join(tmpDir, ".qa.yml"))

	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	found, err := FindConfig(childDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != tmpDir {
		t.Errorf("expected %q, got %q", tmpDir, found)
	}
}

func TestFindConfig_InGrandparentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	createFile(t, filepath.Join(tmpDir, ".qa.yml"))

	nestedDir := filepath.Join(tmpDir, "child", "grandchild")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}

	found, err := FindConfig(nestedDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != tmpDir {
		t.Errorf("expected %q, got %q", tmpDir, found)
	}
}

func TestFindConfig_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindConfig(tmpDir)
	if err != ErrConfigNotFound {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestFindConfig_StopsAtFirstMatch(t *testing.T) {
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	createFile(t, filepath.Join(tmpDir, ".qa.yml"))
	createFile(t, filepath.Join(childDir, ".qa.yml"))

	found, err := FindConfig(childDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != childDir {
		t.Errorf("expected %q (child), got %q", childDir, found)
	}
}

func createFile(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}
