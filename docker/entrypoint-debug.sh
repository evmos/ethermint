#!/bin/sh

set -e

if [ "$1" = 'dlv' ]; then
    ./init.sh

    exec "$@" "--"
fi

exec "$@"
