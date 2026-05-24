package java

import (
	"regexp"
	"strings"

	"github.com/eaoum-ai/copendex/internal/index"
)

var (
	packageRE    = regexp.MustCompile(`^\s*package\s+([A-Za-z_][\w.]*)\s*;`)
	importRE     = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([A-Za-z_][\w.*]*)\s*;`)
	typeRE       = regexp.MustCompile(`\b(class|interface|enum)\s+([A-Za-z_]\w*)`)
	methodRE     = regexp.MustCompile(`^\s*(?:public|protected|private|static|final|synchronized|abstract|native|strictfp|\s)+[\w<>\[\], ?]+\s+([A-Za-z_]\w*)\s*\([^;]*\)\s*(?:throws\s+[\w.,\s]+)?\{?`)
	annotationRE = regexp.MustCompile(`^\s*@([A-Za-z_][\w.]*)`)
)

func Extract(path string, content []byte) []index.Symbol {
	lines := strings.Split(string(content), "\n")
	var packageName string
	var pendingAnnotations []string
	var symbols []index.Symbol
	for i, line := range lines {
		lineNo := i + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if match := packageRE.FindStringSubmatch(line); match != nil {
			packageName = match[1]
			symbols = append(symbols, symbol(path, "package", match[1], packageName, lineNo, nil))
			continue
		}
		if match := importRE.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, symbol(path, "import", match[1], packageName, lineNo, nil))
			continue
		}
		if match := annotationRE.FindStringSubmatch(line); match != nil {
			name := shortName(match[1])
			pendingAnnotations = append(pendingAnnotations, name)
			symbols = append(symbols, symbol(path, "annotation", name, packageName, lineNo, nil))
			continue
		}
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if match := typeRE.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, symbol(path, match[1], match[2], packageName, lineNo, pendingAnnotations))
			pendingAnnotations = nil
			continue
		}
		if match := methodRE.FindStringSubmatch(line); match != nil {
			name := match[1]
			if isControlKeyword(name) {
				continue
			}
			symbols = append(symbols, symbol(path, "method", name, packageName, lineNo, pendingAnnotations))
			pendingAnnotations = nil
			continue
		}
		if !strings.HasPrefix(trimmed, "@") {
			pendingAnnotations = nil
		}
	}
	return symbols
}

func symbol(path, kind, name, packageName string, line int, annotations []string) index.Symbol {
	cp := append([]string(nil), annotations...)
	return index.Symbol{
		Name:        name,
		Kind:        kind,
		Language:    "java",
		File:        path,
		PackageName: packageName,
		Line:        line,
		Annotations: cp,
	}
}

func shortName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func isControlKeyword(name string) bool {
	switch name {
	case "if", "for", "while", "switch", "catch":
		return true
	default:
		return false
	}
}
