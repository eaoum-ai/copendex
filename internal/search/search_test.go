// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies symbol, file, and filtered search behavior.
package search

import (
	"testing"
	"time"

	idx "github.com/eaoum-ai/cosha/internal/index"
)

func TestSearchSymbolsAndFiles(t *testing.T) {
	root := t.TempDir()
	store, err := idx.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	files := []idx.File{{
		Path:         "src/main/java/com/example/AuthorizationService.java",
		Language:     "java",
		SizeBytes:    123,
		LastModified: time.Now(),
		Hash:         "abc",
	}}
	symbols := map[string][]idx.Symbol{
		files[0].Path: {{
			Name:        "AuthorizationService",
			Kind:        "class",
			Language:    "java",
			File:        files[0].Path,
			PackageName: "com.example",
			Line:        12,
			Annotations: []string{"Service"},
		}, {
			Name:        "authorize",
			Kind:        "method",
			Language:    "java",
			File:        files[0].Path,
			PackageName: "com.example",
			Line:        20,
		}},
	}
	if err := store.Rebuild(files, symbols); err != nil {
		t.Fatal(err)
	}
	service := New(store)
	syms, err := service.Symbols("Service")
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) != 1 || syms[0].Name != "AuthorizationService" {
		t.Fatalf("symbols = %#v", syms)
	}
	results, err := service.All("Authorization")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) < 2 {
		t.Fatalf("results = %#v, want symbol and file result", results)
	}
	if results[0].Type != "symbol" {
		t.Fatalf("first result type = %s, want symbol", results[0].Type)
	}
	if results[0].Symbol == nil || results[0].Symbol.Name != results[0].Name {
		t.Fatalf("symbol payload = %#v, wrapper = %#v", results[0].Symbol, results[0])
	}
}

func TestSearchFilters(t *testing.T) {
	root := t.TempDir()
	store, err := idx.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	files := []idx.File{{
		Path:         "src/main/java/com/example/AuthorizationService.java",
		Language:     "java",
		SizeBytes:    123,
		LastModified: time.Now(),
		Hash:         "abc",
	}, {
		Path:         "src/test/java/com/example/AuthorizationServiceTest.java",
		Language:     "java",
		SizeBytes:    456,
		LastModified: time.Now(),
		Hash:         "def",
	}, {
		Path:         "src/main/java/com/example/AuthorizationRepository.java",
		Language:     "java",
		SizeBytes:    789,
		LastModified: time.Now(),
		Hash:         "ghi",
	}}
	symbols := map[string][]idx.Symbol{
		files[0].Path: {{
			Name:        "AuthorizationService",
			Kind:        "class",
			Language:    "java",
			File:        files[0].Path,
			PackageName: "com.example",
			Line:        12,
		}},
		files[1].Path: {{
			Name:        "AuthorizationServiceTest",
			Kind:        "class",
			Language:    "java",
			File:        files[1].Path,
			PackageName: "com.example.test",
			Line:        9,
		}},
		files[2].Path: {{
			Name:        "AuthorizationRepository",
			Kind:        "interface",
			Language:    "java",
			File:        files[2].Path,
			PackageName: "com.example",
			Line:        7,
		}},
	}
	if err := store.Rebuild(files, symbols); err != nil {
		t.Fatal(err)
	}

	service := New(store)
	syms, err := service.SymbolsFiltered("Authorization", idx.QueryFilters{Path: "/test/", PackageName: ".test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) != 1 || syms[0].Name != "AuthorizationServiceTest" {
		t.Fatalf("filtered symbols = %#v", syms)
	}

	results, err := service.AllFiltered("Authorization", idx.QueryFilters{Kind: "class", PackageName: ".test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Type != "symbol" {
		t.Fatalf("filtered results = %#v, want one symbol result", results)
	}

	syms, err = service.SymbolsFiltered("Authorization", idx.QueryFilters{Kind: "class, interface"})
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) != 3 {
		t.Fatalf("multi-kind symbols = %#v, want class and interface symbols", syms)
	}

	results, err = service.AllFiltered("Authorization", idx.QueryFilters{Kind: "class,interface", Path: "src/main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("multi-kind results = %#v, want class and interface symbols from main package", results)
	}
}
