#!/bin/bash

# Simple validation script for WASM modules

echo "🔍 WASM Module Validation"
echo "========================="

validate_module() {
    local module_dir="$1"
    local module_name=$(basename "$module_dir")
    
    echo
    echo "=== $module_name ==="
    
    # Check required files
    local errors=0
    
    [ -f "$module_dir/main.go" ] && echo "✅ main.go" || { echo "❌ main.go"; ((errors++)); }
    [ -f "$module_dir/module.json" ] && echo "✅ module.json" || { echo "❌ module.json"; ((errors++)); }
    [ -f "$module_dir/build.sh" ] && echo "✅ build.sh" || { echo "❌ build.sh"; ((errors++)); }
    [ -f "$module_dir/go.mod" ] && echo "✅ go.mod" || { echo "❌ go.mod"; ((errors++)); }
    [ -f "$module_dir/main.wasm" ] && echo "✅ main.wasm" || echo "⚠️  main.wasm (not built)"
    [ -f "$module_dir/main.wasm.gz" ] && echo "✅ main.wasm.gz" || echo "⚠️  main.wasm.gz"
    [ -f "$module_dir/main.wasm.integrity" ] && echo "✅ integrity hash" || echo "⚠️  integrity hash"
    
    # Check module.json structure
    if [ -f "$module_dir/module.json" ]; then
        if jq -e '.types' "$module_dir/module.json" >/dev/null 2>&1; then
            echo "✅ Types defined"
        else
            echo "❌ Types missing"
            ((errors++))
        fi
        
        if jq -e '.security' "$module_dir/module.json" >/dev/null 2>&1; then
            echo "✅ Security documented"
        else
            echo "❌ Security missing"
            ((errors++))
        fi
        
        if jq -e '.buildInfo' "$module_dir/module.json" >/dev/null 2>&1; then
            echo "✅ BuildInfo present"
        else
            echo "❌ BuildInfo missing"
            ((errors++))
        fi
    fi
    
    # Check Go functions
    if [ -f "$module_dir/main.go" ]; then
        if grep -q "func getAvailableFunctions" "$module_dir/main.go"; then
            echo "✅ getAvailableFunctions"
        else
            echo "❌ getAvailableFunctions"
            ((errors++))
        fi
        
        if grep -q "func setSilentMode" "$module_dir/main.go"; then
            echo "✅ setSilentMode"
        else
            echo "❌ setSilentMode"
            ((errors++))
        fi
    fi
    
    return $errors
}

# Main validation
total=0
passed=0

# Navigate to parent directory to find modules
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(dirname "$script_dir")"

cd "$project_root"

for dir in */; do
    if [ -f "$dir/build.sh" ] && [ -f "$dir/main.go" ]; then
        ((total++))
        if validate_module "$dir"; then
            ((passed++))
        fi
    fi
done

echo
echo "========================="
echo "📊 Results: $passed/$total modules valid"

if [ $passed -eq $total ]; then
    echo "🎉 All modules are compliant!"
    exit 0
else
    echo "❌ Some modules require corrections"
    exit 1
fi
