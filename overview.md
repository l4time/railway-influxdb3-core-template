# Deploy and Host

Deploy InfluxDB 3 Core v3.10.0 as one authenticated service with a persistent
`/data` volume.

## About Hosting

This package generates a 64-character external bearer token and creates a
separate internal admin token that never leaves the volume. It exposes an
unauthenticated `/healthz` endpoint only for deployment health checks and runs
the steady-state database as UID/GID 1500.

## Why Deploy

Use a compact, single-node InfluxDB 3 Core service when you need authenticated
time-series ingestion and SQL queries with persistent storage. The external
bearer token is available to the deployer, while the internal administrative
credential remains separated inside `/data`.

## Common Use Cases

- Collect application, infrastructure, or device metrics.
- Store persistent event and telemetry data.
- Query recent and historical time-series data with SQL.
- Back a small monitoring or analytics workflow with an authenticated API.

This package does not include a GUI, high availability, multi-node clustering,
or automatic upgrades. Railway Serverless is intentionally disabled. Offline
backup/restore and version migration remain operator procedures.

## Dependencies for

### Deployment Dependencies

- One InfluxDB 3 Core service.
- One persistent volume mounted at `/data`.
- One generated `INFLUXDB3_EXTERNAL_BEARER_TOKEN`.

This is an independent community template and is not endorsed by InfluxData.
