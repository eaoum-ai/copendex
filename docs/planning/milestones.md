# Milestones

Cosha uses `YY.MM.N` release versions. Milestones are planning targets and may change as scope becomes clearer.

## 26.05.0

Initial OSS foundation.

- CLI scaffold.
- Repository initialization.
- Java file discovery.
- Lightweight Java symbol extraction.
- SQLite-backed local index.
- Search, symbols, stats, and static UI report commands.
- JSON output for agents.
- Basic tests and CI.
- Apache 2.0 license and DCO workflow.
- Trunk-based development and release documentation.

## 26.06.0

MVP CLI, index, and search functionality.

- Index schema version metadata.
- Explicit rebuild guidance for incompatible local indexes.
- Better structured errors for missing or stale indexes.
- `cosha index --rebuild` and `-r` for explicit local index recreation.
- Better Java extraction coverage for constructors, nested classes, annotations, interfaces, enums, enum constants, and overloaded methods.
- Search filters and documented matching/ranking behavior.
- JSON output contract examples for `stats`, `symbols`, and `search`.
- Real Java repository smoke test checklist.
- Compiled CLI integration test command for generated temporary Java repositories.
- Troubleshooting documentation for `.gitignore`, config, SQLite, and rebuild behavior.

## 26.07.0

Incremental indexing foundation for cheap edit-to-query cycles.

- File fingerprint metadata for indexed files.
- Changed-file detection across repeated index runs.
- Per-file symbol replacement for changed files.
- Deleted-file cleanup without a full rebuild.
- Incremental `cosha index` behavior for compatible indexes.
- `cosha index --rebuild` preserved as the explicit full rebuild path.
- Stats metadata for indexed files and recent incremental updates.
- Tests for changed, unchanged, and deleted Java files.

## 26.08.0

Symbol detail that lets agents disambiguate results before opening source files.

- Symbol signature storage for methods and constructors.
- Java signature extraction for constructors, methods, overloads, and basic parameter lists.
- Leading documentation storage for declarations.
- Java leading doc extraction for adjacent block and line comments.
- Optional signature and documentation fields in JSON output.
- Text output that shows signatures without overwhelming existing listings.
- Fixtures for overloaded methods with distinct signatures and docs.

## 26.09.0

Reference queries that make Cosha useful before rename and signature-change work.

- References table for symbol usages.
- Java reference extraction for same-file method calls and type mentions.
- CLI query path for "who uses this symbol".
- Reference JSON with source spans, resolved symbol ids when available, and unresolved names when not.
- Tests for overloaded calls and unresolved references.

## 26.10.0

Hierarchy and source navigation for common Java change planning questions.

- Relationship table for syntax-level `extends` and `implements` edges.
- Java hierarchy extraction for class extends, interface extends, and implements clauses.
- CLI query paths for implementors and subclasses.
- `cosha source <symbol-id>` for enclosing node source text and location.
- Tests for implements edges, subclass edges, nested types, and method source extraction.

## 26.11.0

MCP scaffold over an index that is useful enough for agent editing loops.

- Read-only MCP server scaffold.
- MCP tools for search and symbol lookup.
- MCP tools for references, hierarchy, and source navigation.
- Initial MCP schema compatibility policy.
- Re-index guidance for missing, stale, or incompatible local indexes.

## Future

- Additional languages.
- Alternative parser integrations where they justify their complexity.
- Deeper semantic resolution beyond syntax-level references and relationships.
- Cross-platform release automation.
