# Marketplace overview

## Name

InfluxDB 3 Core

## Short description

Single-node InfluxDB 3 Core v3.10.0 with authenticated public API access and a
persistent Railway volume.

## Services

- One application service
- One volume at `/data`

## First run

The template sets `PORT=8181` and generates a 64-character external bearer.
The container safely creates a separate internal admin token on the volume,
drops permanently to UID/GID 1500, then starts the database and adapter.

## Persistence and support

Database state and the internal token persist at `/data`. The package supports
one replica and does not claim HA, automatic updates, safe in-place downgrade,
or production sizing. Offline backup/restore and restore-based rollback require
operator action.

This independent community template is not endorsed by InfluxData.
