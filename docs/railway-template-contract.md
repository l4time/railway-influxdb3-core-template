# Railway template contract

| Field | Value |
|---|---|
| Display name | `InfluxDB 3 Core` |
| Public template code | `influxdb-3-core` |
| Public deploy URL | `https://railway.com/deploy/influxdb-3-core` |
| Upstream | `influxdata/influxdb` |
| Version | `3.10.0` |
| Services | One application service |
| Volume | One volume mounted at `/data` |
| Public port | `8181` |
| Health | `GET /healthz`, unauthenticated |
| Required variables | `PORT=8181`; `INFLUXDB3_EXTERNAL_BEARER_TOKEN=${{secret(64)}}` |
| User-supplied variables | None |
| Replicas | One |
| Serverless | Disabled |
| Restart | `ON_FAILURE`, maximum 10 retries |
| Steady-state identity | UID/GID `1500:1500` |
| Internal database bind | `127.0.0.1:8182` |
| Persistence | Database and internal admin token under `/data` |

The public adapter rejects missing, malformed, wrong, and internal bearer
tokens. It substitutes the protected internal bearer only after constant-time
external-token validation. `/healthz` is the sole no-auth route.

Template-editor wiring must create the volume and generator exactly as shown
above. Repository-only deployment must set both variables explicitly.
