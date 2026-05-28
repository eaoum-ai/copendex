# Backlog

Backlog items are intentionally lightweight. Move items into [milestones.md](milestones.md) when they become release candidates.

## Next

- [x] Add index schema version metadata to SQLite.
- [x] Add clear errors for missing, stale, or incompatible indexes.
- [x] Add `cosha index --rebuild` as an explicit rebuild command.
- [x] Add richer query filters for symbols and search.
- [x] Improve Java method extraction coverage.
- [x] Add tests for nested classes, annotations, constructors, interfaces, enums, and enum constants.
- [ ] Document JSON output contracts for `stats`, `symbols`, and `search`.
- [ ] Add real Java repository smoke test checklist.
- [ ] Add `make integration` for compiled CLI tests against generated temporary Java repositories.
- [ ] Add optional `make smoke` using a gitignored `COSHA_SMOKE_REPO` path for real Java repository validation.
- [ ] Add troubleshooting for `.gitignore`, config, SQLite, and rebuild behavior.
- [ ] Document search matching and ranking semantics.
- [ ] Add exact-match and result-limit search options.
- [ ] Add `cosha search --expand` to return primary matches for requested filters first, then append related non-matching results when a kind or other filters are provided.

## Parser And Indexing

- [ ] Evaluate parser replacement options for Java.

## Incremental Indexing

- [ ] Add file fingerprint metadata for indexed files; bumps schema version.
- [ ] Add changed-file detection that compares path, size, modtime, and content hash before extraction.
- [ ] Add per-file symbol replacement so indexing one changed file removes and reinserts only that file's symbols.
- [ ] Add deleted-file cleanup so symbols for removed files disappear without a full rebuild.
- [ ] Add `cosha index` incremental behavior for compatible indexes while preserving `cosha index --rebuild` as the explicit full rebuild path.
- [ ] Add JSON stats fields for indexed file counts and last incremental update metadata.
- [ ] Add tests for changed, unchanged, and deleted Java files across repeated index runs.

## Symbol Detail

- [ ] Add symbol signature storage for methods and constructors; bumps schema version.
- [ ] Add Java signature extraction for constructors, methods, overloads, and basic parameter lists.
- [ ] Add leading documentation storage for declarations; bumps schema version if stored outside the existing symbols table.
- [ ] Add Java leading doc extraction for adjacent block and line comments.
- [ ] Include signatures and leading docs in `symbols --json` and `search --json` as optional fields.
- [ ] Add text output formatting that shows signatures without making existing symbol listings noisy.
- [ ] Add fixtures for overloaded methods with distinct signatures and docs.

## References And Relationships

- [ ] Add a `references` table for symbol usages; bumps schema version.
- [ ] Add Java reference extraction for same-file method calls and type mentions behind a new query path.
- [ ] Add a references query command for "who uses this symbol" without changing existing `search` behavior.
- [ ] Add reference result JSON with source span, referenced symbol id when resolved, and unresolved name when not resolved.
- [ ] Add a `relationships` table for syntax-level edges such as `extends` and `implements`; bumps schema version.
- [ ] Add Java hierarchy extraction for class extends, interface extends, and implements clauses.
- [ ] Add hierarchy queries for "who implements this interface" and "what subclasses this type".
- [ ] Add tests for overloaded calls, unresolved references, implements edges, and subclass edges.

## Source Navigation

- [ ] Add `cosha source <symbol-id>` to return the enclosing node's source text and location.
- [ ] Add JSON output for source navigation with file path, line range, and symbol metadata.
- [ ] Add tests for nested types and methods so source extraction returns the intended enclosing node.

## Agent Integrations

- [ ] Add MCP server scaffold after incremental indexing, symbol detail, and references are useful through the CLI.
- [ ] Add read-only MCP tools for search and symbol lookup.
- [ ] Add read-only MCP tools for references, hierarchy, and source navigation after the matching CLI commands exist.
- [ ] Add initial MCP schema compatibility policy.
- [ ] Add MCP re-index guidance for missing, stale, or incompatible local indexes.

## Distribution

- [ ] Add GoReleaser configuration.
- [ ] Publish Linux, macOS, and Windows binaries.
- [ ] Add install instructions for common package managers after binaries exist.

## Documentation

- [ ] Add examples for indexing a real Java repository.
- [ ] Add JSON output contract examples.
- [ ] Add troubleshooting for `.gitignore`, config, and SQLite issues.
