# Build report

Status: **PASS — ready for independent product-kit QA**

Built and checked: 2026-08-26.

## Runtime identity

- InfluxDB 3 Core: v3.10.0
- OCI index: `sha256:b3e577f38c19963597170d8850a3a7f77af8f0cfa866c64cd13e5de0f238e114`
- Runtime source: `a1e8994464c3fe0b44ee85e95c0714ad557ed7fc`
- Exact R5 wrapper files: six of six byte-identical
- R5 canonical rootfs:
  `93aa230f259dec70c900a5ca1859de8d266edfb6ce42a5754c24bf38c2a96cfa`

## Package checks

- `sha256sum -c SOURCE_SHA256SUMS`: PASS
- `python3 tests/claim_check.py`: PASS
- Two independent `--pull --no-cache --provenance=false` builds: PASS
- Independent image IDs differed, as exporter metadata can, while both canonical
  rootfs digests equaled the R5 value: PASS
- Owned test containers, tags, and temporary digest files removed: PASS
- Stable content inventory excluding this report and its checksum manifest:
  36 files, framed SHA-256
  `25d05fe735c89be3fba45c693c06fbe8d8b24362017a29e7b57c249b084b2385`

`PACKAGE_SHA256SUMS` covers every public package file except itself. The static
checker verifies the exact inventory, R5 source hashes, Railway configuration,
package checksums, and a final framed digest over all 38 public files.

## Prior proof carried into QA

The retained R5 local evidence passed authentication negatives, external-to-
internal bearer substitution, database creation, write/query, restart
persistence, fresh cold restore, restore-based rollback, leakage scans, UID/GID
1500 steady state, resource ceiling, and cleanup. Railway R2 independently
passed the same product contract with peak 226.76 MB and 0.07926 vCPU, followed
by forced cleanup.

No Railway, GitHub, repository, draft, or public-template resource was created
during packaging.
