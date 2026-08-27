# Template Packaging Inventory

| Field | Answer |
|---|---|
| Source type | Official public Docker image plus local wrapper |
| Source | `influxdb:3.10.0-core` at immutable OCI index |
| Railway services | One application service |
| Start command | Image `ENTRYPOINT ["/usr/local/bin/init.sh"]` |
| Health check | `GET /healthz` on `PORT=8181` |
| Public routing | HTTP through authenticated adapter |
| Private networking | Not needed; upstream binds loopback only |
| Native databases/Redis/Buckets | None |
| Volume | `/data`, one replica |
| Required variables | `PORT`; generated external bearer |
| Optional variables | None |
| User-supplied secrets | None |
| Config as code | `Dockerfile`, scripts, Go wrappers, `railway.json` |
| Docs/product kit | README, changelog, operations, environment, support, provenance, marketplace |
| Issue intake | Bug form, config routing, labels |
| Security/legal | SECURITY, MIT wrapper license, NOTICE, TRADEMARKS |
| Asset | Unmodified upstream SVG with provenance; no screenshot/demo |
| Update strategy | Manual immutable-pin update after full revalidation |
| Support burden | Medium; single service, but stateful backup/migration boundaries |

Exact file count and package digest are finalized in `docs/build-report.md` and
`PACKAGE_SHA256SUMS`.
