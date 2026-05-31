# 0005 Use SQLite FTS5 For Symbol And File Search

## Status

Proposed

## Context

`docs/roadmap.md` Phase 1 commits Cosha to making the edit-to-query cycle cheap. Phase 3 (references and hierarchy) and Phase 4 (MCP server) will multiply the number of queries each agent action issues. Search latency is on the critical path for all of this.

The current implementation in `internal/index/store.go::querySymbols` and `queryFiles` issues SQL of the form:

```sql
WHERE lower(symbols.name) LIKE '%' || lower(?) || '%'
```

Three independent properties of this query defeat index use:

- `lower(column)` wraps the indexed column, so the existing `idx_symbols_name` index cannot be used to satisfy the predicate.
- The leading `%` in the LIKE pattern makes any B-tree index unhelpful even if it could be consulted, because B-tree lookups require a fixed prefix.
- Each match joins back to `files` for projection.

The parallel `queryFiles` does the same against `files.path`. On a repository the size of Kafka or Spring (roughly 100k+ symbols once indexed), each query measures at 100–200ms — two full table scans plus a join. That is acceptable for an exploratory CLI but too slow to sit inside an agent's edit-to-query loop, and it will become the dominant cost once references and MCP land.

Cosha needs a search backend that supports token and prefix queries at single-digit-millisecond latency on at least 200k-row symbol indexes, without expanding the cross-platform dependency footprint that ADR-0001 set out to keep small.

## Decision

Adopt SQLite FTS5 as the search backend for symbol-name and file-path queries.

Specifically:

- Add a `symbols_fts` FTS5 virtual table that mirrors the searchable columns of `symbols` (`name`, `kind`, `package_name`), with the underlying `symbols.id` as the FTS rowid so projection joins remain cheap.
- Add a `files_fts` FTS5 virtual table that mirrors `files.path`, keyed to `files.id`.
- Maintain the FTS tables via SQLite triggers on INSERT, UPDATE, and DELETE of the canonical `symbols` and `files` tables. This keeps `Rebuild` and future incremental indexing logic agnostic to the FTS layer.
- Rewrite `querySymbols` and `queryFiles` to issue `MATCH` against the FTS tables and join back to the canonical tables only for the columns needed by `index.Symbol` and `index.File`.
- Use the default `unicode61` tokenizer with `tokenchars` configured to keep `$` and `_` as word characters, so common Java identifier shapes (`Nested$Inner`, `with_underscores`) tokenize as expected.
- Order results by FTS5's built-in BM25 ranking with the existing `rank()` heuristic from `internal/index/search_result.go` retained as a secondary tiebreaker. The current ordering contract (exact match → prefix → substring → other) stays observable to callers; FTS5 is the engine, not the contract.
- Bump `CurrentSchemaVersion`. Existing local indexes surface `IndexError{Kind: IncompatibleIndex}` and require `cosha index --rebuild`, per the existing contract.

This ADR does **not** decide:

- Camel-case fuzzy matching (`AuthSvc` → `AuthorizationService`). FTS5's default tokenizer does not handle this; a custom tokenizer or query-side expansion would be a separate slice.
- A new public CLI surface. The change is internal to the query path; `cosha search` and `cosha symbols` behavior remains the same modulo ranking improvements.
- True substring matching across token boundaries. FTS5 supports `token*` prefix matching natively; a trigram tokenizer can be layered in later if real substring patterns dominate.

## Consequences

- Substring and token queries against 100k+ symbols drop from ~100–200ms full-scan to single-digit milliseconds in typical FTS5 deployments. Phase 3 reference queries inherit the speedup.
- Existing local indexes become incompatible with the new schema. Users see the documented "run `cosha index --rebuild`" error.
- No new dependencies: FTS5 ships with `modernc.org/sqlite` v1.23.1 already in `go.mod`. ADR-0001's posture of minimal cross-platform dependencies is preserved.
- Index storage grows by roughly 20–30% from FTS5's inverted-index payload. Acceptable given Cosha's existing on-disk footprint.
- Triggers add a small write-time cost during `Rebuild`; negligible relative to file I/O and Tree-sitter parsing.
- BM25 ranking changes the observable order of results in edge cases (many fuzzy matches with similar names). The CLI's tie-breaking via the existing `rank()` keeps the deterministic-output guarantee from the project's hard constraints.
- The unicode61 tokenizer segments `AuthorizationService` into `["Authorization", "Service"]`. Queries like `Auth` match by prefix; queries like `AuthSvc` do not match without query-side expansion. This is a known limitation; see "Alternatives Considered" for why deferring camelCase-fuzzy is the right call now.
- Direct arbitrary-infix matching (`%Service%` against `MyServiceImpl`) becomes harder. FTS5 prefix matching plus tokenization covers the dominant agent use cases (find by token, find by prefix). If real-world usage shows infix patterns dominate, a trigram FTS5 tokenizer can be added later without re-architecting the query path.

## Alternatives Considered

- **bleve (pure-Go full-text engine).** Industrial-strength and CGO-free, but adds a third storage surface alongside SQLite (canonical state) and Tree-sitter (parser state). Three index lifecycles to keep in sync during incremental updates is real complexity; the project's stated preference is minimal cross-platform dependency surface. FTS5 lives inside SQLite and inherits its transactional guarantees.
- **Tantivy (Rust via CGO).** Performance leader, but CGO surface fights cross-platform builds. ADR-0001 explicitly chose `modernc.org/sqlite` over CGO SQLite to keep distribution simple. Tree-sitter (ADR-0003) already reintroduced CGO; a second CGO dependency compounds the build-portability cost without proportional benefit for the corpus sizes Cosha targets.
- **Trigram virtual table in SQLite.** Native to SQLite and gives true substring fuzzy matching. Less ergonomic than FTS5's `MATCH` syntax for token-style queries (the dominant agent pattern), and the trigram tokenizer is not as well-tuned for code identifiers. Can be layered in later as a complement if real-world queries show infix patterns dominating.
- **In-memory inverted index rebuilt at CLI startup.** Faster than FTS5 in absolute terms but pays a rebuild cost on every CLI invocation and introduces a "is the in-memory view fresh?" synchronization story that the upcoming incremental indexer will have to solve. FTS5 is persistent and durable through `Rebuild` and incremental updates.
- **Drop `lower()`, use `COLLATE NOCASE` on the existing column.** A tiny, free change that makes prefix queries fast on the existing B-tree index. Does nothing for substring queries (the more common shape). Worth doing alongside FTS5 for the prefix subset, but not a substitute.
- **Defer search-engine work until incremental indexing lands.** Incremental indexing reduces indexing cost; queries stay slow against the indexed corpus regardless. Decoupling the two lets them ship independently. The dependency order is "fingerprint metadata → incremental indexing" and "FTS5 search backend" as parallel tracks; both bump schema and both are blocked only by reaching agreement on the schema version sequence.
- **Custom tokenizer for camelCase decomposition.** Solves the `AuthSvc → AuthorizationService` case at index time. Higher complexity, harder to revert, and the use case is not yet measured. Adopting the unicode61 default first, then iterating once real query traces exist, is the lower-risk path.
