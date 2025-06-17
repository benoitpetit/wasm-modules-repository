#!/bin/bash

# Global build script for all WASM modules
# This script builds all modules with optimizations and security features

set -e  # Exit on error

echo "🏗️  Building all WASM modules with optimizations..."
echo "=================================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

# Check dependencies
check_dependencies() {
    print_header "Checking Dependencies"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    print_status "Go version: $GO_VERSION"
    
    # Check optional tools
    if command -v wasm-opt &> /dev/null; then
        WASM_OPT_VERSION=$(wasm-opt --version | head -n1)
        print_status "wasm-opt available: $WASM_OPT_VERSION"
    else
        print_warning "wasm-opt not found - install binaryen for better optimization"
    fi
    
    if command -v gzip &> /dev/null; then
        print_status "gzip available for compression"
    else
        print_warning "gzip not found - compression will be skipped"
    fi
}

# Build a single module
build_module() {
    local module_dir=$1
    local module_name=$(basename "$module_dir")
    
    print_header "Building $module_name"
    
    if [ ! -d "$module_dir" ]; then
        print_error "Module directory $module_dir not found"
        return 1
    fi
    
    if [ ! -f "$module_dir/build.sh" ]; then
        print_error "Build script not found in $module_dir"
        return 1
    fi
    
    # Change to module directory and run build
    cd "$module_dir"
    
    # Make build script executable
    chmod +x build.sh
    
    # Run the build script
    if ./build.sh; then
        print_status "✅ $module_name built successfully"
        
        # Collect build artifacts info
        if [ -f "main.wasm" ]; then
            local size=$(du -h main.wasm | cut -f1)
            local size_bytes=$(stat -c%s main.wasm)
            echo "   📊 Size: $size ($size_bytes bytes)"
            
            if [ -f "main.wasm.gz" ]; then
                local gzip_size=$(du -h main.wasm.gz | cut -f1)
                echo "   📦 Compressed: $gzip_size"
            fi
            
            if [ -f "main.wasm.integrity" ]; then
                local hash=$(cat main.wasm.integrity)
                echo "   🔑 Integrity: $hash"
            fi
        fi
    else
        print_error "❌ $module_name build failed"
        return 1
    fi
    
    cd - > /dev/null
}

# Generate global build report
generate_report() {
    print_header "Build Report"
    
    local total_size=0
    local total_compressed=0
    local module_count=0
    
    echo "| Module | Original Size | Compressed Size | Compression Ratio |"
    echo "|--------|---------------|-----------------|-------------------|"
    
    for module_dir in */; do
        if [ -f "$module_dir/main.wasm" ]; then
            local module_name=$(basename "$module_dir")
            local size_bytes=$(stat -c%s "$module_dir/main.wasm")
            local size_human=$(du -h "$module_dir/main.wasm" | cut -f1)
            
            total_size=$((total_size + size_bytes))
            module_count=$((module_count + 1))
            
            local compressed_info="N/A"
            local compression_ratio="N/A"
            
            if [ -f "$module_dir/main.wasm.gz" ]; then
                local compressed_bytes=$(stat -c%s "$module_dir/main.wasm.gz")
                local compressed_human=$(du -h "$module_dir/main.wasm.gz" | cut -f1)
                total_compressed=$((total_compressed + compressed_bytes))
                
                local ratio=$((compressed_bytes * 100 / size_bytes))
                compressed_info="$compressed_human"
                compression_ratio="${ratio}%"
            fi
            
            printf "| %-6s | %-13s | %-15s | %-17s |\n" \
                   "$module_name" "$size_human" "$compressed_info" "$compression_ratio"
        fi
    done
    
    echo ""
    print_status "Total modules built: $module_count"
    
    if [ $total_size -gt 0 ]; then
        local total_human=$(numfmt --to=iec-i --suffix=B $total_size)
        print_status "Total uncompressed size: $total_human"
        
        if [ $total_compressed -gt 0 ]; then
            local compressed_human=$(numfmt --to=iec-i --suffix=B $total_compressed)
            local overall_ratio=$((total_compressed * 100 / total_size))
            print_status "Total compressed size: $compressed_human ($overall_ratio% of original)"
        fi
    fi
}

# Main execution
main() {
    local start_time=$(date +%s)
    
    # Navigate to project root
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    project_root="$(dirname "$script_dir")"
    
    cd "$project_root"
    
    # Check dependencies first
    check_dependencies
    
    # Store current directory
    local original_dir=$(pwd)
    
    # Find all module directories (containing build.sh)
    local modules=()
    for dir in */; do
        if [ -f "$dir/build.sh" ]; then
            modules+=("$dir")
        fi
    done
    
    if [ ${#modules[@]} -eq 0 ]; then
        print_error "No modules with build.sh found"
        exit 1
    fi
    
    print_status "Found ${#modules[@]} modules to build"
    
    # Build each module
    local failed_modules=()
    for module in "${modules[@]}"; do
        if ! build_module "$module"; then
            failed_modules+=("$module")
        fi
    done
    
    # Return to original directory
    cd "$original_dir"
    
    # Generate build report
    generate_report
    
    # Calculate build time
    local end_time=$(date +%s)
    local build_time=$((end_time - start_time))
    
    print_header "Build Summary"
    
    if [ ${#failed_modules[@]} -eq 0 ]; then
        print_status "🎉 All modules built successfully in ${build_time}s"
    else
        print_error "❌ ${#failed_modules[@]} modules failed to build:"
        for failed in "${failed_modules[@]}"; do
            echo "   - $failed"
        done
        exit 1
    fi
    
    # Installation instructions
    echo ""
    echo "📋 Next Steps:"
    echo "   1. Test your modules in a web browser"
    echo "   2. Use the .wasm.gz files for production (smaller downloads)"
    echo "   3. Verify integrity using the .integrity files"
    echo "   4. Consider setting up automated builds with these optimizations"
    echo ""
    print_status "Build completed successfully! 🚀"
}

# Run main function
main "$@"
