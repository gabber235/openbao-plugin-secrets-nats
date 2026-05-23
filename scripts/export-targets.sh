#!/bin/bash
set -euo pipefail

DIR="${1:-.}"
PLUGIN_NAME="${2:-"openbao-plugin-secrets-nats"}"

re="$PLUGIN_NAME-([^-]+)-(.+)$"

csv=$(
  for f in "$DIR"/*; do
    name=$(basename $f)
    if [[ "$name" =~ $re ]]; then 
        OS=${BASH_REMATCH[1]};
        ARCH=${BASH_REMATCH[2]};
        
        # OpenBao OCI puller expects linux/arm64/v8 variant
        if [[ "$OS" == "linux" && "$ARCH" == "arm64" ]]; then
            echo "$OS/$ARCH/v8"
        else
            echo "$OS/$ARCH"
        fi
    fi
  done | sort -u | paste -sd,
)

echo $csv
