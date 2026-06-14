#!/usr/bin/env bash
set -euo pipefail

echo "Building..."
bash build-all.sh

echo ""
echo "Running tests..."
go test ./... -count=1 -race

echo ""
echo "Smoke: starting server..."
./bin/rank233-server -addr 127.0.0.1:16320 &
PID=$!
trap "kill $PID 2>/dev/null || true" EXIT
sleep 2

echo "  healthz..."
curl -sf http://127.0.0.1:16320/healthz | grep -q ok

echo "  version..."
curl -sf http://127.0.0.1:16320/version

echo "  create ranklist..."
curl -sf -X POST http://127.0.0.1:16320/api/ranklist/create \
  -H 'Content-Type: application/json' \
  -d '{"name":"smoke","capacity":100}' | grep -q created

echo "  put score..."
curl -sf -X POST http://127.0.0.1:16320/api/ranklist/put \
  -H 'Content-Type: application/json' \
  -d '{"name":"smoke","member":1,"primary":100,"secondary":50,"arrival":1}'

echo "  get rank..."
curl -sf "http://127.0.0.1:16320/api/ranklist/rank?name=smoke&member=1" | grep -q '"rank":1'

echo "  top n..."
curl -sf "http://127.0.0.1:16320/api/ranklist/top?name=smoke&limit=10" | grep -q '"member":1'

echo ""
echo "CI PASSED"
