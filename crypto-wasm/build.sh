#!/bin/bash

# Build script for crypto-wasm module
# Compiles Go source to optimized WebAssembly with security considerations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔐 Building crypto-wasm module...${NC}"

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

# Build flags for optimization and security
BUILD_FLAGS=(
    -ldflags="-s -w"           # Strip debug info and symbol table
    -trimpath                  # Remove file system paths from binary
    -buildmode=default         # Standard build mode
)

# Security-focused build settings
export CGO_ENABLED=0           # Disable CGO for security
export GOEXPERIMENT=""         # No experimental features

echo -e "${YELLOW}⚙️  Compiling Go to WebAssembly...${NC}"

# Build the WASM module
if go build "${BUILD_FLAGS[@]}" -o main.wasm main.go; then
    echo -e "${GREEN}✅ Build successful${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

# Check if wasm file was created
if [ ! -f "main.wasm" ]; then
    echo -e "${RED}❌ WASM file not created${NC}"
    exit 1
fi

# Get original size
ORIGINAL_SIZE=$(stat -c%s main.wasm 2>/dev/null || stat -f%z main.wasm 2>/dev/null || echo "unknown")

echo -e "${BLUE}📊 Original size: ${ORIGINAL_SIZE} bytes${NC}"

# Optimize with wasm-opt if available
if command -v wasm-opt &> /dev/null; then
    echo -e "${YELLOW}🔧 Optimizing with wasm-opt...${NC}"
    wasm-opt -Oz --enable-bulk-memory main.wasm -o main.wasm.tmp
    mv main.wasm.tmp main.wasm
    OPTIMIZED_SIZE=$(stat -c%s main.wasm 2>/dev/null || stat -f%z main.wasm 2>/dev/null || echo "unknown")
    echo -e "${GREEN}📈 Optimized size: ${OPTIMIZED_SIZE} bytes${NC}"
else
    echo -e "${YELLOW}⚠️  wasm-opt not found, skipping optimization${NC}"
fi

# Compress with gzip
echo -e "${YELLOW}🗜️  Compressing with gzip...${NC}"
gzip -9 -k main.wasm
COMPRESSED_SIZE=$(stat -c%s main.wasm.gz 2>/dev/null || stat -f%z main.wasm.gz 2>/dev/null || echo "unknown")
echo -e "${GREEN}📦 Compressed size: ${COMPRESSED_SIZE} bytes${NC}"

# Generate integrity hash
echo -e "${YELLOW}🔒 Generating integrity hash...${NC}"
if command -v sha256sum &> /dev/null; then
    sha256sum main.wasm | cut -d' ' -f1 > main.wasm.integrity
elif command -v shasum &> /dev/null; then
    shasum -a 256 main.wasm | cut -d' ' -f1 > main.wasm.integrity
else
    echo -e "${YELLOW}⚠️  SHA256 tool not found, skipping integrity hash${NC}"
fi

# Calculate size reduction
if [ "$ORIGINAL_SIZE" != "unknown" ] && [ "$COMPRESSED_SIZE" != "unknown" ]; then
    REDUCTION=$(echo "scale=1; (($ORIGINAL_SIZE - $COMPRESSED_SIZE) * 100) / $ORIGINAL_SIZE" | bc 2>/dev/null || echo "unknown")
    if [ "$REDUCTION" != "unknown" ]; then
        echo -e "${GREEN}💾 Size reduction: ${REDUCTION}%${NC}"
    fi
fi

# Security recommendations
echo -e "\n${BLUE}🔐 Security Notes:${NC}"
echo -e "${YELLOW}• All cryptographic operations use secure random number generation${NC}"
echo -e "${YELLOW}• AES encryption uses GCM mode for authenticated encryption${NC}"
echo -e "${YELLOW}• JWT tokens use HMAC-SHA256 for signing${NC}"
echo -e "${YELLOW}• RSA key generation uses PKCS1v15 padding${NC}"
echo -e "${YELLOW}• bcrypt uses configurable cost factor (default: 10)${NC}"
echo -e "${YELLOW}• Password validation enforces strong security policies${NC}"

echo -e "\n${GREEN}🎉 Build completed successfully!${NC}"
echo -e "${BLUE}📁 Generated files:${NC}"
echo -e "   • main.wasm (WebAssembly module)"
echo -e "   • main.wasm.gz (Compressed version)"
if [ -f "main.wasm.integrity" ]; then
    echo -e "   • main.wasm.integrity (SHA256 hash)"
fi

echo -e "\n${BLUE}🚀 Ready for deployment!${NC}"
