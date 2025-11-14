#!/bin/bash
# Simple test script to verify examples compile and parser works

echo "=== Testing go-cyacd library ==="
echo

cd "$(dirname "$0")"

# Check if go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

echo "✓ Go found: $(go version)"
echo

# Test parsing the firmware file
echo "📄 Testing firmware parser..."
if [ -f "/Users/joseluismoffa/Downloads/hec-2-splt-dl-v0.6.0-7c817399.cyacd" ]; then
    go run -C examples/basic main.go
    echo
else
    echo "⚠️  Firmware file not found, using embedded test data"
    go run -C examples/basic main.go
    echo
fi

# Compile all examples
echo "🔨 Compiling examples..."
echo

for example in examples/*/; do
    name=$(basename "$example")
    echo "  Building $name..."
    if go build -o /tmp/test-"$name" "$example"main.go; then
        echo "  ✓ $name compiled successfully"
    else
        echo "  ❌ $name failed to compile"
        exit 1
    fi
done

echo
echo "✅ All examples compiled successfully!"
echo

# Run tests
echo "🧪 Running tests..."
go test -v ./...

echo
echo "✅ All tests completed!"
