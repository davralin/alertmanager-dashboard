# 0009. Define Alertmanager Dashboard Images

Date: 2026-08-01

## Status

Accepted

## Context

ADR 0006 defaults repositories to a single image unless multiple deployable process responsibilities justify separate images.

This repository contains two long-running processes with different responsibilities:

- `receiver` accepts Alertmanager webhook payloads and writes alert state to Valkey.
- `dashboard` reads alert state from Valkey and serves the dashboard UI plus API.

The processes have different entrypoints, deployment surfaces, scaling concerns, and network exposure. Combining them into one image would make workflow topology simpler but would blur deployable ownership and runtime responsibilities.

## Decision

Publish two release images:

- `ghcr.io/davralin/alertmanager-dashboard-receiver`
- `ghcr.io/davralin/alertmanager-dashboard-dashboard`

Build both images from `Containerfile` using the `receiver` and `dashboard` targets.

Keep CI, release, provenance, scan, and documentation behavior aligned with these two deployable units.

## Consequences

The repository intentionally narrows ADR 0006's single-image default.

Each image has its own release artifact, digest, SLSA provenance, SBOM, SARIF scan category, and deployment reference.

Future deployable process additions require an ADR update before adding more release images.
