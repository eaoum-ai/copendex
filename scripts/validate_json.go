// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: validate_json <path>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
