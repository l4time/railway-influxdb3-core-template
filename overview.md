# InfluxDB 3 Core

Deploy InfluxDB 3 Core v3.10.0 as one authenticated Railway service with a persistent `/data` volume.

The package generates a 64-character external bearer token, creates a separate internal admin token that never leaves the volume, exposes an unauthenticated `/healthz` only for Railway health checks, and runs the steady-state database as UID/GID 1500.

## What it creates

- One InfluxDB 3 Core service
- One Railway volume mounted at `/data`
- One generated `INFLUXDB3_EXTERNAL_BEARER_TOKEN`

## Good fit

Use this for persistent single-node time-series ingestion and SQL queries where a simple authenticated API is enough.

This package does not include a GUI, high availability, multi-node clustering, or automatic upgrades. Railway Serverless is intentionally disabled. Offline backup/restore and version migration remain operator procedures.

This is an independent community template and is not endorsed by InfluxData.
