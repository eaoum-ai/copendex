# Roadmap

This roadmap captures direction, not a fixed delivery promise. Detailed task tracking lives in [planning/backlog.md](planning/backlog.md), and release grouping lives in [planning/milestones.md](planning/milestones.md).

## Current Focus

- Keep the Go CLI foundation small, reliable, and easy to build.
- Improve deterministic repository scanning and indexing.
- Make Java indexing useful on real repositories before adding an MCP layer.
- Keep JSON output stable enough for coding agents to consume through the CLI.
- Document compatibility expectations before expanding agent integrations.

## Near Term

- Finish index lifecycle behavior.
  - Add index schema version metadata and compatibility checks.
  - Add explicit rebuild guidance for missing, stale, or incompatible indexes.
  - Require `cosha index --rebuild` or `-r` before overwriting an existing compatible index.
- Improve Java symbol extraction while keeping the extractor replaceable.
  - Cover constructors, nested classes, annotations, interfaces, enums, enum constants, and overloaded methods.
  - Add realistic Java fixtures for common application code shapes.
- Improve CLI search usefulness.
  - Keep filters for kind, language, file path, and package.
  - Add stable ranking behavior and document matching semantics.
  - Consider exact-match and result-limit flags once ranking is settled.
  - Consider an `--expand` flag that returns filtered matches first, followed by related results outside the requested kind or filters.
- Document the CLI JSON contract for `stats`, `symbols`, and `search`.
- Add a real-repository smoke test checklist for init, index, stats, search, and rebuild.
- Add troubleshooting for `.gitignore`, include/exclude config, SQLite, and rebuild behavior.

## After MVP

- Move generated project indexes and UI output to a shared local Cosha home while keeping per-repository config in `.cosha/config.yaml`.
  - Store generated indexes under `~/.cosha/projects/<project-id>/`.
  - Generate one shared static UI under `~/.cosha/ui/`.
  - Let users choose the indexed project from the shared static UI.
  - Keep the UI static-only; do not introduce a local server for normal UI usage.
  - Preserve a low CPU and memory footprint by generating data ahead of time and avoiding background processes.
- Add an MCP server as a stable agent-facing compatibility layer after the CLI, index, search, and JSON contracts are useful on real Java repositories.
- Add release automation for Linux, macOS, and Windows binaries.
  - Keep local development builds as ignored artifacts such as `./cosha` or `dist/cosha`.
  - Install packaged binaries through system paths managed by the package manager, such as Homebrew or Snap.
  - Keep `~/.cosha/` for runtime data only, not installed binaries.
  - Support future Homebrew, Snap, and manual install flows without changing project config locations.

## Later

- Add references and relationships.
- Add incremental indexing.
- Add more languages.
- Add stronger parser integrations where they justify their complexity.

## Non-Goals For Now

- No SaaS or cloud service.
- No LLM dependency inside Cosha.
- No embeddings in v0.
- No browser application beyond generated static local reports.
- No always-on background daemon or local UI server.
- No IDE plugin until the CLI and agent-facing contracts are stable.
