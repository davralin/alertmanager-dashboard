# alertmanager-dashboard

Alertmanager webhook receiver plus a small dashboard for active alerts.

The app is split into two independent Go binaries and two independent container
images:

- `receiver`: receives Alertmanager webhooks at `POST /webhook` and writes state to Valkey.
- `dashboard`: reads Valkey and serves the HTML dashboard plus `GET /api/state`.

`Watchdog` webhooks update `last_ping` but are not shown as active alerts.

## State

Valkey keys:

- `alertmanager-dashboard:last_ping`: RFC3339 timestamp of the last accepted webhook.
- `alertmanager-dashboard:alerts`: hash of active alerts keyed by Alertmanager fingerprint.

Resolved alerts are removed from the hash. If Alertmanager does not provide a
fingerprint, the receiver derives one from sorted labels.

## Configuration

Both binaries use the same environment variables:

- `LISTEN_ADDR`: HTTP listen address, default `:8080`.
- `VALKEY_ADDR`: Valkey address, default `127.0.0.1:6379`.
- `VALKEY_USERNAME`: optional Valkey username.
- `VALKEY_PASSWORD`: optional Valkey password.
- `VALKEY_DB`: Valkey database number, default `0`.
- `STALE_AFTER`: dashboard last-ping stale threshold, default `90m`.
- `HEALTHCHECK_URL`: optional override for the container healthcheck URL.

## Local Development

Run tests:

```sh
go test ./...
```

Run with docker compose:

```sh
docker compose up --build
```

Send a firing alert:

```sh
curl -fsS -X POST http://127.0.0.1:8080/webhook \
  -H 'Content-Type: application/json' \
  --data-binary @testdata/webhook-firing.json
```

Send a Watchdog heartbeat:

```sh
curl -fsS -X POST http://127.0.0.1:8080/webhook \
  -H 'Content-Type: application/json' \
  --data-binary @testdata/webhook-watchdog.json
```

Open the dashboard at <http://127.0.0.1:8081/>.

## Images

Build locally:

```sh
docker build -f Containerfile --target receiver -t alertmanager-dashboard-receiver .
docker build -f Containerfile --target dashboard -t alertmanager-dashboard-dashboard .
```

Published image names are:

- `ghcr.io/davralin/alertmanager-dashboard-receiver`
- `ghcr.io/davralin/alertmanager-dashboard-dashboard`

Final images are `scratch`, run as UID/GID `1000`, and use binary-backed
container healthchecks.
