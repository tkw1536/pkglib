.PHONY: test lint

test:
	go test ./...
lint:
	go tool golangci-lint run ./... --fix
	go tool govulncheck