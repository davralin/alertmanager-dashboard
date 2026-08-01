# 0002. Use Linear History and Conventional Commits

Date: 2026-08-01

## Status

Accepted

## Context

Repository history is part of the maintenance interface. Release automation, changelogs, Renovate grouping, rollback review, and incident analysis all benefit from predictable commit structure.

Merge commits and squash commits hide useful information in different ways. Merge commits add noisy topology for small operational repositories. Squash commits can discard commit-level authorship and reviewed intermediate changes.

A linear rebase-merged history keeps individual commits while avoiding merge bubbles.

## Decision

Use a linear Git history.

Pull requests should be merged with rebase merges. Do not use merge commits or squash commits for normal repository changes.

Use Conventional Commits for commit messages:

- `feat:` for new behavior or capabilities
- `fix:` for corrections
- `chore:` for maintenance, dependency, digest, workflow, and metadata updates
- `docs:` for documentation-only changes
- `refactor:` for behavior-preserving restructuring
- `test:` for test-only changes
- `build:` for build system or packaging changes
- `ci:` for CI workflow behavior changes

Use `!` for breaking changes, for example `feat(container)!:`.

Renovate commit and PR titles should follow the same convention.

Renovate automerge should use PR automerge with rebase merge strategy when automerge is enabled.

GitHub repository settings should enforce this policy by allowing rebase merges and disabling merge commits and squash merges.

## Consequences

History remains linear and reviewable.

Automation and humans use the same commit language.

GitHub repository settings are the enforcement point for human merges. Repository config records the policy and Renovate follows it where possible.

Repos that need a different history model should add a later ADR explaining why.
