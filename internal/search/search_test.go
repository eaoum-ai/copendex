package search

import (
	"testing"
	"time"

	idx "github.com/eaoum-ai/copendex/internal/index"
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
