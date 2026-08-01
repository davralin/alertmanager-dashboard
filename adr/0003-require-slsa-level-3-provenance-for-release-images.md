# 0003. Require SLSA Level 3 Provenance for Release Images

Date: 2026-07-31

## Status

Accepted

## Context

Release container images are deployment artifacts. They should be traceable to source, workflow, and build identity before promotion into deployment repositories.

Tags are convenient for humans but mutable. Provenance verification is most useful when deployment consumes immutable artifacts.

## Decision

Release container images require SLSA Level 3 provenance.

Use the SLSA GitHub generator workflow for release images.

Enable BuildKit `provenance: true` and `sbom: true` for release image builds.

Digest-pin normal GitHub Actions.

Keep `slsa-framework/slsa-github-generator` reusable workflows tag-pinned, not digest-pinned, because verifier-compatible provenance depends on the workflow ref.

Release notes must include the image digest and a SLSA verification command.

Prefer deploying verified immutable image digests:

```text
ghcr.io/OWNER/REPO@sha256:...
```

Repos that do not publish release container images should add a repo-specific ADR documenting why this decision is not applicable.

## Consequences

Release images can be verified against repository source and tag.

Most workflow dependencies remain immutable while the SLSA generator keeps verifier-compatible identity claims.

Deployment diffs are explicit when image digests change.

Release workflows must maintain SLSA generation and verification instructions.
