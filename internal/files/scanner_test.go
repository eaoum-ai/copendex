// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies repository file discovery and ignore handling.
package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eaoum-ai/cosha/internal/config"
)

func TestScannerMatchesNestedJavaAndIgnoresGitignore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src/main/java/com/example/AuthorizationService.java", "class AuthorizationService {}")
	writeFile(t, root, "build/generated/Ignored.java", "class Ignored {}")
	writeFile(t, root, "tmp/IgnoredByGitignore.java", "class IgnoredByGitignore {}")
	writeFile(t, root, ".gitignore", "tmp/\n")

	scanner, err := NewScanner(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	files, err := scanner.Scan()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("files = %#v, want exactly one indexed Java file", files)
	}
	if files[0].Path != "src/main/java/com/example/AuthorizationService.java" {
		t.Fatalf("path = %s", files[0].Path)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
