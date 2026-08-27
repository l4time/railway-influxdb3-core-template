# QA checklist

- [ ] Exact runtime files match `SOURCE_SHA256SUMS`.
- [ ] Entire package matches `PACKAGE_SHA256SUMS`.
- [ ] Dockerfile pins both OCI indices and builds without network-time drift.
- [ ] `railway.json` uses Dockerfile, `/healthz`, one replica, `ON_FAILURE`,
      and Serverless disabled.
- [ ] Template creates one volume at `/data`.
- [ ] Template sets `PORT=8181` and generates `${{secret(64)}}`.
- [ ] Missing, weak, and internal-equal external tokens fail closed.
- [ ] PID 1 and steady state run as UID/GID 1500 after narrow root init.
- [ ] `/healthz` succeeds without auth.
- [ ] `/health` and API routes reject absent/wrong/internal tokens.
- [ ] External token can create a database, write line protocol, and query SQL.
- [ ] Data and internal token survive restart on the same volume.
- [ ] Cold backup restores into a genuinely fresh volume.
- [ ] Restore-based rollback excludes post-backup changes.
- [ ] No external or internal raw token appears in package evidence or logs.
- [ ] README, legal/security files, issue intake, support, and marketplace copy
      contain no unsupported HA, auto-update, cost, or downgrade claims.
- [ ] Runtime/package checks and independent product-kit QA report zero blockers.
