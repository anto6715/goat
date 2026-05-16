#!/usr/bin/env bash

# BASH SETTINGS ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
set -o errexit
set -o nounset
set -o pipefail

# VARIABLES ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
declare -r __prj_dir__="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
declare -r __bin_dir__="$__prj_dir__/bin"
declare -r __cmd_dir__="$__prj_dir__/cmd"

## List of expected commands to build
declare -a __cmds__=(
    gfind
    md5
)

# EXECUTE ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## Create bin directory if it doesn't exist
[ ! -d "$__bin_dir__" ] &&
    mkdir -p "$__bin_dir__"

echo "#####################################################################"
echo "# Building $__prj_dir__"
echo "#####################################################################"

cd "$__prj_dir__"
for cmd in "${__cmds__[@]}"; do
    cmd_dir="$__cmd_dir__/$cmd"
    if [ ! -d "$cmd_dir" ]; then
        echo "Not exists: $cmd"
        exit 1
    fi
    echo "Building: $cmd"
    go build -o "$__bin_dir__/$cmd" $cmd_dir
done
