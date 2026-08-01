# 0001. Use ADRs for Repository Decisions

Date: 2026-07-31

## Status

Accepted

## Context

This repository contains operational and application decisions that affect future maintenance: release policy, security scanning, provenance, deployment conventions, runtime choices, and other durable defaults.

Keeping rationale only in README files, workflow comments, or commit messages makes decisions hard to find and easy to lose.

## Decision

Use Architecture Decision Records for durable repository decisions.

Use ADRs for decisions that affect how this repo is built, released, secured, deployed, or operated.

Do not use ADRs for routine implementation details, typo fixes, dependency bumps, or short-lived tasks.

## Consequences

Important decisions are explicit and reviewable.

Changing a durable default should update or supersede an ADR.

The README and operational docs can stay concise.
