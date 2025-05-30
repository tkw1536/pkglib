.PHONY: test lint

test:
	GOEXPERIMENT=synctest go test ./...
lint:
	GOEXPERIMENT=synctest go tool golangci-lint run ./... --fix
	GOEXPERIMENT=synctest go tool modernize -test ./...
	GOEXPERIMENT=synctest go tool govulncheck