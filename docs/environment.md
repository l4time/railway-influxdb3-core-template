# Environment variables

## Required

| Variable | Template value | Purpose |
|---|---|---|
| `PORT` | `8181` | Public authenticated adapter port and Railway routing target |
| `INFLUXDB3_EXTERNAL_BEARER_TOKEN` | `${{secret(64)}}` | Public API credential; must contain at least 32 characters |

The marketplace template generates the external token. A repository-only
`railway up` must set both variables explicitly.

## Internal state

`/data/admin-token.json` is created on the volume during first boot. It is the
upstream InfluxDB admin token and is readable only by UID/GID 1500 with mode
0600. It is not an environment variable, must not be copied into Railway
Variables, and must never equal the external token.

## Unsupported overrides

The public port, loopback upstream port, node ID, object-store mode, data path,
and UID/GID are intentionally fixed by the wrapper. Do not set InfluxDB
arguments through environment variables or expose port 8182.

## Rotation

Replace the external token with a new random value of at least 32 characters
and redeploy. Existing API clients must then use the new value. External-token
rotation does not rewrite the internal volume token.
