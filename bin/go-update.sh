#!/usr/bin/env bash
# go-update.sh
# updates go deps to latest minor and patch versions.
# to update a major version of a dep, you must update it manually, e.g. go get example.com/somepackage@latest
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR="${SCRIPT_DIR}/.."

cd "${PROJECT_ROOT_DIR}"
echo "root update..."
go get -u ./...

cd "${PROJECT_ROOT_DIR}/tools"
echo "tools update..."
go get -u tool

"${PROJECT_ROOT_DIR}/bin/go-mod-tidy.sh"
