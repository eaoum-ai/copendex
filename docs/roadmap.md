# Roadmap

This roadmap captures direction, not a fixed delivery promise. Detailed task tracking lives in [planning/backlog.md](planning/backlog.md), and release grouping lives in [planning/milestones.md](planning/milestones.md).

## Current Focus

- Keep the Go CLI foundation small, reliable, and easy to build.
- Move the index from structured grep for Java declarations toward data an agent can query while editing code.
- Make edit-to-query cycles cheap before expanding the amount of indexed data.
- Add enough symbol detail to let agents disambiguate results without immediately opening every file.
- Add references and relationships before exposing a broader agent-facing server surface.
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
- Add release automation for Linux, macOS, and Windows binaries.
  - Keep local development builds as ignored artifacts such as `./cosha` or `dist/cosha`.
  - Install packaged binaries through system paths managed by the package manager, such as Homebrew or Snap.
  - Keep `~/.cosha/` for runtime data only, not installed binaries.
  - Support future Homebrew, Snap, and manual install flows without changing project config locations.

## Phase 1: Cheap Edit-To-Query Loop

- Add incremental indexing so agents can re-index changed files after edits without paying a full repository rebuild.
- Track file fingerprints and remove stale symbols when files are changed or deleted.
- Keep full rebuild available as the compatibility fallback for missing, stale, or incompatible indexes.

Why this before richer extraction: every later feature is more useful when agents can update the local index cheaply during an edit loop.

## Phase 2: Disambiguating Symbols

- Add symbol signatures for methods, constructors, and other declarations where the parser can extract them reliably.
- Add leading documentation snippets when they are adjacent to indexed declarations.
- Keep the base symbol identity stable while adding optional fields to JSON output.

Why this before references: signatures and docs are the cheapest way to distinguish overloads and similarly named declarations without forcing agents back to file reads.

## Phase 3: References And Relationships

- Add references/usages for Java symbols.
- Add relationship edges for implements, extends, and subclass queries where syntax-level extraction can support them.
- Add query paths for "who calls X", "who implements Y", and "what subclasses Z" before broadening language coverage.
- Consider `cosha source <symbol-id>` for returning the enclosing syntax node text once symbol identity is stable enough.

Why this before MCP: references and hierarchy are the first capabilities that clearly move Cosha beyond `rg` plus file reads for rename and signature-change workflows.

## Phase 4: Agent-Facing Server

- Add an MCP server as a stable agent-facing compatibility layer after the CLI, index, search, and JSON contracts are useful on real Java repositories.
- Expose read-only tools for search, symbols, references, and source navigation before adding write-oriented behavior.
- Keep schema compatibility and re-index guidance explicit for agent clients.

Why this last in the sequence: an MCP layer should surface a useful local index, not make today's declaration-only structured grep easier to call.

## Later

- Add more languages under `internal/lang/<lang>/` after Java proves the query model.
- Add stronger parser integrations where they justify their complexity.
- Add deeper semantic resolution after syntax-level references and relationships are useful.

## Non-Goals For Now

- No SaaS or cloud service.
- No LLM dependency inside Cosha.
- No embeddings in v0.
- No browser application beyond generated static local reports.
- No always-on background daemon or local UI server.
- No IDE plugin until the CLI and agent-facing contracts are stable.
- No cloud-hosted index, shared remote workspace, or background sync service.
- No full semantic type solver or call graph as part of the first references pass.
