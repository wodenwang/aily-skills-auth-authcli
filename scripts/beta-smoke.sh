#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

cd "$ROOT_DIR"

echo "==> go test ./..."
go test ./...

echo "==> auth-cli check --help"
HELP_OUTPUT="$(go run ./cmd/auth-cli check --help)"
printf '%s\n' "$HELP_OUTPUT" >"$TMP_DIR/help.txt"
grep -q "auth-cli check --skill <skill_id> --user-id <user_id>" "$TMP_DIR/help.txt"
grep -q "Input Priority:" "$TMP_DIR/help.txt"
grep -q "Cache Semantics:" "$TMP_DIR/help.txt"
grep -q "Install And Upgrade:" "$TMP_DIR/help.txt"

echo "==> build temp binary"
go build -o "$TMP_DIR/auth-cli" ./cmd/auth-cli

echo "==> offline invalid-input check"
set +e
"$TMP_DIR/auth-cli" check >"$TMP_DIR/stdout.txt" 2>"$TMP_DIR/stderr.txt"
RC=$?
set -e
if [ "$RC" -ne 20 ]; then
  echo "expected exit code 20, got $RC" >&2
  exit 1
fi
grep -q "AUTHCLI_INVALID_INPUT: missing required flag: --skill" "$TMP_DIR/stderr.txt"

echo "beta smoke passed"
