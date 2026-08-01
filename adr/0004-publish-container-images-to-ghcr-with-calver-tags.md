# 0004. Publish Weekly CalVer Container Releases to GHCR

Date: 2026-07-31

## Status

Accepted

## Context

Container images need a continuous release flow so produced artifacts stay maintained and patched.

Operational tools often need refreshed base images and dependency updates even when application code has not changed.

Rollback needs clear human-readable release points, while deployments should still consume immutable image digests.

## Decision

Publish container images to GHCR.

Publish weekly CalVer container releases.

Use release tags in `YYYY.MM.DD` format.

Release images should publish `linux/amd64,linux/arm64`.

Publishing `latest` is allowed for repos that want a moving human-friendly reference.

## Consequences

Images are continuously rebuilt and kept current with dependency, base image, and security updates.

CalVer tags provide clear rollback points and readable release identifiers.

Deployments can still consume immutable digests for exact artifact promotion.

Repos with library-style versioning can choose a different policy and document it in their own ADR.
