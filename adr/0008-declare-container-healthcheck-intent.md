# 0008. Declare Container Healthcheck Intent

Date: 2026-08-01

## Status

Accepted

## Context

Container scanners and repository quality checks commonly flag images that do not define a `HEALTHCHECK`. That is useful for long-running services, but not every released image is a service. Some images are batch jobs, one-shot tools, workflow exercisers, or intentionally minimal artifacts with no process to probe.

An omitted healthcheck is ambiguous. It can mean the image was not considered, or it can mean a healthcheck is not applicable.

## Decision

Every release `Containerfile` should declare healthcheck intent.

Long-running service images should define a real `HEALTHCHECK` that verifies the container can serve its intended function.

Batch, one-shot, tooling, and workflow-only images should declare `HEALTHCHECK NONE`.

Do not add fake healthchecks only to satisfy scanners.

## Consequences

Container healthcheck behavior is explicit in image metadata.

Quality findings for missing healthchecks are addressed without adding meaningless runtime probes.

Service images still need real probes, and non-service images document that no probe applies.
