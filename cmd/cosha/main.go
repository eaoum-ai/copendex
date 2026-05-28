// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Command cosha starts the local-first codebase intelligence CLI.
package main

import (
	"os"

	"github.com/eaoum-ai/cosha/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
