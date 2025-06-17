#!/bin/bash

# Build script for the goxios-wasm module
# This script compiles Go code to WebAssembly

set -e  # Exit on error

echo "🔨 Building goxios-wasm module..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Check Go version (WASM requires Go 1.11+)
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
echo "📦 Detected Go version: $GO_VERSION"

# Clean old build files
echo "🧹 Cleaning old build files..."
rm -f main.wasm

# Set environment variables for WebAssembly compilation
export GOOS=js
export GOARCH=wasm

echo "🔧 Compiling Go to WebAssembly..."

# Compile to WebAssembly
go build -o main.wasm main.go

# Check if compilation was successful
if [ -f "main.wasm" ]; then
    WASM_SIZE=$(du -h main.wasm | cut -f1)
    echo "✅ Successfully built goxios-wasm!"
    echo "📊 WASM file size: $WASM_SIZE"
    echo "📁 Output: main.wasm"
    
    # Optional: Show file info
    echo ""
    echo "📋 Build Summary:"
    echo "   Module: goxios-wasm"
    echo "   Type: HTTP Client Library"
    echo "   Size: $WASM_SIZE"
    echo "   Target: js/wasm"
    echo "   Features: GET, POST, PUT, DELETE, PATCH, Instances, Error Handling"
    echo ""
    echo "🚀 Ready to use with gowm npm!"
    echo ""
    echo "Usage example:"
    echo "   // Load the WASM module first"
    echo "   // Then use:"
    echo "   const response = await goxios.get('https://api.example.com/data');"
    echo "   console.log(response.data);"
else
    echo "❌ Build failed!"
    exit 1
fi

# Optional: Validate the WASM file
if command -v file &> /dev/null; then
    echo ""
    echo "🔍 WASM file validation:"
    file main.wasm
fi

echo ""
echo "🎉 Build completed successfully!"
