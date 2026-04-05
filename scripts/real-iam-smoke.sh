#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IAM_BASE_URL="${AUTHCLI_REAL_IAM_BASE_URL:-http://127.0.0.1:8000}"

echo "Running authcli real IAM smoke against ${IAM_BASE_URL}"

AUTHCLI_REAL_IAM_BASE_URL="${IAM_BASE_URL}" \
go test ./... -run TestRealIAM
