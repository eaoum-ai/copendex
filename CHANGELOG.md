# Changelog

All notable changes to Copendex will be documented in this file.

This project uses calendar versioning in the `YY.MM.N` format once public versioned releases begin.

Changes stay under `Unreleased` until a release is cut. At release time, entries move into a dated version section such as `## 26.06.0 - 2026-06-12`.

## Unreleased

### Added

- `copendex detect` to report whether the current repository has Java source code or Java project markers.
- Comma-separated kind filters such as `--kind class,interface` for search and symbols commands.
- Makefile install and uninstall targets for putting `copendex` on `PATH`.
- Tree-sitter-backed Java symbol extraction for constructors, nested types, interfaces, enums, enum constants, and overloaded methods.
- SQLite index schema version metadata, exposed in stats JSON.
- Clear errors for missing, stale, or incompatible local indexes.
- `copendex index --rebuild`/`-r` for explicit local index recreation.
- Search filters for kind, language, file path, and package.
- Initial Go CLI foundation for local-first codebase indexing.
- Repository initialization with `.copendex/config.yaml`.
- Java file discovery, lightweight symbol extraction, and SQLite-backed indexing.
- Search, symbols, stats, and static UI report commands.
- JSON output for agent consumption.
- Basic tests, CI, DCO workflow, and contribution docs.
- Trunk-based development and release/version management documentation.
- Compatibility policy for rebuildable local indexes and future MCP server TODO.
