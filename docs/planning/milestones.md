# Milestones

Cosha uses `YY.MM.N` release versions. Milestones are planning targets and may change as scope becomes clearer.

## 26.05.0

Initial OSS foundation.

- CLI scaffold.
- Repository initialization.
- Java file discovery.
- Lightweight Java symbol extraction.
- SQLite-backed local index.
- Search, symbols, stats, and static UI report commands.
- JSON output for agents.
- Basic tests and CI.
- Apache 2.0 license and DCO workflow.
- Trunk-based development and release documentation.

## 26.06.0

MVP CLI, index, and search functionality.

- Index schema version metadata.
- Explicit rebuild guidance for incompatible local indexes.
- Better structured errors for missing or stale indexes.
- `cosha index --rebuild` and `-r` for explicit local index recreation.
- Better Java extraction coverage for constructors, nested classes, annotations, interfaces, enums, enum constants, and overloaded methods.
- Search filters and documented matching/ranking behavior.
- JSON output contract examples for `stats`, `symbols`, and `search`.
- Real Java repository smoke test checklist.
- Troubleshooting documentation for `.gitignore`, config, SQLite, and rebuild behavior.

## Future

- MCP server scaffold for local agent integrations.
- Initial MCP schema compatibility policy.
- More complete Java extraction.
- Additional languages.
- Incremental indexing.
- References and relationships.
- Cross-platform release automation.
