#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR="${SCRIPT_DIR}/.."
TOOLS_DIR="${PROJECT_ROOT_DIR}/tools"

cd "${PROJECT_ROOT_DIR}"
go tool \
-modfile="${TOOLS_DIR}/go.mod" \
oapi-codegen \
--config pkg/api_gen/api.codegen.yaml \
pkg/api_gen/api.yaml
