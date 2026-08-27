#!/bin/bash
set -euo pipefail

if [[ "$(id -u):$(id -g)" != "1500:1500" ]]; then
    echo "fatal: supervisor privilege drop failed" >&2
    exit 70
fi
external_bearer_token="${INFLUXDB3_EXTERNAL_BEARER_TOKEN:-}"
if (( ${#external_bearer_token} < 32 )); then
    echo "fatal: external bearer token must contain at least 32 characters" >&2
    exit 70
fi
unset external_bearer_token

influx_pid=""
adapter_pid=""
stopping=0

stop_children() {
    if (( stopping == 1 )); then
        return
    fi
    stopping=1
    [[ -z "${adapter_pid}" ]] || kill -TERM "${adapter_pid}" 2>/dev/null || true
    [[ -z "${influx_pid}" ]] || kill -TERM "${influx_pid}" 2>/dev/null || true
}

trap stop_children TERM INT HUP

env -u INFLUXDB3_EXTERNAL_BEARER_TOKEN influxdb3 serve \
    --node-id railway-template \
    --object-store file \
    --data-dir /data \
    --http-bind 127.0.0.1:8182 \
    --admin-token-file /data/admin-token.json &
influx_pid=$!

/usr/local/bin/influx-adapter \
    --listen 0.0.0.0:8181 \
    --upstream http://127.0.0.1:8182 \
    --token-file /data/admin-token.json &
adapter_pid=$!

set +e
wait -n "${influx_pid}" "${adapter_pid}"
first_status=$?
set -e
stop_children

set +e
wait "${adapter_pid}"
adapter_status=$?
wait "${influx_pid}"
influx_status=$?
set -e

if (( stopping == 1 )) && (( first_status == 0 || first_status == 143 )); then
    exit 0
fi
if (( first_status != 0 )); then
    exit "${first_status}"
fi
if (( adapter_status != 0 && adapter_status != 143 )); then
    exit "${adapter_status}"
fi
if (( influx_status != 0 && influx_status != 143 )); then
    exit "${influx_status}"
fi
