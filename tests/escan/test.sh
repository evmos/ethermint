#!/bin/bash

set -eu

ENV_FILE='./tests/escan/env.sh'
if [[ ! -f "$ENV_FILE" ]]; then
  echo 'Wrong working directory, must be run from root dir of this project'
  exit 1
fi

source "$ENV_FILE"

TS_EXEC="ts-node ./tests/escan/test.ts"

$TS_EXEC 01

echo 'Done'