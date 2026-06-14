#!/usr/bin/env bash
set -euo pipefail

REPO="yourname/rank233-server"
BINARY="rank233-server"
INSTALL_DIR="${RANK233_INSTALL:-/usr/local/bin}"

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$os" in
    linux)  os="linux" ;;
    darwin) os="darwin" ;;
    mingw*|msys*|cygwin*) os="windows" ;;
    *) echo "unsupported os: $os"; exit 1 ;;
  esac
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) echo "unsupported arch: $arch"; exit 1 ;;
  esac
  echo "${os}-${arch}"
}

resolve_version() {
  if [ -n "${1:-}" ]; then
    echo "$1"
    return
  fi
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
  local platform version url ext tmp
  platform="$(detect_platform)"
  version="$(resolve_version "${1:-}")"
  if [ -z "$version" ]; then
    echo "ERROR: could not determine version. Pass explicitly: install-server.sh v0.1.0"
    exit 1
  fi

  ext=""
  case "$platform" in
    windows-*) ext=".exe" ;;
  esac

  url="https://github.com/${REPO}/releases/download/${version}/${BINARY}-${platform}${ext}"
  echo "Installing ${BINARY} ${version} (${platform})..."
  echo "  Download: ${url}"

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  curl -fsSL -o "${tmp}/${BINARY}${ext}" "$url"
  chmod +x "${tmp}/${BINARY}${ext}"

  if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
  fi

  mv "${tmp}/${BINARY}${ext}" "${INSTALL_DIR}/${BINARY}${ext}"
  echo ""
  echo "Installed: ${INSTALL_DIR}/${BINARY}${ext}"
  echo ""
  echo "Quick start:"
  echo "  ${BINARY}                    # start server on :6320"
  echo "  ${BINARY} -addr :8080        # custom port"
  echo "  ${BINARY} -version           # print version"
  echo ""
  echo "Docker:"
  echo "  docker run -p 6320:6320 ghcr.io/${REPO}:${version}"
}

main "$@"
