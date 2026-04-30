.PHONY: build install clean test

# Build the binary
build:
	go build -o ob ./cmd/ob

# Install to $GOPATH/bin
install:
	go install ./cmd/ob

# Clean build artifacts
clean:
	rm -f ob
	go clean

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download

# Tidy dependencies
tidy:
	go mod tidy

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o ob-linux-amd64 ./cmd/ob
	GOOS=linux GOARCH=arm64 go build -o ob-linux-arm64 ./cmd/ob
	GOOS=darwin GOARCH=amd64 go build -o ob-darwin-amd64 ./cmd/ob
	GOOS=darwin GOARCH=arm64 go build -o ob-darwin-arm64 ./cmd/ob
	GOOS=windows GOARCH=amd64 go build -o ob-windows-amd64.exe ./cmd/ob

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install to \$$GOPATH/bin"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  deps       - Download dependencies"
	@echo "  tidy       - Tidy dependencies"
	@echo "  build-all  - Build for multiple platforms"
