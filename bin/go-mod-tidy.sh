#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR="${SCRIPT_DIR}/.."

cd "${PROJECT_ROOT_DIR}"
echo "root mod tidy..."
go mod tidy

echo "tools mod tidy..."
cd "${PROJECT_ROOT_DIR}/tools"
go mod tidy
