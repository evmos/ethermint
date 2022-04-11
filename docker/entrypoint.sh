#!/bin/sh

set -e 

if [ "$1" = 'ethermintd' ]; then
    ./init.sh

    exec "$@" "--"
fi

exec "$@"
