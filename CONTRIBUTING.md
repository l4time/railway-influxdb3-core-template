# Contributing

Changes should preserve the documented one-service, one-volume contract.

Before proposing a change:

1. Do not include tokens, backups, database contents, or Railway account data.
2. Run `python3 tests/static_check.py` and `python3 tests/claim_check.py`.
3. Build the image twice without cache and compare the canonical rootfs digest
   when changing runtime files or a base image.
4. Re-run auth negatives, write/query, restart persistence, cold restore, and
   restore-based rollback tests for runtime changes.
5. Update `CHANGELOG.md`, provenance, inventory, and checksums.

Do not loosen the UID/GID 1500 privilege drop, expose loopback port 8182,
persist the external bearer, enable Serverless, or promise safe in-place
downgrades without new independent evidence.
