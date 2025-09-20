default:
  just --list

example:
  go run ./cmd httpbin get --key value

test:
  go test -cover ./...

coverage:
  go test -coverprofile=coverage.txt ./...

build:
  go build -o dist/ashttp ./cmd

lint: install-linter
  golangci-lint run

# https://golangci-lint.run/docs/welcome/install/#local-installation
install-linter:
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
