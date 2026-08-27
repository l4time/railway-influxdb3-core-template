# Source provenance

## Runtime

| Item | Immutable identity |
|---|---|
| Upstream project | `https://github.com/influxdata/influxdb` |
| Runtime version | `3.10.0` |
| Runtime source revision | `a1e8994464c3fe0b44ee85e95c0714ad557ed7fc` |
| Official image | `influxdb:3.10.0-core` |
| OCI index | `sha256:b3e577f38c19963597170d8850a3a7f77af8f0cfa866c64cd13e5de0f238e114` |
| Tested linux/arm64 manifest | `sha256:3ea27fec463ff619e7aff1659c0bb67c1b254b4b69fa089c0b45eca42eeb08f7` |
| Official Docker build-source revision | `ddd58c7ba3c15041c5909d54efcf11db322d8250` |
| Go builder index | `golang@sha256:d9132cce84391efab786495288756d60e1da215b1f94e87860aeefc3d4c45b6d` |

The runtime source revision and official Docker build-source revision identify
different upstream repositories and are intentionally not equated.

## Wrapper provenance

`Dockerfile`, `adapter.go`, `init.sh`, `supervisor.sh`, `secureinit.go`, and
`testclient.go` are byte-identical to the retained R5 local proof sources. The
corresponding hashes are recorded in `SOURCE_SHA256SUMS`.

The R5 proof produced canonical rootfs digest
`93aa230f259dec70c900a5ca1859de8d266edfb6ce42a5754c24bf38c2a96cfa`
in two independent no-cache builds. Container exporter ordering and timestamps
are excluded from that canonical digest.

## Graphic

The upstream graphic source, commit, retrieval date, and trademark caveat are
recorded in `assets/README.md`.
