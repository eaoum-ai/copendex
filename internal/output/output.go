// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package output formats command responses as text or JSON.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func JSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func Lines(w io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}
