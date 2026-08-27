# Security policy

## Reporting

For a vulnerability in this wrapper, open a private security advisory in the
template repository. Do not file a public issue containing exploit details,
tokens, variable values, request headers, database contents, or backups.

For an InfluxDB vulnerability, follow InfluxData's upstream security reporting
process. For Railway platform issues, contact Railway support.

## Supported package

Security fixes target the current template package and pinned upstream image.
Upstream pins are changed only after fresh build, auth, persistence, backup,
restore, and rollback-path validation.

## Secret exposure

If the external bearer token is exposed:

1. Replace `INFLUXDB3_EXTERNAL_BEARER_TOKEN` with a new random value of at least
   32 characters.
2. Redeploy the service.
3. Remove the leaked value from issues, logs, shell history, and integrations.

If `/data/admin-token.json` or a backup is exposed, treat the database as
compromised. Restrict access, create a clean deployment or verified restore,
and rotate all external credentials. Never set the external token equal to the
internal token.
