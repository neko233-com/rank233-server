#!/usr/bin/env bash
set -euo pipefail
echo "Running tests..."
go test ./... -count=1 -v
echo ""
echo "ALL TESTS PASSED"
