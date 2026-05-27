# 0003 Use Tree-sitter for MVP Parsing

## Status

Accepted

## Context

Copendex needs Java symbol extraction that is useful on real repositories before adding MCP or broader agent integrations. The initial lightweight extractor keeps the v0 implementation simple, but regex-based parsing is brittle for common Java code shapes such as constructors, nested types, interfaces, enums, annotations, overloaded methods, and varied formatting.

The MVP should prove that Copendex can build reliable syntax-aware local indexes while keeping the core index, CLI, and JSON contracts independent from any one parser implementation.

## Decision

Use Tree-sitter as the parser foundation for MVP Java symbol extraction.

Start with Java by using the official Tree-sitter Go bindings and the Tree-sitter Java grammar. Keep Tree-sitter usage contained under language-specific extraction packages, and expose extracted symbols through Copendex's existing index models.

The MVP parser scope is syntax-level declarations:

- packages
- imports
- classes
- interfaces
- enums
- enum constants
- constructors
- methods
- annotations attached to declarations
- nested types

Do not add semantic resolution in this decision. References, relationships, inheritance, call graphs, and type resolution require separate design work.

## Consequences

- Java extraction should become more accurate and credible for real repositories.
- Tree-sitter creates a path for adding other languages through grammar-specific extractors.
- Copendex now requires Go 1.26 or newer for development and CI.
- Parser dependencies and cross-platform release behavior need explicit validation.
- The selected Go bindings are CGO-backed; `CGO_ENABLED=0` builds are not expected to work while this parser path is active.
- The CLI and index layers should remain parser-agnostic so Tree-sitter can be upgraded, replaced, or supplemented without changing user-facing contracts.
- The MCP layer remains deferred until CLI indexing, search, and JSON output are useful with parser-backed symbols.
