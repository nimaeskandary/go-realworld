#!/usr/bin/env bash
# Lints the files that have been staged for commit. Very useful as a pre-commit hook
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR="${SCRIPT_DIR}/.."

cd "${PROJECT_ROOT_DIR}"

GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [[ -z "$GO_FILES" ]]; then
  exit 0
fi

echo "running go fix..."
go fix ./...

echo "Running gofmt..."
gofmt -w $GO_FILES

echo "Running golangci-lint..."
# this needs to be run per package, not per file
DIRS=$(echo "$GO_FILES" | xargs -n1 dirname | sort -u)

for dir in $DIRS; do
    go tool -modfile=./tools/go.mod golangci-lint run "$dir" --fix
done
