// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies SQLite index migration and compatibility behavior.
package index

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestOpenWritesSchemaVersionMetadata(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var version string
	if err := store.db.QueryRow("SELECT value FROM metadata WHERE key = 'schema_version'").Scan(&version); err != nil {
		t.Fatal(err)
	}
	want := strconv.Itoa(CurrentSchemaVersion)
	if version != want {
		t.Fatalf("schema_version = %q, want %q", version, want)
	}
}

func TestStatsIncludesSchemaVersion(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	stats, err := store.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.IndexSchemaVersion != CurrentSchemaVersion {
		t.Fatalf("IndexSchemaVersion = %d, want %d", stats.IndexSchemaVersion, CurrentSchemaVersion)
	}
}

func TestOpenExistingReportsMissingIndex(t *testing.T) {
	_, err := OpenExisting(t.TempDir())
	assertIndexError(t, err, MissingIndex)
}

func TestOpenExistingReportsStaleIndex(t *testing.T) {
	root := t.TempDir()
	path := DBPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE files (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = OpenExisting(root)
	assertIndexError(t, err, StaleIndex)
}

func TestOpenExistingReportsIncompatibleIndex(t *testing.T) {
	root := t.TempDir()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec("UPDATE metadata SET value = ? WHERE key = 'schema_version'", "999"); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = OpenExisting(root)
	assertIndexError(t, err, IncompatibleIndex)
}

func TestRebuildPopulatesIndexedAt(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	before := time.Now().UTC()
	files := []File{
		{
			Path:         "src/main/java/com/example/Service.java",
			Language:     "java",
			SizeBytes:    100,
			LastModified: time.Now().UTC().Add(-time.Hour),
			Hash:         "abc123",
		},
	}
	if err := store.Rebuild(files, nil); err != nil {
		t.Fatal(err)
	}
	after := time.Now().UTC()

	results, err := store.SearchAll("Service")
	if err != nil {
		t.Fatal(err)
	}
	var found *File
	for _, r := range results {
		if r.Type == "file" && r.File != nil {
			f := r.File
			found = f
			break
		}
	}
	if found == nil {
		t.Fatal("no file result returned for path substring match")
	}
	if found.IndexedAt.IsZero() {
		t.Fatal("IndexedAt is zero; expected Rebuild to populate it")
	}
	if found.IndexedAt.Before(before.Add(-time.Second)) || found.IndexedAt.After(after.Add(time.Second)) {
		t.Fatalf("IndexedAt = %v, want between %v and %v", found.IndexedAt, before, after)
	}
}

func TestRebuildAssignsSameIndexedAtToAllFiles(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	files := []File{
		{Path: "a/Alpha.java", Language: "java", SizeBytes: 1, LastModified: time.Now().UTC(), Hash: "h1"},
		{Path: "a/Beta.java", Language: "java", SizeBytes: 2, LastModified: time.Now().UTC(), Hash: "h2"},
		{Path: "a/Gamma.java", Language: "java", SizeBytes: 3, LastModified: time.Now().UTC(), Hash: "h3"},
	}
	if err := store.Rebuild(files, nil); err != nil {
		t.Fatal(err)
	}

	results, err := store.SearchAll(".java")
	if err != nil {
		t.Fatal(err)
	}
	var stamps []time.Time
	for _, r := range results {
		if r.Type == "file" && r.File != nil {
			stamps = append(stamps, r.File.IndexedAt)
		}
	}
	if len(stamps) != len(files) {
		t.Fatalf("got %d file results, want %d", len(stamps), len(files))
	}
	for i := 1; i < len(stamps); i++ {
		if !stamps[i].Equal(stamps[0]) {
			t.Fatalf("IndexedAt values diverge within one Rebuild: %v vs %v", stamps[0], stamps[i])
		}
	}
}

func assertIndexError(t *testing.T, err error, kind IndexErrorKind) {
	t.Helper()
	var indexErr IndexError
	if !errors.As(err, &indexErr) {
		t.Fatalf("err = %v, want IndexError", err)
	}
	if indexErr.Kind != kind {
		t.Fatalf("IndexError.Kind = %s, want %s", indexErr.Kind, kind)
	}
}
