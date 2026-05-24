package index

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

func DBPath(root string) string {
	return filepath.Join(root, ".copendex", "index", "copendex.db")
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

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate() error {
	_, err := s.db.Exec(`
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
`)
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

func (s *Store) SearchSymbols(query string) ([]Symbol, error) {
	return s.querySymbols(query)
}

func (s *Store) SearchAll(query string) ([]SearchResult, error) {
	var results []SearchResult
	symbols, err := s.querySymbols(query)
	if err != nil {
		return nil, err
	}
	for _, sym := range symbols {
		symbol := sym
		results = append(results, SearchResult{Type: "symbol", Path: sym.File, Name: sym.Name, Kind: sym.Kind, Language: sym.Language, Line: sym.Line, Rank: rank(query, sym.Name, false), Symbol: &symbol})
	}
	fileRows, err := s.db.Query(`SELECT id, path, language, size_bytes, modified_at, hash FROM files WHERE lower(path) LIKE '%' || lower(?) || '%'`, query)
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
	sortResults(results)
	return results, nil
}

func (s *Store) querySymbols(query string) ([]Symbol, error) {
	rows, err := s.db.Query(`
SELECT symbols.id, symbols.file_id, symbols.name, symbols.kind, symbols.language, files.path, symbols.package_name, symbols.line, symbols.annotations_json
FROM symbols
JOIN files ON files.id = symbols.file_id
WHERE lower(symbols.name) LIKE '%' || lower(?) || '%'
ORDER BY symbols.name, files.path
`, query)
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
