# 0006. Define Container Images by Deployable Process Responsibility

Date: 2026-07-31

## Status

Accepted

## Context

This repository provides an active single-image workflow as the default operational path.

Container image topology affects permissions, dependencies, scaling, release notes, provenance, vulnerability scanning, and deployment ownership.

## Decision

Define container images by deployable process responsibility.

Default to one image when the repository produces one deployable process or tool.

Use multiple images when the repository contains independently deployed process types with different responsibilities, entrypoints, dependencies, permissions, scaling profiles, or security boundaries.

Do not create multiple images only because the repository could technically build them.

Do not force unrelated deployables into one image for workflow simplicity.

Workflow topology follows the repository ADRs that define deployable units.

## Consequences

This repository focuses on the single-image path.

Multiple images require a later repo-specific ADR explaining the separate deployable responsibilities.

Workflow, release, provenance, and scan configuration must match the declared deployable units.
