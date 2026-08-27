# Changelog

## 1.0.0 - 2026-08-26

- Pin InfluxDB 3 Core v3.10.0 to OCI index `sha256:b3e577f38c19963597170d8850a3a7f77af8f0cfa866c64cd13e5de0f238e114`.
- Add a public bearer-token adapter and unauthenticated `/healthz`.
- Add descriptor-confined root initialization and permanent UID/GID 1500 privilege drop.
- Persist the protected internal admin token and database at `/data`.
- Configure one replica, `PORT=8181`, Serverless disabled, and Railway health checks.
- Document cold backup/restore and restore-based rollback. No automatic upgrade or safe in-place downgrade is claimed.
