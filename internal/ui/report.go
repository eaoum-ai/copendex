package ui

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/eaoum-ai/copendex/internal/index"
)

//go:embed static/index.html
var templateFS embed.FS

type ReportData struct {
	Stats   index.Stats          `json:"stats"`
	Results []index.SearchResult `json:"results"`
}

func DefaultReportPath(root string) string {
	return filepath.Join(root, ".copendex", "ui", "index.html")
}

func WriteReport(store *index.Store, outPath string) error {
	stats, err := store.Stats()
	if err != nil {
		return err
	}
	results, err := store.SearchAll("")
	if err != nil {
		return err
	}
	data, err := json.Marshal(ReportData{Stats: stats, Results: results})
	if err != nil {
		return err
	}
	template, err := templateFS.ReadFile("static/index.html")
	if err != nil {
		return err
	}
	html := strings.Replace(string(template), "__COPENDEX_DATA__", string(data), 1)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(html), 0o644)
}
