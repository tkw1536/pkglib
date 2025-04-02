#!/bin/bash
set -e

echo "=> golangci-lint"
go tool golangci-lint run ./... --fix

echo "=> govulncheck"
go tool govulncheck
