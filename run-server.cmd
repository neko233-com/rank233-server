@echo off
setlocal
echo Running rank233-server on :6320 ...
go run ./cmd/rank233-server -addr 0.0.0.0:6320 %*
