#!/bin/sh

set -eu

REPO_OWNER="wodenwang"
REPO_NAME="aily-skills-auth-authcli"
DEFAULT_VERSION="v0.2.0"
DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
DEFAULT_BIN_NAME="auth-cli"

usage() {
  cat <<'EOF'
Usage:
  install-authcli.sh [--version <tag>] [--install-dir <dir>] [--bin-name <name>]

Examples:
  sh install-authcli.sh
  sh install-authcli.sh --version v0.2.0
  sh install-authcli.sh --install-dir /usr/local/bin
EOF
}

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'install-authcli.sh: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

detect_platform() {
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Darwin) os_slug="darwin" ;;
    Linux) os_slug="linux" ;;
    *) fail "unsupported operating system: $os" ;;
  esac

  case "$arch" in
    arm64|aarch64) arch_slug="arm64" ;;
    x86_64|amd64) arch_slug="amd64" ;;
    *) fail "unsupported architecture: $arch" ;;
  esac

  case "${os_slug}-${arch_slug}" in
    darwin-arm64|linux-amd64)
      ASSET_SUFFIX="${os_slug}-${arch_slug}"
      ;;
    *)
      fail "unsupported platform: ${os_slug}-${arch_slug}; supported: darwin-arm64, linux-amd64"
      ;;
  esac
}

download() {
  url="$1"
  output="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
    return
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$output" "$url"
    return
  fi

  fail "missing downloader: curl or wget is required"
}

install_file() {
  src="$1"
  dst="$2"

  if command -v install >/dev/null 2>&1; then
    install "$src" "$dst"
  else
    cp "$src" "$dst"
    chmod 0755 "$dst"
  fi
}

verify_binary() {
  bin_path="$1"
  stderr_file="$2"

  set +e
  "$bin_path" check >/dev/null 2>"$stderr_file"
  rc=$?
  set -e

  if [ "$rc" -ne 20 ]; then
    fail "verification failed: expected exit code 20, got $rc"
  fi

  if ! grep -q "AUTHCLI_INVALID_INPUT: missing required flag: --skill" "$stderr_file"; then
    fail "verification failed: stderr did not contain the expected frozen prefix"
  fi
}

VERSION="$DEFAULT_VERSION"
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
BIN_NAME="$DEFAULT_BIN_NAME"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || fail "missing value for --version"
      VERSION="$2"
      shift 2
      ;;
    --install-dir)
      [ "$#" -ge 2 ] || fail "missing value for --install-dir"
      INSTALL_DIR="$2"
      shift 2
      ;;
    --bin-name)
      [ "$#" -ge 2 ] || fail "missing value for --bin-name"
      BIN_NAME="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

need_cmd tar
need_cmd mktemp
need_cmd grep

detect_platform

ASSET_NAME="auth-cli-${ASSET_SUFFIX}.tar.gz"
EXTRACTED_NAME="auth-cli-${ASSET_SUFFIX}"
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${ASSET_NAME}"

TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${ASSET_NAME}"
EXTRACT_DIR="${TMP_DIR}/extract"
STDERR_FILE="${TMP_DIR}/verify.stderr"
INSTALL_PATH="${INSTALL_DIR%/}/${BIN_NAME}"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

mkdir -p "$EXTRACT_DIR"
mkdir -p "$INSTALL_DIR"

log "Downloading ${DOWNLOAD_URL}"
download "$DOWNLOAD_URL" "$ARCHIVE_PATH"

log "Extracting ${ASSET_NAME}"
tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"

[ -f "${EXTRACT_DIR}/${EXTRACTED_NAME}" ] || fail "archive did not contain ${EXTRACTED_NAME}"

log "Installing ${INSTALL_PATH}"
install_file "${EXTRACT_DIR}/${EXTRACTED_NAME}" "$INSTALL_PATH"

log "Verifying ${INSTALL_PATH}"
verify_binary "$INSTALL_PATH" "$STDERR_FILE"

cat <<EOF
Installed:
- binary: ${INSTALL_PATH}
- version: ${VERSION}
- asset: ${ASSET_NAME}

Next steps:
1. Ensure ${INSTALL_DIR} is in PATH.
2. Set AUTHCLI_IAM_BASE_URL to your iam-service base URL.
3. Optional: set AUTHCLI_CACHE_PATH and AUTHCLI_CONFIG_FILE.
4. Run: ${INSTALL_PATH} check --skill <skill_id> --user-id <user_id> --format json
EOF
