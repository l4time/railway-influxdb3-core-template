# Operations

## Health and logs

Railway checks `GET /healthz` on port 8181 without authentication. This route
is mapped to the upstream `/health` endpoint using the protected internal
token. A 200 response means the adapter can reach the database.

Normal `/health` and API routes require:

```text
Authorization: Bearer <INFLUXDB3_EXTERNAL_BEARER_TOKEN>
```

Expected startup logs identify the adapter listener and InfluxDB startup.
Never paste full request headers, environment dumps, `/data/admin-token.json`,
or database contents into an issue.

## Restart and redeploy

A normal restart or redeploy reuses `/data`. Verify:

1. `/healthz` returns 200.
2. An authenticated `/health` returns 200.
3. The database catalog contains the expected databases.
4. A representative pre-restart query returns the expected data.

Keep one replica. A Railway volume is not multi-writer cluster storage.

## Cold backup

The supported package-level backup is an offline volume backup or cold copy:

1. Stop client writes.
2. Record the exact template version, upstream v3.10.0 pin, and volume identity.
3. Stop the service completely.
4. Create a Railway volume backup/snapshot or cold copy of all `/data`.
5. Retain the protected token file with its metadata; do not extract or log its
   value.
6. Restart the original service.
7. Verify health, catalog, and representative queries.

A copy taken while InfluxDB is running is outside this package's validated
contract.

## Restore

Restore only while the destination service is stopped:

1. Create a fresh volume and populate it from the complete cold backup.
2. Mount it at `/data` on a service using the exact image and wrapper version
   recorded with the backup.
3. Start exactly one replica.
4. Verify health, database catalog, protected-token continuity, and
   representative queries before directing writers to it.
5. Keep the old volume until the restored service is accepted.

The local proof restored a stopped pre-change snapshot into a genuinely fresh
volume and verified catalog, token, and query parity.

## Update

This package is deliberately pinned; it does not auto-update.

1. Read the upstream release notes and license/security changes.
2. Take and verify a cold backup.
3. Test the candidate image against a restored copy.
4. Re-run auth negatives, write/query, restart persistence, resource bounds,
   and backup/restore.
5. Update the immutable digest, source provenance, changelog, and checksums.
6. Obtain independent QA before deploying the update.

## Rollback

An upgrade may perform a one-way migration. Do not run an older image against a
volume already opened by a newer image.

Rollback means stopping the service, selecting the previously pinned image and
wrapper, and restoring the pre-upgrade cold backup into a fresh volume. Verify
the restored catalog and queries before switching traffic. There is no hidden
automatic rollback and no safe in-place downgrade promise.

## Troubleshooting

| Symptom | Check |
|---|---|
| `/healthz` fails | Startup logs, volume mount at `/data`, and `PORT=8181` |
| Service exits with code 70 | Missing/weak/equal external token or unsafe token filesystem entry |
| API returns 401 | Exact `Authorization: Bearer ...` header and current external token |
| Data missing after redeploy | Correct volume is still mounted at `/data` |
| Permission error | Volume is writable and no unsupported ownership changes were made |
| Restore fails | Service was stopped, full cold backup restored, exact prior package used |
