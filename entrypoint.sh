#!/bin/sh

set -eu

if [ "$(id -u)" -eq 0 ]; then
    chown -R 65534:65534 /data
    chmod 0750 /data

    exec setpriv \
        --reuid=65534 \
        --regid=65534 \
        --clear-groups \
        --no-new-privs \
        "$@"
fi

exec "$@"
