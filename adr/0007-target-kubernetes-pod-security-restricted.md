# 0007. Target Kubernetes Pod Security Restricted

Date: 2026-07-31

## Status

Accepted

## Context

Kubernetes workloads should have a safe baseline before app-specific exceptions are considered.

Pod Security Admission applies built-in Kubernetes Pod Security Standards at namespace admission time.

The `restricted` profile is the most locked-down standard profile. It prevents common privilege escalation paths and keeps deployment manifests compatible with hardened namespaces.

In practice, this means a normal application pod should not need root, privileged containers, host networking, host PID/IPC namespaces, broad Linux capabilities, or writable container root filesystems. If an image only works as root or expects to mutate its installed filesystem, that is usually an image/runtime problem to fix rather than a deployment default to accept.

For many application images, creating a dedicated application user with a fixed UID/GID and ending the `Containerfile` with `USER <that-user>` gets most of the way there. The application should write only to declared data/cache paths, not to the installed application tree.

This is not meant to block legitimate infrastructure workloads. Some controllers, node agents, storage components, and debugging tools need privileges that application workloads should not have. Those exceptions should be deliberate and documented.

Reference: <https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted>

## Decision

Target Kubernetes Pod Security Admission `restricted` for workloads by default.

Workloads should run as non-root, avoid privileged mode, avoid host namespaces, drop unnecessary capabilities, use read-only root filesystems where practical, and avoid hostPath mounts unless a later ADR justifies an exception.

Any workload that cannot meet `restricted` must document the reason and compensating controls in a repo-specific ADR.

## Consequences

Default manifests are safer to deploy into hardened clusters.

Security exceptions become explicit architecture decisions.

Some images may need runtime or packaging changes before they can be deployed.

Manifest authors should think about security context early, not after deployment fails admission.
