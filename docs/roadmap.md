# Roadmap

This roadmap captures direction, not a fixed delivery promise. Detailed task tracking lives in [planning/backlog.md](planning/backlog.md), and release grouping lives in [planning/milestones.md](planning/milestones.md).

## Current Focus

- Keep the Go CLI foundation small, reliable, and easy to build.
- Improve deterministic repository scanning and indexing.
- Keep JSON output stable enough for coding agents to consume.
- Document compatibility expectations before expanding agent integrations.

## Near Term

- Move generated project indexes and UI output to a shared local Copendex home while keeping per-repository config in `.copendex/config.yaml`.
  - Store generated indexes under `~/.copendex/projects/<project-id>/`.
  - Generate one shared static UI under `~/.copendex/ui/`.
  - Let users choose the indexed project from the shared static UI.
  - Keep the UI static-only; do not introduce a local server for normal UI usage.
  - Preserve a low CPU and memory footprint by generating data ahead of time and avoiding background processes.
- Add index schema version metadata and compatibility checks.
- Improve Java symbol extraction while keeping the extractor replaceable.
- Add richer search filters for kind, language, file path, and package.
- Add an MCP server as a stable agent-facing compatibility layer.
- Add release automation for Linux, macOS, and Windows binaries.
  - Keep local development builds as ignored artifacts such as `./copendex` or `dist/copendex`.
  - Install packaged binaries through system paths managed by the package manager, such as Homebrew or Snap.
  - Keep `~/.copendex/` for runtime data only, not installed binaries.
  - Support future Homebrew, Snap, and manual install flows without changing project config locations.

## Later

- Add references and relationships.
- Add incremental indexing.
- Add more languages.
- Add stronger parser integrations where they justify their complexity.

## Non-Goals For Now

- No SaaS or cloud service.
- No LLM dependency inside Copendex.
- No embeddings in v0.
- No browser application beyond generated static local reports.
- No always-on background daemon or local UI server.
- No IDE plugin until the CLI and agent-facing contracts are stable.
