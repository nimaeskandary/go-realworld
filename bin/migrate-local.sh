#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR="${SCRIPT_DIR}/.."

cd "${PROJECT_ROOT_DIR}"
go run cmd/migrations/main.go -config-path config/local.yaml -target-database realworld_app -action apply-all
