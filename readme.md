# WASM Modules

High-performance WebAssembly modules collection written in Go.

## Available Modules

| Module | Description | Functions | Size |
|--------|-------------|-----------|--------|
| **goxios-wasm** | HTTP Client (axios-like) | get, post, put, delete, patch, request, create | 9.9M → 2.7M |
| **math-wasm** | Mathematical calculations | add, subtract, multiply, divide, power, factorial | 2.4M → 688K |
| **image-wasm** | Image processing | compressJPEG, compressPNG, convertToWebP, resizeImage | 3.0M → 852K |

## Quick Start

```bash
# Build all modules with optimizations
./wasm-manager.sh build

# Validate module structure
./wasm-manager.sh validate

# Test function implementations  
./wasm-manager.sh test

# Install optimization tools
./wasm-manager.sh install-tools

# Clean build artifacts
./wasm-manager.sh clean
```

## Individual Module Build

```bash
# Specific module
cd module-name/
./build.sh
```

## Usage

```javascript
// 1. Load WASM module
const go = new Go();
const result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
go.run(result.instance);

// 2. List available functions
const functions = getAvailableFunctions();
console.log('Functions:', functions);

// 3. Use functions
setSilentMode(true); // Silent mode
const result = add(5, 3); // Example math-wasm
```

## Standard Structure

Each module follows this architecture and uniform naming:

```
module-name/
├── main.go         # Optimized Go source code
├── module.json     # Complete metadata and types
├── build.sh        # Build script with optimizations
├── main.wasm       # Optimized WebAssembly binary
├── main.wasm.gz    # Compressed version (production)
├── main.wasm.integrity # SHA256 integrity hash
└── go.mod          # Dependencies
```

### Required Functions

All modules implement standard functions:
- `getAvailableFunctions()` → Array<string> - Lists all available functions
- `setSilentMode(boolean)` → boolean - Enable/disable logging

### Types and Metadata

Each `module.json` contains:
- **Functions** : Complete definitions with parameters and return types
- **Types** : TypeScript structures for auto-completion
- **Security** : Implemented security features
- **Performance** : Applied technical optimizations
- **Compatibility** : Browser and environment support
- **Examples** : Practical usage examples
- **BuildInfo** : Compilation configuration
- **WasmConfig** : WebAssembly parameters

### Module Example

```go
//go:build js && wasm
package main

import "syscall/js"

var silentMode = false

func add(this js.Value, args []js.Value) interface{} {
    return js.ValueOf(args[0].Float() + args[1].Float())
}

func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
    return js.ValueOf([]string{"add", "getAvailableFunctions", "setSilentMode"})
}

func setSilentMode(this js.Value, args []js.Value) interface{} {
    if len(args) == 1 {
        silentMode = args[0].Bool()
    }
    return js.ValueOf(silentMode)
}

func main() {
    js.Global().Set("add", js.FuncOf(add))
    js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
    js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
    select {}
}
```

## Project Structure

```
wasm-projects/
├── wasm-manager.sh          # Main management script
├── scripts/                 # All utility scripts
│   ├── build-all.sh        # Build all modules
│   ├── validate-simple.sh  # Validate modules
│   ├── test-functions.sh   # Test implementations
│   └── install-tools.sh    # Install tools
├── goxios-wasm/            # HTTP client module
├── math-wasm/              # Math operations module
├── image-wasm/             # Image processing module
└── readme.md               # This file
```

## Testing

```bash
# Validate all modules
./wasm-manager.sh validate

# Test function implementations
./wasm-manager.sh test
```

Verifies that each module correctly implements the standard functions.

---

**Applied optimizations:**
- Complete TypeScript types for auto-completion
- Optimized compression: 16MB → 4.1MB (74% reduction)
- SHA256 integrity hash for security
- Standardized build scripts with optimizations
- Complete documentation with practical examples
- Consistent structure across all modules