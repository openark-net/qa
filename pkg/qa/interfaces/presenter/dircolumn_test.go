package presenter

import (
	"testing"

	"github.com/openark-net/qa/pkg/qa/domain"
)

func TestDirColumn_AlignsCommandStart(t *testing.T) {
	root := "/repo"
	cfg := domain.ConfigSet{
		Checks: []domain.Command{
			{Cmd: "npm run whatever", WorkingDir: "/repo/a/long/directory/path"},
			{Cmd: "go vulncheck", WorkingDir: "/repo/api"},
			{Cmd: "go test ./...", WorkingDir: "/repo"},
		},
	}

	col := NewDirColumn(cfg, root)

	cases := map[string]string{
		"/repo/a/long/directory/path": "./a/long/directory/path: ",
		"/repo/api":                   "./api:                   ",
		"/repo":                       ".:                       ",
	}

	width := len(cases["/repo/a/long/directory/path"])
	for dir, want := range cases {
		got := col.Prefix(dir)
		if got != want {
			t.Errorf("Prefix(%q) = %q, want %q", dir, got, want)
		}
		if len(got) != width {
			t.Errorf("Prefix(%q) width = %d, want %d", dir, len(got), width)
		}
	}
}

func TestDirColumn_IncludesFormatDirs(t *testing.T) {
	cfg := domain.ConfigSet{
		Format: map[string][]domain.Command{
			"/repo/web": {{Cmd: "prettier", WorkingDir: "/repo/web"}},
		},
	}

	col := NewDirColumn(cfg, "/repo")

	if got := col.Prefix("/repo/web"); got != "./web: " {
		t.Errorf("Prefix(/repo/web) = %q, want %q", got, "./web: ")
	}
}

func TestDirColumn_UnknownDirHasNoPrefix(t *testing.T) {
	col := NewDirColumn(domain.ConfigSet{}, "/repo")

	if got := col.Prefix("/repo/missing"); got != "" {
		t.Errorf("Prefix(unknown) = %q, want empty", got)
	}
}
