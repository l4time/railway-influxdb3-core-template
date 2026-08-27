#!/bin/bash
set -euo pipefail

readonly DATA_DIR="/data"
readonly TOKEN_FILE="${DATA_DIR}/admin-token.json"

if [[ "$(id -u)" != "0" ]]; then
    echo "fatal: init must start as root" >&2
    exit 70
fi

/usr/local/bin/secure-init

exec setpriv --reuid=1500 --regid=1500 --init-groups \
    /usr/local/bin/supervisor.sh
