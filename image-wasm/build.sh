#!/bin/bash

# Build script for the image-wasm module
# This script compiles Go code to WebAssembly

set -e  # Exit on error

echo "🔨 Building image-wasm module..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Check Go version (WASM requires Go 1.11+)
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
echo "📦 Detected Go version: $GO_VERSION"

# Clean old build files
echo "🧹 Cleaning old files..."
rm -f main.wasm

# Set environment variables for WebAssembly
export GOOS=js
export GOARCH=wasm

# Compile the module
echo "⚙️  Compiling..."
go build -o main.wasm main.go

# Check if compilation succeeded
if [ -f "main.wasm" ]; then
    WASM_SIZE=$(du -h main.wasm | cut -f1)
    echo "✅ Compilation successful!"
    echo "📏 WASM file size: $WASM_SIZE"
    echo "📁 Generated file: main.wasm"
    
    # Check if wasm_exec.js exists
    if [ ! -f "wasm_exec.js" ]; then
        echo "⚠️  Warning: wasm_exec.js not found"
        echo "💡 To get it, run:"
        echo "   cp \"\$(go env GOROOT)/misc/wasm/wasm_exec.js\" ."
    fi
    
    echo ""
    echo "🚀 Module is ready to use!"
    echo "   To test, include wasm_exec.js and main.wasm in your web project"
    echo ""
    echo "🖼️  Available features:"
    echo "   • JPEG/PNG compression"
    echo "   • Image resizing" 
    echo "   • WebP conversion"
    echo "   • Image information"
else
    echo "❌ Error: Compilation failed"
    exit 1
fi
