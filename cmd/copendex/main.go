package main

import (
	"os"

	"github.com/eaoum-ai/copendex/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
