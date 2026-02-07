package config

import (
	"testing"
	"testing/fstest"

	"github.com/openark-net/qa/pkg/qa/domain"
)

func TestLoad_FormatCommands(t *testing.T) {
	fsys := fstest.MapFS{
		".qa.yml": &fstest.MapFile{
			Data: []byte(`format:
  - "go fmt ./..."
  - "goimports -w ."
`),
		},
	}

	loader := New(fsys)
	cfg, err := loader.Load(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	formatCmds := cfg.Format["."]
	if len(formatCmds) != 2 {
		t.Fatalf("expected 2 format commands, got %d", len(formatCmds))
	}

	assertCommand(t, formatCmds[0], "go fmt ./...", ".")
	assertCommand(t, formatCmds[1], "goimports -w .", ".")
}

func TestLoad_CheckCommands(t *testing.T) {
	fsys := fstest.MapFS{
		".qa.yml": &fstest.MapFile{
			Data: []byte(`checks:
  - "go test ./..."
  - "go vet ./..."
`),
		},
	}

	loader := New(fsys)
	cfg, err := loader.Load(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Checks) != 2 {
		t.Fatalf("expected 2 check commands, got %d", len(cfg.Checks))
	}

	assertCommand(t, cfg.Checks[0], "go test ./...", ".")
	assertCommand(t, cfg.Checks[1], "go vet ./...", ".")
}

func TestLoad_IncludesFile(t *testing.T) {
	fsys := fstest.MapFS{
		".qa.yml": &fstest.MapFile{
			Data: []byte(`includes:
  - "subdir/.qa.yml"
format:
  - "root-fmt"
`),
		},
		"subdir/.qa.yml": &fstest.MapFile{
			Data: []byte(`format:
  - "sub-fmt"
checks:
  - "sub-check"
`),
		},
	}

	loader := New(fsys)
	cfg, err := loader.Load(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rootFmt := cfg.Format["."]
	if len(rootFmt) != 1 {
		t.Fatalf("expected 1 root format command, got %d", len(rootFmt))
	}
	assertCommand(t, rootFmt[0], "root-fmt", ".")

	subFmt := cfg.Format["subdir"]
	if len(subFmt) != 1 {
		t.Fatalf("expected 1 subdir format command, got %d", len(subFmt))
	}
	assertCommand(t, subFmt[0], "sub-fmt", "subdir")

	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check command, got %d", len(cfg.Checks))
	}
	assertCommand(t, cfg.Checks[0], "sub-check", "subdir")
}

func TestLoad_NestedIncludes(t *testing.T) {
	fsys := fstest.MapFS{
		".qa.yml": &fstest.MapFile{
			Data: []byte(`includes:
  - "level1/.qa.yml"
checks:
  - "root-check"
`),
		},
		"level1/.qa.yml": &fstest.MapFile{
			Data: []byte(`includes:
  - "level2/.qa.yml"
checks:
  - "level1-check"
`),
		},
		"level1/level2/.qa.yml": &fstest.MapFile{
			Data: []byte(`checks:
  - "level2-check"
`),
		},
	}

	loader := New(fsys)
	cfg, err := loader.Load(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Checks) != 3 {
		t.Fatalf("expected 3 check commands, got %d", len(cfg.Checks))
	}

	assertCommand(t, cfg.Checks[0], "root-check", ".")
	assertCommand(t, cfg.Checks[1], "level1-check", "level1")
	assertCommand(t, cfg.Checks[2], "level2-check", "level1/level2")
}

func TestLoad_CircularInclude(t *testing.T) {
	fsys := fstest.MapFS{
		".qa.yml": &fstest.MapFile{
			Data: []byte(`includes:
  - "a/.qa.yml"
`),
		},
		"a/.qa.yml": &fstest.MapFile{
			Data: []byte(`includes:
  - "b/.qa.yml"
`),
		},
		"a/b/.qa.yml": &fstest.MapFile{
			Data: []byte(`includes:
  - "../../.qa.yml"
`),
		},
	}

	loader := New(fsys)
	_, err := loader.Load(".")
	if err == nil {
		t.Fatal("expected circular include error")
	}
}

func assertCommand(t *testing.T, cmd domain.Command, expectedCmd, expectedDir string) {
	t.Helper()
	if cmd.Cmd != expectedCmd {
		t.Errorf("expected cmd %q, got %q", expectedCmd, cmd.Cmd)
	}
	if cmd.WorkingDir != expectedDir {
		t.Errorf("expected working dir %q, got %q", expectedDir, cmd.WorkingDir)
	}
}
