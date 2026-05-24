# 0001 Use Go, Cobra, and SQLite for v0

## Status

Accepted

## Context

Copendex should be local-first, deterministic-first, cross-platform, and easy for coding agents to call. The v0 foundation should avoid cloud services, LLM dependencies, embeddings, and complex runtime requirements.

## Decision

Use Go for the v0 implementation, Cobra for the CLI, and SQLite for local index storage.

Prefer `modernc.org/sqlite` over CGO-based SQLite to keep cross-platform builds simpler.

## Consequences

- Copendex can be distributed as a single binary.
- The CLI can expose text and JSON output without requiring a service process.
- SQLite gives the index a durable local storage layer.
- Parser integrations should be chosen carefully so they do not undermine simple cross-platform builds.
