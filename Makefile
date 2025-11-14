.PHONY: help test fmt vet lint examples clean

help:
	@echo "Available targets:"
	@echo "  test      - Run all tests"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  lint      - Run linters (requires golangci-lint)"
	@echo "  examples  - Build all examples"
	@echo "  clean     - Clean build artifacts"

test:
	@echo "Running tests..."
	go test -v -race ./...

fmt:
	@echo "Formatting code..."
	gofmt -s -w .

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running linters..."
	golangci-lint run

examples:
	@echo "Building examples..."
	@mkdir -p bin
	go build -o bin/basic ./examples/basic

clean:
	@echo "Cleaning..."
	rm -rf bin/
