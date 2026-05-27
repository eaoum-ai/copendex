# Copendex

Copendex is a local-first codebase intelligence tool for coding agents.

It runs on a developer's machine and exposes fast, structured codebase search and navigation data without using an LLM internally. The v0 foundation is a single-binary Go CLI backed by a local SQLite index.

## Why It Exists

Coding agents need to find the right files, symbols, usages, tests, configs, entrypoints, and related code quickly while they work. Copendex focuses on deterministic codebase intelligence first: file discovery, ignore handling, language-aware indexing, symbol extraction, and JSON output that agents can consume directly.

Copendex does not send source code to a cloud service. It does not use embeddings or an LLM internally.

## Current Scope

- CLI-first Go application
- Repository initialization
- File discovery with default excludes and practical `.gitignore` support
- Java indexing
- Lightweight Java symbol extraction for packages, imports, classes, interfaces, enums, methods, and annotations
- Local SQLite storage under `.copendex/index/copendex.db`
- Deterministic file and symbol search
- Text and JSON output
- Basic tests and CI

## Development Setup

Requirements:

- Go 1.26 or newer

Install dependencies and run tests:

```sh
go mod tidy
go test ./...
```

Or use the project test scaffold, which keeps Go caches inside the repository:

```sh
make test
```

Build the CLI:

```sh
go build -o copendex ./cmd/copendex
```

Or:

```sh
make build
```

Install the CLI so `copendex` is available from any directory:

```sh
make install
```

By default this installs to `$(go env GOPATH)/bin/copendex`. Ensure that directory is on your `PATH`, or choose a different install directory:

```sh
make install BINDIR=/usr/local/bin
```

Remove an installed local binary:

```sh
make uninstall
```

## CLI Usage

Initialize Copendex in a repository:

```sh
copendex init
```

This creates:

```text
.copendex/config.yaml
```

Index the repository:

```sh
copendex index
copendex index --rebuild
copendex index -r
```

Show index stats:

```sh
copendex stats
copendex stats --json
```

Search symbols:

```sh
copendex symbols Service
copendex symbols Service --json
copendex symbols Service --kind class --package com.example
```

Search files and symbols:

```sh
copendex search AuthorizationService
copendex search AuthorizationService --json
copendex search AuthorizationService --language java --path src/main
```

Generate the local HTML UI:

```sh
copendex ui
```

Then open the generated file:

```text
.copendex/ui/index.html
```

The UI is a static HTML snapshot with embedded index data from `copendex index`. It does not start a server.

## Default Config

```yaml
version: 1
include:
  - "src/**/*.java"
  - "**/*.java"
exclude:
  - "build/**"
  - "target/**"
  - ".git/**"
  - ".copendex/**"
  - "node_modules/**"
index:
  languages:
    - java
output:
  defaultFormat: text
```

## Roadmap

- More robust Java parser integration
- References and relationships tables
- Incremental indexing
- Additional languages
- Richer query filters
- GoReleaser builds for Linux, macOS, and Windows
- MCP server after the CLI foundation is stable, with versioned agent-facing schemas and clear re-index guidance for incompatible local indexes

## Project Process

- Development strategy: [docs/development.md](docs/development.md)
- Release and version management: [docs/releases.md](docs/releases.md)
- Roadmap: [docs/roadmap.md](docs/roadmap.md)
- Text-based planning: [docs/planning/README.md](docs/planning/README.md)
- Architecture decisions: [docs/decisions](docs/decisions)
- Changelog: [CHANGELOG.md](CHANGELOG.md)

## License

Copendex is licensed under the [Apache License 2.0](LICENSE).

You may use, modify, distribute, and build commercial or proprietary products using Copendex, subject to the terms of the Apache License 2.0.
