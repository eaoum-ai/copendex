# Changelog

All notable changes to Cosha will be documented in this file.

This project uses calendar versioning in the `YY.MM.N` format once public versioned releases begin.

Changes stay under `Unreleased` until a release is cut. At release time, entries move into a dated version section such as `## 26.06.0 - 2026-06-12`.

## Unreleased

### Changed

- Restructure roadmap and planning docs around the agent-useful index sequence: incremental indexing, symbol detail, references and hierarchy, then MCP.
- Rewrite symbol and file search to use SQLite FTS5 virtual tables backed by triggers on the canonical tables. Token and prefix queries on large Java repositories drop from full-table-scan latency to single-digit milliseconds. Bumps SQLite index schema to version 2; existing local indexes surface `IncompatibleIndex` and must be rebuilt with `cosha index --rebuild`. See `docs/decisions/0005-use-sqlite-fts5-for-search.md`.
- Search semantics shift from substring-anywhere (the old `LIKE '%q%'` behavior) to token-prefix matching. Queries against the start of an identifier work natively (`cosha search Owner` finds `OwnerController` and `OwnerRepository`); queries that target the middle of an identifier (`cosha search Service` against `AuthorizationService`) no longer match. CamelCase-fuzzy matching is deferred to a later slice.

### Added
- Rebrand the project and command to Cosha, short for Code Shodha.
- `cosha detect` to report whether the current repository has Java source code or Java project markers.
- Comma-separated kind filters such as `--kind class,interface` for search and symbols commands.
- Makefile install and uninstall targets for putting `cosha` on `PATH`.
- Tree-sitter-backed Java symbol extraction for constructors, nested types, interfaces, enums, enum constants, and overloaded methods.
- SQLite index schema version metadata, exposed in stats JSON.
- Clear errors for missing, stale, or incompatible local indexes.
- `cosha index --rebuild`/`-r` for explicit local index recreation.
- Search filters for kind, language, file path, and package.
- Initial Go CLI foundation for local-first codebase indexing.
- Repository initialization with `.cosha/config.yaml`.
- Java file discovery, lightweight symbol extraction, and SQLite-backed indexing.
- Search, symbols, stats, and static UI report commands.
- JSON output for agent consumption.
- Basic tests, CI, DCO workflow, and contribution docs.
- Trunk-based development and release/version management documentation.
- Compatibility policy for rebuildable local indexes and future MCP server TODO.
