# Agent Guide

Use this file as the first stop when working in this repository.

## Where To Look

- Project overview and CLI usage: `README.md`
- Development workflow, branch names, and PR expectations: `docs/development.md`
- Release process and versioning: `docs/releases.md`
- Roadmap and scope boundaries: `docs/roadmap.md`
- Planning notes and backlog: `docs/planning/README.md`
- Architecture decisions: `docs/decisions/`
- User-visible changes: `CHANGELOG.md`

## Code Map

- CLI wiring: `internal/cli/`
- Repository config handling: `internal/config/`
- File discovery and ignore handling: `internal/files/`
- Index models and SQLite storage: `internal/index/`
- Java extraction: `internal/lang/java/`
- Search behavior: `internal/search/`
- Output formatting: `internal/output/`
- Static report UI: `internal/ui/`
- Project scripts: `scripts/`

## Local Commands

Use the Makefile commands when possible:

```sh
make test
make build
make check
make tidy
make clean
```

`make test`, `make build`, and `make tidy` keep Go caches under `.cache/`, which is ignored. The built CLI is `./cosha`, also ignored.

## Generated Files

Do not commit local Cosha runtime output:

- `.cosha/`
- `.cache/`
- `cosha`

The static UI source lives at `internal/ui/static/index.html`; generated UI snapshots live under `.cosha/ui/`.
