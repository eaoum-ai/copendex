// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies user-facing CLI command behavior.
package cli

import (
	"bytes"
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

func runCommand(args ...string) error {
	cmd := NewRootCommand()
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	return cmd.Execute()
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
