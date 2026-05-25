// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package config loads and writes per-repository Copendex configuration.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirName        = ".copendex"
	ConfigFileName = "config.yaml"
)

type Config struct {
	Version int
	Include []string
	Exclude []string
	Index   IndexConfig
	Output  OutputConfig
}

type IndexConfig struct {
	Languages []string
}

type OutputConfig struct {
	DefaultFormat string
}

func Default() Config {
	return Config{
		Version: 1,
		Include: []string{"src/**/*.java", "**/*.java"},
		Exclude: []string{
			"build/**",
			"target/**",
			".git/**",
			".copendex/**",
			"node_modules/**",
		},
		Index:  IndexConfig{Languages: []string{"java"}},
		Output: OutputConfig{DefaultFormat: "text"},
	}
}

func ConfigPath(root string) string {
	return filepath.Join(root, DirName, ConfigFileName)
}

func EnsureDefault(root string) error {
	dir := filepath.Join(root, DirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := ConfigPath(root)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(path, []byte(DefaultYAML()), 0o644)
}

func DefaultYAML() string {
	return `version: 1
include:
  - "src/**/*.java"
  - "**/*.java"
exclude:
  - "build/**"
  - "target/**"
  - ".git/**"
  - ".copendex/**"
  - "node_modules/**"
index:
  languages:
    - java
output:
  defaultFormat: text
`
}

func Load(root string) (Config, error) {
	cfg := Default()
	path := ConfigPath(root)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var section string
	var subsection string
	var listKey string
	for scanner.Scan() {
		raw := strings.TrimSpace(stripComment(scanner.Text()))
		if raw == "" {
			continue
		}
		if strings.HasSuffix(raw, ":") && !strings.HasPrefix(raw, "-") {
			key := strings.TrimSuffix(raw, ":")
			if key == "include" || key == "exclude" {
				section = ""
				subsection = ""
				listKey = key
				if key == "include" {
					cfg.Include = nil
				} else {
					cfg.Exclude = nil
				}
				continue
			}
			section = key
			subsection = ""
			listKey = ""
			continue
		}
		if strings.HasPrefix(raw, "-") {
			value := unquote(strings.TrimSpace(strings.TrimPrefix(raw, "-")))
			switch listKey {
			case "include":
				cfg.Include = append(cfg.Include, value)
			case "exclude":
				cfg.Exclude = append(cfg.Exclude, value)
			case "index.languages":
				cfg.Index.Languages = append(cfg.Index.Languages, value)
			}
			continue
		}
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return cfg, fmt.Errorf("invalid config line: %s", raw)
		}
		key := strings.TrimSpace(parts[0])
		value := unquote(strings.TrimSpace(parts[1]))
		switch {
		case section == "" && key == "version":
			fmt.Sscanf(value, "%d", &cfg.Version)
		case section == "index" && key == "languages" && value == "":
			subsection = "languages"
			listKey = "index.languages"
			cfg.Index.Languages = nil
		case section == "output" && key == "defaultFormat":
			cfg.Output.DefaultFormat = value
		default:
			if section == "index" && subsection == "languages" {
				listKey = "index.languages"
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func stripComment(s string) string {
	if i := strings.Index(s, "#"); i >= 0 {
		return s[:i]
	}
	return s
}

func unquote(s string) string {
	return strings.Trim(s, `"'`)
}
