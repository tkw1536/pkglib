#!/bin/bash
set -e

echo "=> golangci-lint"
go tool golangci-lint run ./... --fix

echo "=> modernize"
go tool modernize -test ./...

echo "=> govulncheck"
go tool govulncheck
