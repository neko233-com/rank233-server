#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)"

LDFLAGS="-s -w \
  -X github.com/yourname/rank233-server/internal/version.Version=${VERSION} \
  -X github.com/yourname/rank233-server/internal/version.Commit=${COMMIT} \
  -X github.com/yourname/rank233-server/internal/version.Date=${DATE}"

mkdir -p bin
echo "Building rank233-server..."
go build -ldflags="${LDFLAGS}" -o bin/rank233-server ./cmd/rank233-server
echo "Done: bin/rank233-server"
