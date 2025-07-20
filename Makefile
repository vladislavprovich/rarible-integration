.PHONY: start lint vulncheck fmt test build

start:
	@echo "Starting local service..."
	-go run cmd/raribleintegration/main.go --config.go=./.env || true

lint:
	@echo "Running linters..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout=5m

vulncheck:
	@echo "Checking for security vulnerabilities..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

fmt:
	@echo "Formatting code..."
	gofmt -w .

test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building the project..."
	go build -o userinfo cmd/main.go
