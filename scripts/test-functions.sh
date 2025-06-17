#!/bin/bash

echo "🧪 Testing getAvailableFunctions in all modules..."
echo "=================================================="

# Function to check if getAvailableFunctions exists in Go source
check_module() {
    local module_dir="$1"
    local module_name="$2"
    
    echo
    echo "=== Testing $module_name ==="
    
    if [ ! -d "$module_dir" ]; then
        echo "❌ Module directory not found: $module_dir"
        return 1
    fi
    
    if [ ! -f "$module_dir/main.go" ]; then
        echo "❌ main.go not found in $module_dir"
        return 1
    fi
    
    if [ ! -f "$module_dir/module.json" ]; then
        echo "❌ module.json not found in $module_dir"
        return 1
    fi
    
    # Check if getAvailableFunctions exists in Go source
    if grep -q "func getAvailableFunctions" "$module_dir/main.go"; then
        echo "✅ getAvailableFunctions function found in main.go"
    else
        echo "❌ getAvailableFunctions function NOT found in main.go"
        return 1
    fi
    
    # Check if it's registered in main()
    if grep -q "js.FuncOf(getAvailableFunctions)" "$module_dir/main.go"; then
        echo "✅ getAvailableFunctions is registered in main()"
    else
        echo "❌ getAvailableFunctions is NOT properly registered"
        return 1
    fi
    
    # Check if it's documented in module.json
    if grep -q '"name": "getAvailableFunctions"' "$module_dir/module.json"; then
        echo "✅ getAvailableFunctions is documented in module.json"
    else
        echo "❌ getAvailableFunctions is NOT documented in module.json"
        return 1
    fi
    
    # Check if setSilentMode exists too
    if grep -q "func setSilentMode" "$module_dir/main.go"; then
        echo "✅ setSilentMode function found"
    else
        echo "❌ setSilentMode function NOT found"
        return 1
    fi
    
    echo "✅ $module_name passes all tests"
    return 0
}

# Navigate to project root
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(dirname "$script_dir")"

cd "$project_root"

# Test all modules
modules_passed=0
modules_total=0

# Test goxios-wasm
((modules_total++))
if check_module "goxios-wasm" "goxios-wasm"; then
    ((modules_passed++))
fi

# Test math-wasm  
((modules_total++))
if check_module "math-wasm" "math-wasm"; then
    ((modules_passed++))
fi

# Test image-wasm
((modules_total++))
if check_module "image-wasm" "image-wasm"; then
    ((modules_passed++))
fi

# Test crypto-wasm
((modules_total++))
if check_module "crypto-wasm" "crypto-wasm"; then
    ((modules_passed++))
fi

# Test qr-wasm
((modules_total++))
if check_module "qr-wasm" "qr-wasm"; then
    ((modules_passed++))
fi

echo
echo "=================================================="
echo "📊 Test Results:"
echo "   Modules tested: $modules_total"
echo "   Modules passed: $modules_passed"
echo "   Modules failed: $((modules_total - modules_passed))"

if [ "$modules_passed" -eq "$modules_total" ]; then
    echo "🎉 All modules have getAvailableFunctions implemented correctly!"
    exit 0
else
    echo "❌ Some modules are missing getAvailableFunctions implementation"
    exit 1
fi
