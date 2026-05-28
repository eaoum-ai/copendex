// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package detect classifies repositories from local files and project markers.
package detect

import (
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/eaoum-ai/cosha/internal/config"
	"github.com/eaoum-ai/cosha/internal/files"
)

type Repository struct {
	IsJavaRepository  bool           `json:"isJavaRepository"`
	ContainsJavaCode  bool           `json:"containsJavaCode"`
	JavaFileCount     int            `json:"javaFileCount"`
	JavaProjectFiles  []string       `json:"javaProjectFiles"`
	IndexedFileCount  int            `json:"indexedFileCount"`
	LanguageFileCount map[string]int `json:"languageFileCount"`
	Languages         []string       `json:"languages"`
}

var javaProjectMarkers = map[string]bool{
	"pom.xml":             true,
	"build.gradle":        true,
	"build.gradle.kts":    true,
	"settings.gradle":     true,
	"settings.gradle.kts": true,
	"gradlew":             true,
	"gradlew.bat":         true,
	"mvnw":                true,
	"mvnw.cmd":            true,
}

func RepositoryType(root string, cfg config.Config) (Repository, error) {
	scanner, err := files.NewScanner(root, cfg)
	if err != nil {
		return Repository{}, err
	}
	indexedFiles, err := scanner.Scan()
	if err != nil {
		return Repository{}, err
	}
	languageFileCount := map[string]int{}
	for _, file := range indexedFiles {
		languageFileCount[file.Language]++
	}
	languages := make([]string, 0, len(languageFileCount))
	for language := range languageFileCount {
		languages = append(languages, language)
	}
	sort.Strings(languages)

	projectFiles, err := javaProjectFiles(root)
	if err != nil {
		return Repository{}, err
	}
	javaFileCount := languageFileCount["java"]
	return Repository{
		IsJavaRepository:  javaFileCount > 0 || len(projectFiles) > 0,
		ContainsJavaCode:  javaFileCount > 0,
		JavaFileCount:     javaFileCount,
		JavaProjectFiles:  projectFiles,
		IndexedFileCount:  len(indexedFiles),
		LanguageFileCount: languageFileCount,
		Languages:         languages,
	}, nil
}

func javaProjectFiles(root string) ([]string, error) {
	out := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() && skippedMarkerDir(name) {
			return filepath.SkipDir
		}
		if entry.IsDir() || !javaProjectMarkers[name] {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

func skippedMarkerDir(name string) bool {
	switch name {
	case ".git", ".cosha", ".cache", "build", "target", "node_modules":
		return true
	default:
		return false
	}
}
