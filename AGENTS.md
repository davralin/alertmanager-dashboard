# Agent instructions for alertmanager-dashboard

## Repository standards

- Read `adr/` before changing workflows, release policy, Renovate, container behavior, or deployment assumptions.
- Treat ADRs as inherited decisions. Do not delete ADRs to make them not apply; add a later ADR that supersedes, narrows, or marks a decision not applicable.
- Keep Git history and commit messages aligned with ADR 0002: linear history and Conventional Commits.
- This repository intentionally publishes separate `receiver` and `dashboard` images as documented in ADR 0009.
- Keep normal GitHub Actions digest-pinned. Keep the SLSA generator workflow tag-pinned as documented in ADR 0003.
- Keep Renovate behavior aligned with ADR 0005 and the existing commit naming style.
- Target Kubernetes PSA `restricted` for workload guidance unless a later ADR documents an exception.

## CI / GitHub Actions

### SHA pinning
All third-party actions in all workflows must be pinned to a commit SHA with a version comment:

```yaml
uses: docker/login-action@c94ce9fb468520275223c153574b00df6fe4bcc9  # v3
```

When bumping an action version, resolve the new commit SHA:

```sh
curl -sf "https://api.github.com/repos/<owner>/<action>/tags?per_page=5"
```

Update both the SHA and the version comment.

### CI vs Release split
Two workflows build container images with different attestation levels:

- **`ci.yml`** - triggered on push to `main` and PRs. Builds test images tagged `:latest`, `:main`, and `:sha-xxxxx` without SLSA provenance or SBOM. Trivy scans run and publish SARIF.
- **`release.yml`** - triggered on schedule (Monday 09:00 UTC) or `workflow_dispatch`. Builds attested images with full SLSA L3 provenance and SBOM, tagged with the CalVer date (`YYYY.MM.DD`) and `:latest`. Creates the GitHub Release only after image build and provenance generation succeed. Vulnerability scans publish SARIF but are advisory by default.

### slsa-github-generator tag pin exception
`slsa-framework/slsa-github-generator` must be pinned by version tag, not SHA:

```yaml
uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@v2.1.0
```

The generator embeds the `workflow_ref` from its OIDC token into the provenance certificate. `slsa-verifier` expects a versioned tag in that claim; a SHA pin produces a non-verifiable certificate. Do not "fix" this to a SHA.

### Provenance jobs
The `provenance-receiver` and `provenance-dashboard` jobs exist only in `release.yml`. They are reusable workflow calls (`uses:`) and cannot contain `steps:`. Image-build logic belongs in the image jobs, which expose registry digests via outputs.

### Trivy policy
`severity: CRITICAL,HIGH`, `ignore-unfixed: true`. Trivy scans publish SARIF for GitHub code scanning, but do not set `exit-code: '1'`; vulnerability tracking is handled through SARIF alerts rather than blocking PRs/releases.

### Weekly scan
`weekly-scan.yml` scans `ghcr.io/davralin/alertmanager-dashboard-receiver:latest` and `ghcr.io/davralin/alertmanager-dashboard-dashboard:latest` every Sunday 09:00 UTC. It is independent of the release workflow and does not gate Monday's release.

### SBOM
`sbom: true` on `docker/build-push-action` generates a Syft SBOM and attaches it as an OCI attestation alongside each image in GHCR. Inspect with:

```sh
docker buildx imagetools inspect ghcr.io/davralin/alertmanager-dashboard-receiver:latest
```

### Containerfile HEALTHCHECK
Every release `Containerfile` stage must include a `HEALTHCHECK` instruction. For single-shot containers, `HEALTHCHECK NONE` is the accepted value.

## Go / tests

- Keep Go source formatted with `gofmt`.
- Run `go vet ./...` after behavior or API changes.
- Run `go test ./...` after behavior changes.
- Do not introduce network-dependent unit tests unless they are isolated behind explicit integration-test setup.

## Local validation

- Run `git diff --check` before committing.
- Run the CI gofmt check after Go changes: `git ls-files -z '*.go' | xargs -0 -r gofmt -l`.
- Run `go vet ./...` after Go changes.
- Run `go test ./...` after behavior changes.
- Run `docker build -f Containerfile --target receiver -t alertmanager-dashboard-receiver:local .` after changing `Containerfile` or image workflows when practical.
- Run `docker build -f Containerfile --target dashboard -t alertmanager-dashboard-dashboard:local .` after changing `Containerfile` or image workflows when practical.

## Git

- Do not commit, amend, or push unless explicitly requested.
