// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// Package java extracts Java symbols for the local index using Tree-sitter.
package java

import (
	"strings"

	"github.com/eaoum-ai/cosha/internal/index"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

func Extract(path string, content []byte) []index.Symbol {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_java.Language())); err != nil {
		return nil
	}
	tree := parser.Parse(content, nil)
	if tree == nil {
		return nil
	}
	defer tree.Close()

	extractor := treeExtractor{
		path:    path,
		content: content,
	}
	extractor.walk(tree.RootNode())
	return extractor.symbols
}

type treeExtractor struct {
	path        string
	content     []byte
	packageName string
	symbols     []index.Symbol
}

func (e *treeExtractor) walk(node *tree_sitter.Node) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "package_declaration":
		e.extractPackage(node)
	case "import_declaration":
		e.extractImport(node)
	case "class_declaration":
		e.extractNamedDeclaration(node, "class")
	case "interface_declaration":
		e.extractNamedDeclaration(node, "interface")
	case "enum_declaration":
		e.extractNamedDeclaration(node, "enum")
	case "constructor_declaration":
		e.extractNamedDeclaration(node, "constructor")
	case "method_declaration":
		e.extractNamedDeclaration(node, "method")
	case "enum_constant":
		e.extractNamedDeclaration(node, "enumConstant")
	case "annotation", "marker_annotation":
		e.extractAnnotation(node)
	}
	cursor := node.Walk()
	defer cursor.Close()
	for _, child := range node.NamedChildren(cursor) {
		child := child
		e.walk(&child)
	}
}

func (e *treeExtractor) extractPackage(node *tree_sitter.Node) {
	name := firstNamedText(e.content, node, "scoped_identifier", "identifier")
	if name == "" {
		return
	}
	e.packageName = name
	e.symbols = append(e.symbols, symbol(e.path, "package", name, e.packageName, line(node), nil))
}

func (e *treeExtractor) extractImport(node *tree_sitter.Node) {
	text := strings.TrimSpace(node.Utf8Text(e.content))
	text = strings.TrimPrefix(text, "import")
	text = strings.TrimSpace(strings.TrimSuffix(text, ";"))
	text = strings.TrimPrefix(text, "static")
	name := strings.TrimSpace(text)
	if name == "" {
		return
	}
	e.symbols = append(e.symbols, symbol(e.path, "import", name, e.packageName, line(node), nil))
}

func (e *treeExtractor) extractNamedDeclaration(node *tree_sitter.Node, kind string) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}
	name := nameNode.Utf8Text(e.content)
	if name == "" {
		return
	}
	e.symbols = append(e.symbols, symbol(e.path, kind, name, e.packageName, line(nameNode), annotations(e.content, node)))
}

func (e *treeExtractor) extractAnnotation(node *tree_sitter.Node) {
	name := annotationName(e.content, node)
	if name == "" {
		return
	}
	e.symbols = append(e.symbols, symbol(e.path, "annotation", name, e.packageName, line(node), nil))
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

func annotations(content []byte, node *tree_sitter.Node) []string {
	var out []string
	cursor := node.Walk()
	defer cursor.Close()
	for _, child := range node.NamedChildren(cursor) {
		child := child
		if child.Kind() != "modifiers" {
			continue
		}
		modCursor := child.Walk()
		for _, modifier := range child.NamedChildren(modCursor) {
			modifier := modifier
			switch modifier.Kind() {
			case "annotation", "marker_annotation":
				if name := annotationName(content, &modifier); name != "" {
					out = append(out, name)
				}
			}
		}
		modCursor.Close()
	}
	return out
}

func annotationName(content []byte, node *tree_sitter.Node) string {
	text := strings.TrimSpace(node.Utf8Text(content))
	text = strings.TrimPrefix(text, "@")
	if idx := strings.IndexAny(text, "(\r\n\t "); idx >= 0 {
		text = text[:idx]
	}
	return shortName(text)
}

func firstNamedText(content []byte, node *tree_sitter.Node, kinds ...string) string {
	cursor := node.Walk()
	defer cursor.Close()
	for _, child := range node.NamedChildren(cursor) {
		child := child
		for _, kind := range kinds {
			if child.Kind() == kind {
				return strings.TrimSpace(child.Utf8Text(content))
			}
		}
		if text := firstNamedText(content, &child, kinds...); text != "" {
			return text
		}
	}
	return ""
}

func line(node *tree_sitter.Node) int {
	return int(node.StartPosition().Row) + 1
}

func shortName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}
