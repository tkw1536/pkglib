#!/bin/bash
set -e

echo "=> golangci-lint"
go tool golangci-lint run ./...

echo "=> govulncheck"
go tool govulncheck
