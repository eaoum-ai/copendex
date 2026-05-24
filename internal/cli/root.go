package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eaoum-ai/copendex/internal/config"
	"github.com/eaoum-ai/copendex/internal/files"
	idx "github.com/eaoum-ai/copendex/internal/index"
	"github.com/eaoum-ai/copendex/internal/lang/java"
	"github.com/eaoum-ai/copendex/internal/output"
	"github.com/eaoum-ai/copendex/internal/search"
	"github.com/eaoum-ai/copendex/internal/ui"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copendex",
		Short: "Local-first codebase intelligence for coding agents",
	}
	cmd.AddCommand(newInitCommand(), newIndexCommand(), newSearchCommand(), newSymbolsCommand(), newStatsCommand(), newUICommand())
	return cmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Copendex in the current repository",
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
	return &cobra.Command{
		Use:   "index",
		Short: "Index the current repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
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
}

func newSearchCommand() *cobra.Command {
	var jsonOut bool
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
			results, err := search.New(store).All(args[0])
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
	return cmd
}

func newSymbolsCommand() *cobra.Command {
	var jsonOut bool
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
			symbols, err := search.New(store).Symbols(args[0])
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
	return cmd
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
		Short: "Generate a static HTML UI for the current Copendex index",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote Copendex UI to %s\n", outPath)
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
	return idx.Open(root)
}
