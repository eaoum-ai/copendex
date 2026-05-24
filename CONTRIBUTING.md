# Contributing

Thanks for your interest in contributing to Copendex.

## Development

Copendex uses trunk-based development. Create short-lived branches from `main`, keep pull requests small, follow the branch naming standard, and merge through the protected `main` branch after review and checks pass.

See [docs/development.md](docs/development.md) for the full development strategy.

Run the standard checks before opening a pull request:

```sh
make check
```

You can also run the individual steps:

```sh
make test
make build
```

## Guidelines

- Keep Copendex local-first.
- Do not add cloud, SaaS, LLM, or embedding dependencies to the core v0 CLI.
- Prefer deterministic behavior and structured JSON output.
- Keep dependencies small and cross-platform.
- Add focused tests for parser, indexing, and search behavior.

## Developer Certificate of Origin

Copendex uses the Developer Certificate of Origin (DCO) for contributions.

By contributing to this project, you certify that you have the right to submit your contribution and that it can be licensed under the Apache License 2.0.

All commits must include a `Signed-off-by` line.

You can add this automatically by committing with:

```bash
git commit -s -m "Your commit message"
```

The sign-off should look like:

```text
Signed-off-by: Your Name <your.email@example.com>
```

See the Developer Certificate of Origin at https://developercertificate.org/.

## Pull Requests

Please include:

- A concise summary of the change.
- Tests for behavior changes where practical.
- Notes about any new dependency or platform assumption.
- A [CHANGELOG.md](CHANGELOG.md) entry for user-visible changes.

## Releases

Release and version management is documented in [docs/releases.md](docs/releases.md).
