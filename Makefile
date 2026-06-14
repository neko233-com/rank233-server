VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)

MODULE  = github.com/yourname/rank233-server
LDFLAGS = -s -w \
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: test lint build tidy run-server

test:
	go test ./... -count=1

test-race:
	go test ./... -count=1 -race

lint:
	golangci-lint run --timeout=5m

build:
	go build -ldflags="$(LDFLAGS)" -o bin/rank233-server.exe ./cmd/rank233-server

run-server: build
	bin/rank233-server.exe

tidy:
	go mod tidy
