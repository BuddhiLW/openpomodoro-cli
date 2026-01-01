#!/bin/bash
# contribution.sh - Helper script for switching between fork and upstream module paths
#
# Usage:
#   ./contribution.sh upstream  - Switch to upstream module path (for PRs)
#   ./contribution.sh fork      - Switch to fork module path (for personal use)
#   ./contribution.sh status    - Show current module path

set -e

UPSTREAM_PATH="github.com/open-pomodoro/openpomodoro-cli"
FORK_PATH="github.com/BuddhiLW/openpomodoro-cli"

GO_FILES=$(find . -name "*.go" -type f | grep -v vendor)

current_path() {
    grep "^module" go.mod | awk '{print $2}'
}

switch_to() {
    local from=$1
    local to=$2

    echo "Switching from $from to $to..."

    # Update go.mod
    sed -i "s|module $from|module $to|g" go.mod

    # Update all Go imports
    for file in $GO_FILES; do
        sed -i "s|$from|$to|g" "$file"
    done

    # Update README.md
    sed -i "s|$from|$to|g" README.md

    # Tidy modules
    go mod tidy

    echo "Done! Module path is now: $to"
}

case "$1" in
    upstream)
        current=$(current_path)
        if [ "$current" = "$UPSTREAM_PATH" ]; then
            echo "Already using upstream path: $UPSTREAM_PATH"
            exit 0
        fi
        switch_to "$FORK_PATH" "$UPSTREAM_PATH"
        ;;
    fork)
        current=$(current_path)
        if [ "$current" = "$FORK_PATH" ]; then
            echo "Already using fork path: $FORK_PATH"
            exit 0
        fi
        switch_to "$UPSTREAM_PATH" "$FORK_PATH"
        ;;
    status)
        echo "Current module path: $(current_path)"
        ;;
    *)
        echo "Usage: $0 {upstream|fork|status}"
        echo ""
        echo "  upstream  - Switch to upstream module path (for PRs to open-pomodoro)"
        echo "  fork      - Switch to fork module path (for personal use/releases)"
        echo "  status    - Show current module path"
        exit 1
        ;;
esac
