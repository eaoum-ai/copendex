# 0004 File Fingerprint Metadata For Incremental Indexing

## Status

Proposed

## Context

`docs/roadmap.md` Phase 1 commits Cosha to making the edit-to-query cycle cheap before expanding extraction. The first backlog item under "Incremental Indexing" in `docs/planning/backlog.md` calls for "file fingerprint metadata for indexed files; bumps schema version," ahead of a separate slice for changed-file detection.

The `files` table at `internal/index/store.go` already stores `path`, `language`, `size_bytes`, `modified_at`, and `hash` (sha256 of file contents, populated in `internal/files/scanner.go`). These columns capture *the file's* identity: when the file was modified and what its content was the last time we looked. They do not capture *the indexer's* state: when Cosha last wrote a row for that path.

Changed-file detection needs both. The file's mtime says when the file was touched on disk. The indexer's timestamp says when Cosha last processed it. Comparing them is what tells us whether a file is stale relative to the index. Surfacing the indexer timestamp also unblocks the later backlog item to report "last incremental update metadata" in `cosha stats --json`.

Today both timestamps would collapse to the same value because `internal/index/store.go::Rebuild` wipes and reinserts every file every run. That changes the moment changed-file detection lands; this slice prepares the schema so the next slice has a comparison anchor.

## Decision

Add an `indexed_at` column to the `files` table that records, in UTC RFC3339Nano, the moment Cosha wrote each row.

Specifically:

- Add `indexed_at TEXT NOT NULL` to the `files` table CREATE in `internal/index/store.go`.
- Populate it from `time.Now().UTC().Format(time.RFC3339Nano)` inside `Rebuild`, for every inserted row.
- Surface it as `IndexedAt time.Time` on `index.File`, alongside the existing `LastModified time.Time` (file mtime).
- Read it back in `scanFile` so any query path that returns `File` records carries the timestamp.
- Bump `CurrentSchemaVersion` from `1` to `2`. Existing local indexes will surface `IndexError{Kind: IncompatibleIndex}` and require `cosha index --rebuild`, per the existing contract documented in `CLAUDE.md` and ADR-0001's consequences.

This slice does **not** change CLI behavior, JSON output for `stats`, `symbols`, or `search`, or the scanner. The new field is structurally present and populated, ready to be read by the next slice. Stats reporting of last-touched timestamps and the changed-file detection logic itself are deliberately out of scope to keep this PR small and revertable.

## Consequences

- Existing local Cosha indexes become incompatible and report a clear "run `cosha index --rebuild`" error on the next query. This is the contract the project already documents for any schema change.
- The next slice (changed-file detection) has a stable anchor: for each file row, it can compare on-disk size/mtime/hash against the stored fingerprint and skip extraction when nothing has changed.
- A later slice can surface "last incremental update" in `cosha stats --json` by reading `MAX(indexed_at)` without any further schema work.
- One additional column write per file per indexing run. Negligible compared to file I/O and Tree-sitter parsing.
- `File` JSON output gains an `indexedAt` field. This is additive — agents already consuming `lastModified` are unaffected.
- Distinguishing "file modified" from "indexer touched" prevents future confusion when stats begin reporting freshness; today both happen to be the same moment, so picking the naming now avoids retrofitting later.

## Alternatives Considered

- **Use `modified_at` (file mtime) as the freshness signal.** Conflates the file system's record of the file with Cosha's record of when it was indexed. The two diverge the moment incremental indexing lands. Picking now is cheaper than renaming later.
- **Add a separate `index_runs` table and key files into it by run id.** Useful for grouping files by indexing session, but premature for v0. The roadmap does not call for per-run reporting yet, and the next two slices (changed-file detection, deleted-file cleanup) don't require it. Can be layered on later without altering the `indexed_at` semantics.
- **Skip the schema change and rely entirely on content-hash comparison during change detection.** Correct, but pays the full file-read + sha256 cost on every file on every run. Storing `indexed_at` is the minimum addition that lets the next slice short-circuit on the size/mtime fast path before falling back to hash.
- **Defer the schema bump until changed-file detection lands and combine the two slices.** The maintainer's backlog explicitly separates them, and a smaller PR is easier to review and roll back. Keeping the slices separate also lets the schema change ship even if the detection algorithm needs more iteration.
