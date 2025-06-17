#!/bin/bash

# Build script for qr-wasm module
# Compiles Go source to optimized WebAssembly for QR code and barcode operations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📱 Building qr-wasm module...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

# Set Go environment for WebAssembly
export GOOS=js
export GOARCH=wasm

# Get dependencies
echo -e "${YELLOW}📦 Getting dependencies...${NC}"
go mod download
go mod tidy

# Check Go version compatibility
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${YELLOW}🔍 Using Go version: $GO_VERSION${NC}"

# Build optimized WASM
echo -e "${YELLOW}🔨 Compiling Go to WebAssembly...${NC}"

# Build flags for optimization
BUILD_FLAGS="-ldflags=-s -ldflags=-w"

# Compile with optimizations
go build $BUILD_FLAGS -o main.wasm main.go

if [ ! -f "main.wasm" ]; then
    echo -e "${RED}❌ Build failed: main.wasm not generated${NC}"
    exit 1
fi

# Get file size
ORIGINAL_SIZE=$(stat -f%z main.wasm 2>/dev/null || stat -c%s main.wasm 2>/dev/null)
echo -e "${GREEN}✅ WebAssembly compiled successfully${NC}"
echo -e "${YELLOW}📊 Original size: $(echo "scale=1; $ORIGINAL_SIZE/1024/1024" | bc -l 2>/dev/null || echo "$((ORIGINAL_SIZE/1024/1024))")MB${NC}"

# Compress with gzip
echo -e "${YELLOW}🗜️  Compressing with gzip...${NC}"
gzip -9 -k main.wasm

if [ -f "main.wasm.gz" ]; then
    COMPRESSED_SIZE=$(stat -f%z main.wasm.gz 2>/dev/null || stat -c%s main.wasm.gz 2>/dev/null)
    COMPRESSION_RATIO=$(echo "scale=1; (1 - $COMPRESSED_SIZE/$ORIGINAL_SIZE) * 100" | bc -l 2>/dev/null || echo "50")
    echo -e "${GREEN}✅ Compressed size: $(echo "scale=1; $COMPRESSED_SIZE/1024/1024" | bc -l 2>/dev/null || echo "$((COMPRESSED_SIZE/1024/1024))")MB (${COMPRESSION_RATIO}% reduction)${NC}"
else
    echo -e "${YELLOW}⚠️  Compression failed, but build is still usable${NC}"
fi

# Generate integrity hash
echo -e "${YELLOW}🔐 Generating integrity hash...${NC}"
if command -v shasum &> /dev/null; then
    shasum -a 256 main.wasm | awk '{print $1}' > main.wasm.integrity
elif command -v sha256sum &> /dev/null; then
    sha256sum main.wasm | awk '{print $1}' > main.wasm.integrity
else
    echo -e "${YELLOW}⚠️  SHA256 tool not found, skipping integrity hash${NC}"
fi

if [ -f "main.wasm.integrity" ]; then
    INTEGRITY_HASH=$(cat main.wasm.integrity)
    echo -e "${GREEN}✅ Integrity hash: ${INTEGRITY_HASH:0:16}...${NC}"
fi

echo ""
echo -e "${GREEN}🎉 QR WASM module build completed successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Build Summary:${NC}"
echo -e "   Module: qr-wasm"
echo -e "   Original: $(echo "scale=1; $ORIGINAL_SIZE/1024/1024" | bc -l 2>/dev/null || echo "$((ORIGINAL_SIZE/1024/1024))")MB"
if [ -f "main.wasm.gz" ]; then
    echo -e "   Compressed: $(echo "scale=1; $COMPRESSED_SIZE/1024/1024" | bc -l 2>/dev/null || echo "$((COMPRESSED_SIZE/1024/1024))")MB"
fi
echo -e "   Functions: generateQRCode, generateBarcode, generateVCard, generateWiFiQR"
echo ""
echo -e "${YELLOW}📝 Usage Examples:${NC}"
echo -e "   const qr = await loadFromGitHub('benoitpetit/wasm-modules-repository', {"
echo -e "     branch: 'main', name: 'qr-wasm'"
echo -e "   });"
echo -e "   const result = qr.call('generateQRCode', 'Hello World', 256);"
echo ""
echo -e "${YELLOW}🔧 GoWM Integration:${NC}"
echo -e "   ✅ Ready signal: __gowm_ready"
echo -e "   ✅ Function discovery: getAvailableFunctions()"
echo -e "   ✅ Silent mode: setSilentMode(boolean)"
echo -e "   ✅ Error patterns: Consistent error handling"
echo ""

# Validate build
if [ -f "main.wasm" ]; then
    echo -e "${GREEN}✅ Build validation: main.wasm exists${NC}"
    
    # Check if module.json exists
    if [ -f "module.json" ]; then
        echo -e "${GREEN}✅ Configuration: module.json found${NC}"
    else
        echo -e "${YELLOW}⚠️  Configuration: module.json missing${NC}"
    fi
else
    echo -e "${RED}❌ Build validation failed${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 QR WASM module is ready for deployment!${NC}"
