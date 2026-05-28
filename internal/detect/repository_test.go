// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies repository language and Java project marker detection.
package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eaoum-ai/cosha/internal/config"
)

func TestRepositoryTypeDetectsJavaSourceAndBuildMarkers(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "pom.xml", "<project></project>")
	writeFile(t, root, "module/build.gradle.kts", "plugins { java }")
	writeFile(t, root, "src/main/java/com/example/App.java", "class App {}")
	writeFile(t, root, "target/generated/Ignored.java", "class Ignored {}")

	result, err := RepositoryType(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsJavaRepository {
		t.Fatal("IsJavaRepository = false, want true")
	}
	if !result.ContainsJavaCode {
		t.Fatal("ContainsJavaCode = false, want true")
	}
	if result.JavaFileCount != 1 {
		t.Fatalf("JavaFileCount = %d, want 1", result.JavaFileCount)
	}
	wantMarkers := []string{"module/build.gradle.kts", "pom.xml"}
	if len(result.JavaProjectFiles) != len(wantMarkers) {
		t.Fatalf("JavaProjectFiles = %#v, want %#v", result.JavaProjectFiles, wantMarkers)
	}
	for i, marker := range wantMarkers {
		if result.JavaProjectFiles[i] != marker {
			t.Fatalf("JavaProjectFiles = %#v, want %#v", result.JavaProjectFiles, wantMarkers)
		}
	}
}

func TestRepositoryTypeDetectsJavaProjectWithoutSource(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "build.gradle", "plugins { id 'java' }")

	result, err := RepositoryType(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsJavaRepository {
		t.Fatal("IsJavaRepository = false, want true")
	}
	if result.ContainsJavaCode {
		t.Fatal("ContainsJavaCode = true, want false")
	}
}

func TestRepositoryTypeReportsNonJavaRepository(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "# Example")

	result, err := RepositoryType(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if result.IsJavaRepository {
		t.Fatal("IsJavaRepository = true, want false")
	}
	if result.ContainsJavaCode {
		t.Fatal("ContainsJavaCode = true, want false")
	}
	if result.JavaFileCount != 0 {
		t.Fatalf("JavaFileCount = %d, want 0", result.JavaFileCount)
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
