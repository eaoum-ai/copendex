# Backlog

Backlog items are intentionally lightweight. Move items into [milestones.md](milestones.md) when they become release candidates.

## Next

- [x] Add index schema version metadata to SQLite.
- [x] Add clear errors for missing, stale, or incompatible indexes.
- [x] Add `copendex index --rebuild` as an explicit rebuild command.
- [x] Add richer query filters for symbols and search.
- [x] Improve Java method extraction coverage.
- [x] Add tests for nested classes, annotations, constructors, interfaces, enums, and enum constants.
- [ ] Document JSON output contracts for `stats`, `symbols`, and `search`.
- [ ] Add real Java repository smoke test checklist.
- [ ] Add troubleshooting for `.gitignore`, config, SQLite, and rebuild behavior.
- [ ] Document search matching and ranking semantics.
- [ ] Add exact-match and result-limit search options.

## Parser And Indexing

- [ ] Add reference and relationship tables.
- [ ] Add incremental indexing.
- [ ] Evaluate parser replacement options for Java.

## Agent Integrations

- [ ] Add MCP server scaffold after the CLI foundation is functional on real Java repositories.
- [ ] Add initial MCP schema compatibility policy.

## Distribution

- [ ] Add GoReleaser configuration.
- [ ] Publish Linux, macOS, and Windows binaries.
- [ ] Add install instructions for common package managers after binaries exist.

## Documentation

- [ ] Add examples for indexing a real Java repository.
- [ ] Add JSON output contract examples.
- [ ] Add troubleshooting for `.gitignore`, config, and SQLite issues.
