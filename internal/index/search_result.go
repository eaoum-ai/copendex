// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file ranks and orders file and symbol search results.
package index

import (
	"sort"
	"strings"
)

type SearchResult struct {
	Type     string  `json:"type"`
	Path     string  `json:"path"`
	Name     string  `json:"name,omitempty"`
	Kind     string  `json:"kind,omitempty"`
	Language string  `json:"language,omitempty"`
	Line     int     `json:"line,omitempty"`
	Rank     int     `json:"rank"`
	File     *File   `json:"file,omitempty"`
	Symbol   *Symbol `json:"symbol,omitempty"`
}

func sortSymbols(query string, symbols []Symbol) {
	sort.SliceStable(symbols, func(i, j int) bool {
		left := rank(query, symbols[i].Name, false)
		right := rank(query, symbols[j].Name, false)
		if left != right {
			return left < right
		}
		if symbols[i].Name != symbols[j].Name {
			return symbols[i].Name < symbols[j].Name
		}
		return symbols[i].File < symbols[j].File
	})
}

func sortResults(results []SearchResult) {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Rank != results[j].Rank {
			return results[i].Rank < results[j].Rank
		}
		if results[i].Type != results[j].Type {
			return results[i].Type > results[j].Type
		}
		if results[i].Name != results[j].Name {
			return results[i].Name < results[j].Name
		}
		return results[i].Path < results[j].Path
	})
}

func rank(query, value string, filePath bool) int {
	q := strings.ToLower(query)
	v := strings.ToLower(value)
	switch {
	case v == q:
		return 1
	case strings.HasPrefix(v, q):
		return 2
	case strings.Contains(v, q) && !filePath:
		return 3
	case strings.Contains(v, q):
		return 4
	default:
		return 100
	}
}
