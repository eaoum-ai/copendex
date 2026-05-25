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
	"testing"
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
	if version != "1" {
		t.Fatalf("schema_version = %q, want 1", version)
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
