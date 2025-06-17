#!/bin/bash

# Main management script for WASM projects
# Provides easy access to all project tools

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}🚀 WASM Projects Manager${NC}"
    echo -e "${BLUE}========================${NC}"
}

print_usage() {
    print_header
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Available commands:"
    echo "  build         Build all WASM modules with optimizations"
    echo "  validate      Validate module structure and requirements"
    echo "  test          Test getAvailableFunctions implementation"
    echo "  install-tools Install WASM optimization tools"
    echo "  clean         Clean all build artifacts"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build                    # Build all modules"
    echo "  $0 validate                 # Validate all modules"
    echo "  $0 test                     # Test function implementations"
    echo "  $0 install-tools --check    # Check installed tools"
    echo ""
}

run_script() {
    local script_name="$1"
    shift
    local script_path="./scripts/$script_name"
    
    if [ ! -f "$script_path" ]; then
        echo -e "${RED}❌ Script not found: $script_path${NC}"
        exit 1
    fi
    
    chmod +x "$script_path"
    "$script_path" "$@"
}

clean_artifacts() {
    echo -e "${YELLOW}🧹 Cleaning build artifacts...${NC}"
    
    local cleaned=0
    for module_dir in */; do
        if [ -d "$module_dir" ] && [ -f "$module_dir/main.go" ]; then
            echo "Cleaning $module_dir..."
            rm -f "$module_dir"/*.wasm
            rm -f "$module_dir"/*.wasm.gz
            rm -f "$module_dir"/*.integrity
            ((cleaned++))
        fi
    done
    
    echo -e "${GREEN}✅ Cleaned $cleaned modules${NC}"
}

main() {
    case "${1:-help}" in
        "build")
            shift
            run_script "build-all.sh" "$@"
            ;;
        "validate")
            shift
            run_script "validate-simple.sh" "$@"
            ;;
        "test")
            shift
            run_script "test-functions.sh" "$@"
            ;;
        "install-tools")
            shift
            run_script "install-tools.sh" "$@"
            ;;
        "clean")
            clean_artifacts
            ;;
        "help"|"--help"|"-h")
            print_usage
            ;;
        *)
            echo -e "${RED}❌ Unknown command: $1${NC}"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"
