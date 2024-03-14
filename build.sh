#!/usr/bin/env bash

shopt -s globstar
shopt -s extglob

function getRealPath()
{
    local -i traversals=0
    currentDir="$1"
    basename=''
    while :; do
        [[ "$currentDir" == '.' ]] && { echo "$1"; return 1; }
        [[ $traversals -eq 0 ]] && pwd=$(cd "$currentDir" 2>&1 && pwd) && { echo "$pwd/$basename"; return 0; }
        currentBasename="$(basename "$currentDir")"
        currentDir="$(dirname "$currentDir")"
        [[ "$currentBasename" == '..' ]] && (( ++traversals )) || { [[ traversals -gt 0 ]] && (( traversals-- )) || basename="$currentBasename/$basename"; }
    done
}

# Generate go code
echo "Generate go code"
path="./**/*.capnp"
go_stdlib=$(getRealPath "../go-capnp/std")
capnp compile -I $go_stdlib --verbose -ogo $path

# Build go code
echo "Build go code"
rm -r bin
mkdir -p bin
go build -gcflags=all="-N -l" -o bin/server ./cmds/server/main.go
go build -gcflags=all="-N -l" -o bin/client ./cmds/client/main.go