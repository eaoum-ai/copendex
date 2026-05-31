// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package index stores discovered files and symbols in a local SQLite index.
package index

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

const CurrentSchemaVersion = 2

type IndexErrorKind string

const (
	MissingIndex      IndexErrorKind = "missing"
	StaleIndex        IndexErrorKind = "stale"
	IncompatibleIndex IndexErrorKind = "incompatible"
)

type IndexError struct {
	Kind    IndexErrorKind
	Path    string
	Version int
}

func (e IndexError) Error() string {
	switch e.Kind {
	case MissingIndex:
		return fmt.Sprintf("missing Cosha index at %s; run `cosha index`", e.Path)
	case StaleIndex:
		return fmt.Sprintf("stale Cosha index at %s: missing schema version metadata; run `cosha index --rebuild`", e.Path)
	case IncompatibleIndex:
		return fmt.Sprintf("incompatible Cosha index at %s: schema version %d, expected %d; run `cosha index --rebuild`", e.Path, e.Version, CurrentSchemaVersion)
	default:
		return fmt.Sprintf("invalid Cosha index at %s", e.Path)
	}
}

func DBPath(root string) string {
	return filepath.Join(root, ".cosha", "index", "cosha.db")
}

func Open(root string) (*Store, error) {
	path := DBPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db, path: path}
	if err := store.Migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func OpenExisting(root string) (*Store, error) {
	path := DBPath(root)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, IndexError{Kind: MissingIndex, Path: path}
		}
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db, path: path}
	if err := store.CheckCompatible(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate() error {
	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS metadata (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS files (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL UNIQUE,
	language TEXT NOT NULL,
	size_bytes INTEGER NOT NULL,
	modified_at TEXT NOT NULL,
	hash TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS symbols (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	file_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	language TEXT NOT NULL,
	package_name TEXT,
	line INTEGER NOT NULL,
	annotations_json TEXT NOT NULL DEFAULT '[]',
	FOREIGN KEY(file_id) REFERENCES files(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
CREATE INDEX IF NOT EXISTS idx_symbols_file_id ON symbols(file_id);
CREATE VIRTUAL TABLE IF NOT EXISTS symbols_fts USING fts5(
	name, kind, package_name,
	content='symbols',
	content_rowid='id',
	tokenize="unicode61 tokenchars '$_'"
);
CREATE VIRTUAL TABLE IF NOT EXISTS files_fts USING fts5(
	path,
	content='files',
	content_rowid='id',
	tokenize="unicode61 tokenchars '$_'"
);
CREATE TRIGGER IF NOT EXISTS symbols_ai AFTER INSERT ON symbols BEGIN
	INSERT INTO symbols_fts(rowid, name, kind, package_name)
	VALUES (new.id, new.name, new.kind, COALESCE(new.package_name, ''));
END;
CREATE TRIGGER IF NOT EXISTS symbols_ad AFTER DELETE ON symbols BEGIN
	INSERT INTO symbols_fts(symbols_fts, rowid, name, kind, package_name)
	VALUES ('delete', old.id, old.name, old.kind, COALESCE(old.package_name, ''));
END;
CREATE TRIGGER IF NOT EXISTS symbols_au AFTER UPDATE ON symbols BEGIN
	INSERT INTO symbols_fts(symbols_fts, rowid, name, kind, package_name)
	VALUES ('delete', old.id, old.name, old.kind, COALESCE(old.package_name, ''));
	INSERT INTO symbols_fts(rowid, name, kind, package_name)
	VALUES (new.id, new.name, new.kind, COALESCE(new.package_name, ''));
END;
CREATE TRIGGER IF NOT EXISTS files_ai AFTER INSERT ON files BEGIN
	INSERT INTO files_fts(rowid, path) VALUES (new.id, new.path);
END;
CREATE TRIGGER IF NOT EXISTS files_ad AFTER DELETE ON files BEGIN
	INSERT INTO files_fts(files_fts, rowid, path) VALUES ('delete', old.id, old.path);
END;
CREATE TRIGGER IF NOT EXISTS files_au AFTER UPDATE ON files BEGIN
	INSERT INTO files_fts(files_fts, rowid, path) VALUES ('delete', old.id, old.path);
	INSERT INTO files_fts(rowid, path) VALUES (new.id, new.path);
END;
`); err != nil {
		return err
	}
	_, err := s.db.Exec(`
INSERT INTO metadata(key, value)
VALUES ('schema_version', ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
`, strconv.Itoa(CurrentSchemaVersion))
	return err
}

func (s *Store) Rebuild(files []File, symbolsByPath map[string][]Symbol) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM symbols"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM files"); err != nil {
		return err
	}
	fileStmt, err := tx.Prepare("INSERT INTO files(path, language, size_bytes, modified_at, hash) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer fileStmt.Close()
	symbolStmt, err := tx.Prepare("INSERT INTO symbols(file_id, name, kind, language, package_name, line, annotations_json) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer symbolStmt.Close()
	for _, file := range files {
		res, err := fileStmt.Exec(file.Path, file.Language, file.SizeBytes, file.LastModified.UTC().Format(time.RFC3339Nano), file.Hash)
		if err != nil {
			return err
		}
		fileID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		for _, sym := range symbolsByPath[file.Path] {
			ann, err := json.Marshal(sym.Annotations)
			if err != nil {
				return err
			}
			if _, err := symbolStmt.Exec(fileID, sym.Name, sym.Kind, sym.Language, sym.PackageName, sym.Line, string(ann)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *Store) Stats() (Stats, error) {
	var stats Stats
	stats.Languages = map[string]int64{}
	schemaVersion, err := s.SchemaVersion()
	if err != nil {
		return stats, err
	}
	stats.IndexSchemaVersion = schemaVersion
	if err := s.db.QueryRow("SELECT COUNT(*) FROM files").Scan(&stats.FileCount); err != nil {
		return stats, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&stats.SymbolCount); err != nil {
		return stats, err
	}
	if err := s.db.QueryRow("SELECT COUNT(DISTINCT language) FROM files").Scan(&stats.LanguageCount); err != nil {
		return stats, err
	}
	rows, err := s.db.Query("SELECT language, COUNT(*) FROM files GROUP BY language ORDER BY language")
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var language string
		var count int64
		if err := rows.Scan(&language, &count); err != nil {
			return stats, err
		}
		stats.Languages[language] = count
	}
	if info, err := os.Stat(s.path); err == nil {
		stats.IndexSize = info.Size()
	}
	return stats, rows.Err()
}

func (s *Store) SchemaVersion() (int, error) {
	var version string
	if err := s.db.QueryRow("SELECT value FROM metadata WHERE key = 'schema_version'").Scan(&version); err != nil {
		if errors.Is(err, sql.ErrNoRows) || isMissingMetadataError(err) {
			return 0, IndexError{Kind: StaleIndex, Path: s.path}
		}
		return 0, err
	}
	parsed, err := strconv.Atoi(version)
	if err != nil {
		return 0, IndexError{Kind: IncompatibleIndex, Path: s.path}
	}
	return parsed, nil
}

func (s *Store) CheckCompatible() error {
	version, err := s.SchemaVersion()
	if err != nil {
		return err
	}
	if version != CurrentSchemaVersion {
		return IndexError{Kind: IncompatibleIndex, Path: s.path, Version: version}
	}
	return nil
}

func isMissingMetadataError(err error) bool {
	return strings.Contains(err.Error(), "no such table: metadata")
}

func (s *Store) SearchSymbols(query string) ([]Symbol, error) {
	return s.SearchSymbolsFiltered(query, QueryFilters{})
}

func (s *Store) SearchSymbolsFiltered(query string, filters QueryFilters) ([]Symbol, error) {
	return s.querySymbols(query, filters)
}

func (s *Store) SearchAll(query string) ([]SearchResult, error) {
	return s.SearchAllFiltered(query, QueryFilters{})
}

func (s *Store) SearchAllFiltered(query string, filters QueryFilters) ([]SearchResult, error) {
	var results []SearchResult
	symbols, err := s.querySymbols(query, filters)
	if err != nil {
		return nil, err
	}
	for _, sym := range symbols {
		symbol := sym
		results = append(results, SearchResult{Type: "symbol", Path: sym.File, Name: sym.Name, Kind: sym.Kind, Language: sym.Language, Line: sym.Line, Rank: rank(query, sym.Name, false), Symbol: &symbol})
	}
	if filters.Kind == "" && filters.PackageName == "" {
		fileRows, err := s.queryFiles(query, filters)
		if err != nil {
			return nil, err
		}
		defer fileRows.Close()
		for fileRows.Next() {
			file, err := scanFile(fileRows)
			if err != nil {
				return nil, err
			}
			results = append(results, SearchResult{Type: "file", Path: file.Path, Language: file.Language, Rank: rank(query, file.Path, true), File: &file})
		}
		if err := fileRows.Err(); err != nil {
			return nil, err
		}
	}
	sortResults(results)
	return results, nil
}

func (s *Store) querySymbols(query string, filters QueryFilters) ([]Symbol, error) {
	matchExpr := ftsMatchExpr(query)
	var sqlQuery string
	var args []any
	if matchExpr == "" {
		sqlQuery = `
SELECT symbols.id, symbols.file_id, symbols.name, symbols.kind, symbols.language, files.path, symbols.package_name, symbols.line, symbols.annotations_json
FROM symbols
JOIN files ON files.id = symbols.file_id
WHERE 1=1
`
	} else {
		sqlQuery = `
SELECT symbols.id, symbols.file_id, symbols.name, symbols.kind, symbols.language, files.path, symbols.package_name, symbols.line, symbols.annotations_json
FROM symbols_fts
JOIN symbols ON symbols.id = symbols_fts.rowid
JOIN files ON files.id = symbols.file_id
WHERE symbols_fts MATCH ?
`
		args = append(args, matchExpr)
	}
	sqlQuery, args = appendFilters(sqlQuery, args, filters, true)
	if matchExpr == "" {
		sqlQuery += " ORDER BY symbols.name, files.path"
	} else {
		sqlQuery += " ORDER BY symbols_fts.rank, symbols.name, files.path"
	}
	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var symbols []Symbol
	for rows.Next() {
		var sym Symbol
		var annotations string
		if err := rows.Scan(&sym.ID, &sym.FileID, &sym.Name, &sym.Kind, &sym.Language, &sym.File, &sym.PackageName, &sym.Line, &annotations); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(annotations), &sym.Annotations)
		symbols = append(symbols, sym)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortSymbols(query, symbols)
	return symbols, nil
}

func (s *Store) queryFiles(query string, filters QueryFilters) (*sql.Rows, error) {
	matchExpr := ftsMatchExpr(query)
	var sqlQuery string
	var args []any
	if matchExpr == "" {
		sqlQuery = `
SELECT files.id, files.path, files.language, files.size_bytes, files.modified_at, files.hash
FROM files
WHERE 1=1
`
	} else {
		sqlQuery = `
SELECT files.id, files.path, files.language, files.size_bytes, files.modified_at, files.hash
FROM files_fts
JOIN files ON files.id = files_fts.rowid
WHERE files_fts MATCH ?
`
		args = append(args, matchExpr)
	}
	sqlQuery, args = appendFilters(sqlQuery, args, filters, false)
	if matchExpr == "" {
		sqlQuery += " ORDER BY files.path"
	} else {
		sqlQuery += " ORDER BY files_fts.rank, files.path"
	}
	return s.db.Query(sqlQuery, args...)
}

func ftsMatchExpr(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	escaped := strings.ReplaceAll(raw, `"`, `""`)
	return `"` + escaped + `"` + "*"
}

func appendFilters(sqlQuery string, args []any, filters QueryFilters, symbols bool) (string, []any) {
	if filters.Kind != "" && symbols {
		kinds := splitFilterValues(filters.Kind)
		if len(kinds) == 1 {
			sqlQuery += " AND lower(symbols.kind) = lower(?)"
			args = append(args, kinds[0])
		} else if len(kinds) > 1 {
			sqlQuery += " AND lower(symbols.kind) IN (" + placeholders(len(kinds)) + ")"
			for _, kind := range kinds {
				args = append(args, strings.ToLower(kind))
			}
		}
	}
	if filters.Language != "" {
		if symbols {
			sqlQuery += " AND lower(symbols.language) = lower(?)"
		} else {
			sqlQuery += " AND lower(files.language) = lower(?)"
		}
		args = append(args, filters.Language)
	}
	if filters.Path != "" {
		sqlQuery += " AND lower(files.path) LIKE '%' || lower(?) || '%'"
		args = append(args, filters.Path)
	}
	if filters.PackageName != "" && symbols {
		sqlQuery += " AND lower(symbols.package_name) LIKE '%' || lower(?) || '%'"
		args = append(args, filters.PackageName)
	}
	return sqlQuery, args
}

func splitFilterValues(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func placeholders(count int) string {
	values := make([]string, count)
	for i := range values {
		values[i] = "?"
	}
	return strings.Join(values, ", ")
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFile(row scanner) (File, error) {
	var file File
	var modified string
	if err := row.Scan(&file.ID, &file.Path, &file.Language, &file.SizeBytes, &modified, &file.Hash); err != nil {
		return file, err
	}
	t, err := time.Parse(time.RFC3339Nano, modified)
	if err != nil {
		return file, err
	}
	file.LastModified = t
	return file, nil
}
