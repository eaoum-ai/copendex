# 0002 Use Text-Based Project Management

## Status

Accepted

## Context

Cosha is an OSS project with a small early scope. The project needs visible planning without introducing a heavyweight external project management dependency.

## Decision

Keep roadmap, backlog, milestones, release planning, changelog, and architecture decisions in Markdown files committed to the repository.

## Consequences

- Planning changes can be reviewed in pull requests.
- Project history stays close to the code.
- Contributors can understand direction without access to private tools.
- GitHub Issues can still be used for public discussion when needed, but repository text files remain the canonical lightweight planning source.
