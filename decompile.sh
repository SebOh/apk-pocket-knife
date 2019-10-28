#!/usr/bin/env bash

# Stops the execution of a script if a command or pipeline has an error
set -e

# Default Arguments
ARG1=${1:bash}

if [ "$ARG1" = 'decompile' ]; then
    exec .././script/entrypoint "$@"
fi

exec "$@"

