# InfluxDB 3 Core

A single-service Railway package for InfluxDB 3 Core v3.10.0 with persistent storage, a generated public bearer token, and a protected volume-only admin token.

This is an independent community template. It is not endorsed by InfluxData.

[![Deploy InfluxDB 3 Core](https://railway.com/button.svg)](https://railway.com/deploy/influxdb-3-core)

## Architecture

| Component | Purpose |
|---|---|
| InfluxDB 3 Core | Stores and queries time-series data on `/data` |
| Authenticated adapter | Listens on `PORT=8181`, validates the external bearer, and forwards requests to loopback |
| Railway volume | Persists the database and internal admin token at `/data` |

The container starts briefly as root only to validate `/data`, safely create or validate the internal token, and set exact ownership. It permanently drops to UID/GID `1500:1500` before starting InfluxDB and the adapter. The exact upstream OCI index is pinned in `Dockerfile`.

## Deploy from the Railway template

Use the canonical [InfluxDB 3 Core template](https://railway.com/deploy/influxdb-3-core).

The template creates one service and one volume mounted at `/data`. It sets:

- `PORT=8181`
- `INFLUXDB3_EXTERNAL_BEARER_TOKEN=${{secret(64)}}`

No user input is required. Save the generated external bearer token after deployment; Railway keeps it in the service variables.

The service is ready when Railway reports `GET /healthz` healthy. The first start can take a few minutes while the image builds and the internal token is generated.

## Deploy this repository with Railway CLI

Repository-only `railway up` does not create the marketplace variable contract for you. Set both required variables explicitly and attach a volume before deploying:

```bash
railway init --name influxdb3-core
railway add --service influxdb3-core --json
railway service link influxdb3-core
railway volume add --mount-path /data --json
railway variable set --service influxdb3-core PORT=8181 INFLUXDB3_EXTERNAL_BEARER_TOKEN="$(openssl rand -hex 32)"
railway up --service influxdb3-core
railway domain --service influxdb3-core --port 8181 --json
```

The final command creates the Railway public domain used as `INFLUX_URL`
below. The generated token value is 64 hexadecimal characters. Never reuse
the internal token stored in `/data/admin-token.json`.

## Use the API

Set your public Railway URL and the external token from Railway Variables:

```bash
export INFLUX_URL="https://your-domain.example"
export INFLUX_TOKEN="your-generated-external-token"

curl -fsS -H "Authorization: Bearer ${INFLUX_TOKEN}" \
  "${INFLUX_URL}/health"

curl -fsS -X POST \
  -H "Authorization: Bearer ${INFLUX_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"db":"metrics"}' \
  "${INFLUX_URL}/api/v3/configure/database"

curl -fsS -X POST \
  -H "Authorization: Bearer ${INFLUX_TOKEN}" \
  --data-binary 'cpu,host=demo usage=42.5' \
  "${INFLUX_URL}/api/v3/write_lp?db=metrics&precision=nanosecond"

curl -fsS -X POST \
  -H "Authorization: Bearer ${INFLUX_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"db":"metrics","q":"SELECT * FROM cpu","format":"json"}' \
  "${INFLUX_URL}/api/v3/query_sql"
```

`/healthz` is intentionally unauthenticated for Railway health checks. All normal upstream routes, including `/health`, require the external bearer token. The internal admin token is rejected at the public adapter.

## Persistence and restarts

All database state and the internal admin token live on `/data`. A normal restart or redeploy reuses the same volume and token. Keep exactly one replica: Railway volumes are not multi-writer storage and this package does not provide a cluster.

## Backup, restore, and updates

Backups are offline operator procedures. Stop writes and stop the service before taking a Railway volume backup or cold copy. Restore into a fresh volume while the service is stopped, then start and verify health, catalog, and representative queries. See [operations](docs/operations.md) for the complete checklist.

InfluxDB 3 Core may perform one-way data migrations. Before changing the pinned image, create and verify a cold backup. Test the new version against a restored copy first. Rollback means restoring the pre-upgrade cold backup with the prior image; it is not an automatic in-place downgrade.

## Security

- Rotate the external bearer by replacing `INFLUXDB3_EXTERNAL_BEARER_TOKEN` with a new value of at least 32 characters, then redeploy.
- Do not publish service variables, `/data/admin-token.json`, logs containing request headers, or database backups.
- Do not expose port `8182`; the upstream database binds only to loopback inside the container.
- Report wrapper vulnerabilities privately as described in [SECURITY.md](SECURITY.md).

## Limits

- Single-node InfluxDB 3 Core; no high availability or horizontal replicas.
- One Railway service plus one volume; no GUI is included.
- Railway Serverless is disabled because databases must remain available and background/outbound activity can prevent sleep.
- Capacity depends on workload. The tested contract stayed below 1 GiB during bounded local and Railway smoke tests, but that is not a production sizing guarantee.
- Backup/restore, migration, and rollback require operator action and downtime.

## Documentation

- [Environment variables](docs/environment.md)
- [Operations, backups, updates, and rollback](docs/operations.md)
- [Support boundary](docs/support.md)
- [Source provenance](docs/source-provenance.md)
- [Railway template contract](docs/railway-template-contract.md)

The wrapper is MIT-licensed. InfluxDB 3 Core retains its upstream licenses; see [NOTICE.md](NOTICE.md).
