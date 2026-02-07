package main

import (
	"os"

	"github.com/openark-net/qa/pkg/qa/interfaces/cli"
)

func main() {
	os.Exit(cli.Run())
}
