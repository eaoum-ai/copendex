# Development Strategy

Copendex uses trunk-based development.

## Principles

- `main` is the trunk and should always be releasable.
- Work happens in short-lived branches created from `main`.
- Changes merge through pull requests.
- Direct pushes to `main` are blocked by branch protection.
- Pull requests should be small enough to review quickly.
- Large work should be split into incremental, working changes.
- Incomplete behavior should stay behind config, commands, or internal paths that do not disrupt existing workflows.

## Branches

Use short-lived branches with descriptive names.

Branch names must follow this format:

```text
<type>/<short-kebab-description>
```

Allowed branch types:

- `feature/`: user-visible features or larger capabilities.
- `fix/`: bug fixes.
- `docs/`: documentation-only changes.
- `test/`: test-only changes.
- `refactor/`: behavior-preserving code restructuring.
- `chore/`: maintenance, tooling, CI, or dependency work.
- `release/`: release preparation.
- `hotfix/`: urgent fixes based from `main`.

Rules:

- Use lowercase kebab-case after the slash.
- Keep names short and descriptive.
- Do not include author, tool, vendor, or agent prefixes.
- Do not use branches such as `codex/...`, `agent/...`, `user/...`, or personal name prefixes.
- Every contributor and automation tool follows the same naming standard.

Examples:

```text
feature/java-symbols
fix/gitignore-scanning
docs/release-process
chore/dco-workflow
release/26.05.0
hotfix/index-schema-check
```

Delete branches after merge.

Long-running release, develop, integration, or stabilization branches should be avoided unless there is a specific temporary need.

## Pull Requests

Before opening a pull request:

```sh
make check
```

Pull requests should include:

- A concise description of the change.
- Tests for behavior changes where practical.
- Documentation updates for user-visible behavior.
- A changelog entry when the change affects users, packaging, CLI behavior, configuration, compatibility, or contributor workflow.

## Merge Expectations

The protected `main` branch requires:

- Pull request review.
- Passing CI.
- Passing DCO check.
- Resolved conversations.
- Up-to-date branch status before merge.

Squash merge is preferred for small pull requests if repository settings allow it. The final commit message should be clear and should preserve the DCO sign-off where required.
