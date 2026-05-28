// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package files discovers source files while honoring config and ignore rules.
package files

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eaoum-ai/cosha/internal/config"
	"github.com/eaoum-ai/cosha/internal/index"
)

type Scanner struct {
	root          string
	cfg           config.Config
	gitignore     []ignorePattern
	languageAllow map[string]bool
}

func NewScanner(root string, cfg config.Config) (*Scanner, error) {
	patterns, err := readGitignore(root)
	if err != nil {
		return nil, err
	}
	allow := map[string]bool{}
	for _, language := range cfg.Index.Languages {
		allow[strings.ToLower(language)] = true
	}
	return &Scanner{root: root, cfg: cfg, gitignore: patterns, languageAllow: allow}, nil
}

func (s *Scanner) Scan() ([]index.File, error) {
	var out []index.File
	err := filepath.WalkDir(s.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == s.root {
			return nil
		}
		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if s.excluded(rel + "/") {
				return filepath.SkipDir
			}
			return nil
		}
		if s.excluded(rel) || !s.included(rel) {
			return nil
		}
		language := languageForPath(rel)
		if language == "" || (len(s.languageAllow) > 0 && !s.languageAllow[language]) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		hash, err := hashFile(path)
		if err != nil {
			return err
		}
		out = append(out, index.File{
			Path:         rel,
			Language:     language,
			SizeBytes:    info.Size(),
			LastModified: info.ModTime(),
			Hash:         hash,
		})
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

func (s *Scanner) included(rel string) bool {
	if len(s.cfg.Include) == 0 {
		return true
	}
	for _, pattern := range s.cfg.Include {
		if globMatch(pattern, rel) {
			return true
		}
	}
	return false
}

func (s *Scanner) excluded(rel string) bool {
	for _, pattern := range s.cfg.Exclude {
		if globMatch(pattern, rel) {
			return true
		}
	}
	for _, pattern := range s.gitignore {
		if pattern.matches(rel) {
			return true
		}
	}
	return false
}

func languageForPath(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".java") {
		return "java"
	}
	return ""
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type ignorePattern struct {
	raw       string
	directory bool
	anchored  bool
}

func readGitignore(root string) ([]ignorePattern, error) {
	file, err := os.Open(filepath.Join(root, ".gitignore"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var patterns []ignorePattern
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		p := ignorePattern{raw: filepath.ToSlash(line)}
		if strings.HasPrefix(p.raw, "/") {
			p.anchored = true
			p.raw = strings.TrimPrefix(p.raw, "/")
		}
		if strings.HasSuffix(p.raw, "/") {
			p.directory = true
			p.raw = strings.TrimSuffix(p.raw, "/")
		}
		patterns = append(patterns, p)
	}
	return patterns, nil
}

func (p ignorePattern) matches(rel string) bool {
	rel = strings.TrimSuffix(rel, "/")
	if p.raw == "" {
		return false
	}
	if p.anchored {
		return rel == p.raw || strings.HasPrefix(rel, p.raw+"/") || globMatch(p.raw, rel)
	}
	if strings.Contains(p.raw, "/") {
		return rel == p.raw || strings.HasPrefix(rel, p.raw+"/") || globMatch(p.raw, rel)
	}
	for _, part := range strings.Split(rel, "/") {
		if part == p.raw || globMatch(p.raw, part) {
			return true
		}
	}
	return false
}

func globMatch(pattern, rel string) bool {
	pattern = filepath.ToSlash(strings.TrimPrefix(pattern, "./"))
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if pattern == rel {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return rel == prefix || strings.HasPrefix(rel, prefix+"/")
	}
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		if ok, _ := filepath.Match(suffix, rel); ok {
			return true
		}
		if ok, _ := filepath.Match(suffix, filepath.Base(rel)); ok {
			return true
		}
		if strings.HasPrefix(suffix, "*") {
			return strings.HasSuffix(rel, strings.TrimPrefix(suffix, "*"))
		}
		return strings.HasSuffix(rel, "/"+suffix)
	}
	if strings.Contains(pattern, "**/") {
		parts := strings.Split(pattern, "**/")
		if len(parts) != 2 || !strings.HasPrefix(rel, parts[0]) {
			return false
		}
		tail := parts[1]
		if ok, _ := filepath.Match(tail, filepath.Base(rel)); ok {
			return true
		}
		return strings.HasPrefix(tail, "*") && strings.HasSuffix(rel, strings.TrimPrefix(tail, "*"))
	}
	ok, _ := filepath.Match(pattern, rel)
	return ok
}
