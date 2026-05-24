package index

import "time"

type File struct {
	ID           int64     `json:"id,omitempty"`
	Path         string    `json:"path"`
	Language     string    `json:"language"`
	SizeBytes    int64     `json:"sizeBytes"`
	LastModified time.Time `json:"lastModified"`
	Hash         string    `json:"hash"`
}

type Symbol struct {
	ID          int64    `json:"id,omitempty"`
	FileID      int64    `json:"fileId,omitempty"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Language    string   `json:"language"`
	File        string   `json:"file"`
	PackageName string   `json:"package,omitempty"`
	Line        int      `json:"line"`
	Annotations []string `json:"annotations,omitempty"`
}

type Stats struct {
	FileCount     int64            `json:"fileCount"`
	SymbolCount   int64            `json:"symbolCount"`
	LanguageCount int64            `json:"languageCount"`
	IndexSize     int64            `json:"indexSizeBytes"`
	Languages     map[string]int64 `json:"languages"`
}
