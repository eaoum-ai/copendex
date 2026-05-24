# Release and Version Management

Copendex uses calendar versioning for public releases.

## Version Format

Use `YY.MM.N`:

```text
YY.MM.N
```

- `YY`: two-digit release year.
- `MM`: two-digit release month.
- `N`: release number for that month, starting at `0`.

Examples:

```text
26.05.0
26.05.1
26.06.0
```

The version number communicates release timing, not compatibility. Compatibility expectations are documented in the compatibility policy and changelog.

## Compatibility Policy

Copendex treats the local index as rebuildable state. An index schema change may require users to run:

```sh
copendex index
```

or a future explicit rebuild command.

That policy only applies to local index data under `.copendex/index/`. It should not be used as a reason to break user-facing contracts casually.

Higher-care compatibility surfaces include:

- CLI command names, flags, and exit behavior.
- JSON output fields consumed by agents.
- `.copendex/config.yaml` shape and defaults.
- Future MCP tools, resources, and response schemas.

Breaking changes to those surfaces must be documented clearly in [CHANGELOG.md](../CHANGELOG.md), because the `YY.MM.N` version alone does not communicate compatibility.

## Next Phase: MCP Compatibility Layer

Copendex should add an MCP server after the CLI foundation is stable.

The MCP server should provide a stable agent-facing contract over Copendex's local index and should handle compatibility concerns that are awkward for direct CLI/JSON consumers:

- Expose versioned tool and response schemas.
- Report index schema compatibility and required re-index actions.
- Return structured errors when the local index is missing, stale, or incompatible.
- Preserve backward-compatible agent workflows where practical.
- Keep all codebase data local to the developer's machine.

## Tags

Release tags should use a leading `v`:

```text
v26.05.0
v26.05.1
v26.06.0
```

The tag should point to a commit on `main`.

## Changelog

User-visible changes should be recorded in [CHANGELOG.md](../CHANGELOG.md) under `Unreleased`.

Use these sections when applicable:

- `Added`
- `Changed`
- `Deprecated`
- `Removed`
- `Fixed`
- `Security`

At release time:

1. Move entries from `Unreleased` into a new version section.
2. Add the release date in `YYYY-MM-DD` format.
3. Commit the changelog update.
4. Tag the release commit.
5. Push the tag.

## Release Checklist

1. Confirm `main` is green in CI.
2. Run local checks:

   ```sh
   make check
   ```

3. Update [CHANGELOG.md](../CHANGELOG.md).
4. Choose the next `YY.MM.N` version.
5. Create a signed release commit if needed:

   ```sh
   git commit -s -m "Prepare v26.05.0 release"
   ```

6. Create and push the tag:

   ```sh
   git tag v26.05.0
   git push origin main
   git push origin v26.05.0
   ```

7. Create GitHub release notes from the changelog.

GoReleaser can be added later for automated Linux, macOS, and Windows binaries.
