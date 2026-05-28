// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies user-facing CLI command behavior.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eaoum-ai/copendex/internal/config"
)

func TestIndexReportsExistingIndexWithoutRebuild(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	writeJavaFile(t, root)
	if err := config.EnsureDefault(root); err != nil {
		t.Fatal(err)
	}

	if err := runCommand("index"); err != nil {
		t.Fatal(err)
	}
	err := runCommand("index")
	if err == nil {
		t.Fatal("second index succeeded, want existing index error")
	}
	if !strings.Contains(err.Error(), "index is already built") {
		t.Fatalf("err = %v, want existing index guidance", err)
	}
	if !strings.Contains(err.Error(), "--rebuild or -r") {
		t.Fatalf("err = %v, want rebuild shorthand guidance", err)
	}
}

func TestIndexRebuildShorthand(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	writeJavaFile(t, root)
	if err := config.EnsureDefault(root); err != nil {
		t.Fatal(err)
	}

	if err := runCommand("index"); err != nil {
		t.Fatal(err)
	}
	if err := runCommand("index", "-r"); err != nil {
		t.Fatal(err)
	}
}

func TestDetectReportsJavaRepository(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	writeJavaFile(t, root)
	if err := os.WriteFile(filepath.Join(root, "pom.xml"), []byte("<project></project>\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCommandOutput("detect")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Java repository: true") {
		t.Fatalf("output = %q, want Java repository detection", out)
	}
	if !strings.Contains(out, "Contains Java source: true") {
		t.Fatalf("output = %q, want Java source detection", out)
	}
}

func TestDetectJSONReportsNonJavaRepository(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Example\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCommandOutput("detect", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		IsJavaRepository bool `json:"isJavaRepository"`
		ContainsJavaCode bool `json:"containsJavaCode"`
		JavaFileCount    int  `json:"javaFileCount"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	if result.IsJavaRepository || result.ContainsJavaCode || result.JavaFileCount != 0 {
		t.Fatalf("result = %#v, want non-Java repository", result)
	}
}

func runCommand(args ...string) error {
	_, err := runCommandOutput(args...)
	return err
}

func runCommandOutput(args ...string) (string, error) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func writeJavaFile(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "Example.java")
	if err := os.WriteFile(path, []byte("package com.example;\nclass Example {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	})
}
