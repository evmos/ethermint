#!/bin/sh

set -e 

if [ "$1" = 'evmosd' ]; then
    ./init.sh

    exec "$@" "--"
fi

exec "$@"