// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies static report generation embeds index data.
package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eaoum-ai/copendex/internal/index"
)

func TestWriteReportEmbedsIndexData(t *testing.T) {
	root := t.TempDir()
	store, err := index.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	files := []index.File{{
		Path:         "src/main/java/com/example/AuthorizationService.java",
		Language:     "java",
		SizeBytes:    42,
		LastModified: time.Now(),
		Hash:         "hash",
	}}
	symbols := map[string][]index.Symbol{
		files[0].Path: {{
			Name:        "AuthorizationService",
			Kind:        "class",
			Language:    "java",
			File:        files[0].Path,
			PackageName: "com.example",
			Line:        12,
		}},
	}
	if err := store.Rebuild(files, symbols); err != nil {
		t.Fatal(err)
	}

	outPath := DefaultReportPath(root)
	if err := WriteReport(store, outPath); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	html := string(content)
	if strings.Contains(html, "__COPENDEX_DATA__") {
		t.Fatal("report still contains placeholder")
	}
	if !strings.Contains(html, "AuthorizationService") {
		t.Fatalf("report does not contain indexed symbol: %s", html)
	}
}
