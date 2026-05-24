# Changelog

All notable changes to Copendex will be documented in this file.

This project uses calendar versioning in the `YY.MM.N` format once public versioned releases begin.

## Unreleased

### Added

- SQLite index schema version metadata, exposed in stats JSON.
- Clear errors for missing, stale, or incompatible local indexes.
- `copendex index --rebuild` for explicit local index recreation.
- MCP server scaffold with stdio JSON-RPC initialization.
- Search filters for kind, language, file path, and package.
- Initial Go CLI foundation for local-first codebase indexing.
- Repository initialization with `.copendex/config.yaml`.
- Java file discovery, lightweight symbol extraction, and SQLite-backed indexing.
- Search, symbols, stats, and static UI report commands.
- JSON output for agent consumption.
- Basic tests, CI, DCO workflow, and contribution docs.
- Trunk-based development and release/version management documentation.
- Compatibility policy for rebuildable local indexes and future MCP server TODO.
