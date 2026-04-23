#!/usr/bin/env bash
# Local pre-commit verification for this Go module.
# Mirrors the GitHub Actions CI: build, vet, gofmt, race-tested tests.
#
# Usage:
#   ./.scripts/check.sh
#
# Exits non-zero on the first failing check so pre-commit hooks and
# developer workflows fail fast.

set -euo pipefail

# Always run from the module root, regardless of CWD when the script is invoked.
cd "$(dirname "$0")/.."

step() {
    printf '\n==> %s\n' "$1"
}

step "go build ./..."
go build ./...

step "go vet ./..."
go vet ./...

step "gofmt -l . (expecting empty output)"
unformatted=$(gofmt -l .)
if [[ -n "$unformatted" ]]; then
    printf 'gofmt found unformatted files:\n%s\n' "$unformatted"
    exit 1
fi

step "go test -race ./..."
go test -race ./...

printf '\nAll local checks passed.\n'
