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

echo "🔧 Compiling Go to WebAssembly with optimizations..."

# Build flags for optimization and security
BUILD_FLAGS=(
    -ldflags="-s -w"                    # Strip debugging info and symbol table
    -trimpath                           # Remove local path info for security
    -buildmode=default                  # Default build mode for WASM
)

# Additional optimization flags
export CGO_ENABLED=0                    # Disable CGO for better security and smaller size
export GOWASM=""                        # Use default WASM features

# Compile to WebAssembly with optimizations
echo "📝 Build flags: ${BUILD_FLAGS[*]}"
go build "${BUILD_FLAGS[@]}" -o main.wasm main.go

# Check if compilation was successful
if [ -f "main.wasm" ]; then
    WASM_SIZE=$(du -h main.wasm | cut -f1)
    WASM_SIZE_BYTES=$(stat -c%s main.wasm)
    echo "✅ Successfully built goxios-wasm!"
    echo "📊 WASM file size: $WASM_SIZE ($WASM_SIZE_BYTES bytes)"
    echo "📁 Output: main.wasm"
    
    # Optimize with wasm-opt if available
    if command -v wasm-opt &> /dev/null; then
        echo "🚀 Optimizing WASM with wasm-opt..."
        cp main.wasm main.wasm.backup
        wasm-opt -Oz --enable-bulk-memory --enable-sign-ext --enable-mutable-globals \
                 --enable-nontrapping-float-to-int main.wasm.backup -o main.wasm
        
        NEW_SIZE=$(du -h main.wasm | cut -f1)
        NEW_SIZE_BYTES=$(stat -c%s main.wasm)
        REDUCTION=$((WASM_SIZE_BYTES - NEW_SIZE_BYTES))
        REDUCTION_PERCENT=$((REDUCTION * 100 / WASM_SIZE_BYTES))
        
        echo "✨ Optimized size: $NEW_SIZE ($NEW_SIZE_BYTES bytes)"
        echo "📉 Size reduction: $REDUCTION bytes ($REDUCTION_PERCENT%)"
        rm main.wasm.backup
    else
        echo "⚠️  wasm-opt not found. Install binaryen for better optimization:"
        echo "   sudo apt install binaryen  # Ubuntu/Debian"
        echo "   brew install binaryen      # macOS"
    fi
    
    # Create compressed version
    echo "🗜️  Creating compressed version..."
    gzip -9 -k main.wasm
    if [ -f "main.wasm.gz" ]; then
        GZIP_SIZE=$(du -h main.wasm.gz | cut -f1)
        GZIP_SIZE_BYTES=$(stat -c%s main.wasm.gz)
        GZIP_REDUCTION=$((WASM_SIZE_BYTES - GZIP_SIZE_BYTES))
        GZIP_REDUCTION_PERCENT=$((GZIP_REDUCTION * 100 / WASM_SIZE_BYTES))
        echo "📦 Gzipped size: $GZIP_SIZE ($GZIP_SIZE_BYTES bytes)"
        echo "📉 Gzip reduction: $GZIP_REDUCTION bytes ($GZIP_REDUCTION_PERCENT%)"
    fi
    
    # Generate integrity hash
    echo "🔐 Generating integrity hash..."
    HASH=$(sha256sum main.wasm | cut -d' ' -f1)
    echo "sha256-$(echo -n $HASH | base64)" > main.wasm.integrity
    echo "🔑 Integrity hash: sha256-$(echo -n $HASH | base64)"
    
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
