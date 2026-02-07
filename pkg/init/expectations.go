package setup

import (
	_ "embed"
	"errors"
	"os"
)

//go:embed templates/CLAUDE.md
var expectationsTemplate string

var ErrFileAlreadyExists = errors.New("file already exists")

func CopyExpectations(dest string) error {
	if dest == "" {
		dest = "./CLAUDE.md"
	}

	if _, err := os.Stat(dest); err == nil {
		return ErrFileAlreadyExists
	}

	return os.WriteFile(dest, []byte(expectationsTemplate), 0644)
}
