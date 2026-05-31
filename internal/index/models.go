// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file defines index records and JSON-facing data models.
package index

import "time"

type File struct {
	ID           int64     `json:"id,omitempty"`
	Path         string    `json:"path"`
	Language     string    `json:"language"`
	SizeBytes    int64     `json:"sizeBytes"`
	LastModified time.Time `json:"lastModified"`
	Hash         string    `json:"hash"`
	IndexedAt    time.Time `json:"indexedAt"`
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
	FileCount          int64            `json:"fileCount"`
	SymbolCount        int64            `json:"symbolCount"`
	LanguageCount      int64            `json:"languageCount"`
	IndexSize          int64            `json:"indexSizeBytes"`
	IndexSchemaVersion int              `json:"indexSchemaVersion"`
	Languages          map[string]int64 `json:"languages"`
}

type QueryFilters struct {
	Kind        string
	Language    string
	Path        string
	PackageName string
}
