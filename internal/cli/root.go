// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package cli wires Cosha commands, flags, and command output.
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eaoum-ai/cosha/internal/config"
	"github.com/eaoum-ai/cosha/internal/detect"
	"github.com/eaoum-ai/cosha/internal/files"
	idx "github.com/eaoum-ai/cosha/internal/index"
	"github.com/eaoum-ai/cosha/internal/lang/java"
	"github.com/eaoum-ai/cosha/internal/output"
	"github.com/eaoum-ai/cosha/internal/search"
	"github.com/eaoum-ai/cosha/internal/ui"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cosha",
		Short: "Local-first codebase intelligence for coding agents",
	}
	cmd.AddCommand(newInitCommand(), newDetectCommand(), newIndexCommand(), newSearchCommand(), newSymbolsCommand(), newStatsCommand(), newUICommand())
	return cmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Cosha in the current repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := config.EnsureDefault(root); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized %s\n", filepath.Join(config.DirName, config.ConfigFileName))
			return nil
		},
	}
}

func newIndexCommand() *cobra.Command {
	var rebuild bool
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index the current repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if !rebuild {
				existing, err := idx.OpenExisting(root)
				if err == nil {
					existing.Close()
					return fmt.Errorf("Cosha index is already built at %s; use --rebuild or -r to force rebuild the index", idx.DBPath(root))
				}
				var indexErr idx.IndexError
				if !errors.As(err, &indexErr) || indexErr.Kind != idx.MissingIndex {
					return err
				}
			}
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			scanner, err := files.NewScanner(root, cfg)
			if err != nil {
				return err
			}
			discovered, err := scanner.Scan()
			if err != nil {
				return err
			}
			if len(discovered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Java source files found for the current Cosha config")
			}
			symbolsByPath := map[string][]idx.Symbol{}
			for _, file := range discovered {
				if file.Language != "java" {
					continue
				}
				content, err := os.ReadFile(filepath.Join(root, file.Path))
				if err != nil {
					return err
				}
				symbolsByPath[file.Path] = java.Extract(file.Path, content)
			}
			if rebuild {
				if err := os.Remove(idx.DBPath(root)); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
			store, err := idx.Open(root)
			if err != nil {
				return err
			}
			defer store.Close()
			if err := store.Rebuild(discovered, symbolsByPath); err != nil {
				return err
			}
			var symbolCount int
			for _, symbols := range symbolsByPath {
				symbolCount += len(symbols)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Indexed %d files and %d symbols\n", len(discovered), symbolCount)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&rebuild, "rebuild", "r", false, "remove and recreate the local index before indexing")
	return cmd
}

func newDetectCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect repository languages and Java project markers",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			result, err := detect.RepositoryType(root, cfg)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(cmd.OutOrStdout(), result)
			}
			lines := []string{
				fmt.Sprintf("Java repository: %t", result.IsJavaRepository),
				fmt.Sprintf("Contains Java source: %t", result.ContainsJavaCode),
				fmt.Sprintf("Java source files: %d", result.JavaFileCount),
			}
			if len(result.JavaProjectFiles) > 0 {
				lines = append(lines, "Java project files:")
				for _, path := range result.JavaProjectFiles {
					lines = append(lines, fmt.Sprintf("  %s", path))
				}
			}
			return output.Lines(cmd.OutOrStdout(), lines)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "write structured JSON output")
	return cmd
}

func newSearchCommand() *cobra.Command {
	var jsonOut bool
	var filters idx.QueryFilters
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search indexed files and symbols",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()
			results, err := search.New(store).AllFiltered(args[0], filters)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(cmd.OutOrStdout(), results)
			}
			lines := make([]string, 0, len(results))
			for _, result := range results {
				if result.Type == "symbol" {
					lines = append(lines, fmt.Sprintf("%s %-10s %s:%d %s", result.Type, result.Kind, result.Path, result.Line, result.Name))
				} else {
					lines = append(lines, fmt.Sprintf("%s %-10s %s", result.Type, result.Language, result.Path))
				}
			}
			return output.Lines(cmd.OutOrStdout(), lines)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "write structured JSON output")
	addQueryFilterFlags(cmd, &filters)
	return cmd
}

func newSymbolsCommand() *cobra.Command {
	var jsonOut bool
	var filters idx.QueryFilters
	cmd := &cobra.Command{
		Use:   "symbols <query>",
		Short: "Search indexed symbols",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()
			symbols, err := search.New(store).SymbolsFiltered(args[0], filters)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(cmd.OutOrStdout(), symbols)
			}
			lines := make([]string, 0, len(symbols))
			for _, sym := range symbols {
				lines = append(lines, fmt.Sprintf("%-10s %s:%d %s", sym.Kind, sym.File, sym.Line, sym.Name))
			}
			return output.Lines(cmd.OutOrStdout(), lines)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "write structured JSON output")
	addQueryFilterFlags(cmd, &filters)
	return cmd
}

func addQueryFilterFlags(cmd *cobra.Command, filters *idx.QueryFilters) {
	cmd.Flags().StringVar(&filters.Kind, "kind", "", "filter symbols by kind, or comma-separated kinds")
	cmd.Flags().StringVar(&filters.Language, "language", "", "filter by language")
	cmd.Flags().StringVar(&filters.Path, "path", "", "filter by file path substring")
	cmd.Flags().StringVar(&filters.PackageName, "package", "", "filter symbols by package substring")
}

func newStatsCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show index statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()
			stats, err := store.Stats()
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(cmd.OutOrStdout(), stats)
			}
			lines := []string{
				fmt.Sprintf("Files: %d", stats.FileCount),
				fmt.Sprintf("Symbols: %d", stats.SymbolCount),
				fmt.Sprintf("Languages: %d", stats.LanguageCount),
				fmt.Sprintf("Index size: %d bytes", stats.IndexSize),
			}
			return output.Lines(cmd.OutOrStdout(), lines)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "write structured JSON output")
	return cmd
}

func newUICommand() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Generate a static HTML UI for the current Cosha index",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()
			outPath := out
			if outPath == "" {
				outPath = ui.DefaultReportPath(root)
			}
			if err := ui.WriteReport(store, outPath); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote Cosha UI to %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "path for the generated HTML file")
	return cmd
}

func openStore() (*idx.Store, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return idx.OpenExisting(root)
}
