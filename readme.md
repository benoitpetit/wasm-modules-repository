# WASM Modules

High-performance WebAssembly modules collection written in Go, designed for seamless integration with [GoWM (Go WebAssembly Manager)](https://github.com/benoitpetit/gowm).

## Available Modules

| Module | Description | Functions | Size |
|--------|-------------|-----------|--------|
| **goxios-wasm** | HTTP Client (axios-like) | get, post, put, delete, patch, request, create | 9.9M → 2.7M |
| **math-wasm** | Mathematical calculations | add, subtract, multiply, divide, power, factorial | 2.4M → 688K |
| **image-wasm** | Image processing | compressJPEG, compressPNG, convertToWebP, resizeImage | 3.0M → 852K |
| **crypto-wasm** | Cryptographic operations | hashSHA256, encryptAES, generateRSA, JWT, bcrypt, UUID | 6.1M → 1.7M |

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

## Usage with GoWM

### Installation

```bash
npm install gowm
```

### Basic Usage

```javascript
import { loadFromGitHub } from 'gowm';

// Load math module
const math = await loadFromGitHub('your-org/wasm-projects', {
  path: 'math-wasm',
  name: 'math'
});

// Configure module
math.call('setSilentMode', true);

// Get available functions
const functions = math.call('getAvailableFunctions');
console.log('Available functions:', functions);

// Use math functions with error handling
const result = math.call('add', 5, 3);
if (typeof result === 'string' && result.includes('Erreur')) {
  console.error('Math error:', result);
} else {
  console.log('5 + 3 =', result); // 8
}
```

### Advanced Examples

#### Math Module

```javascript
// Load math module
const math = await loadFromGitHub('your-org/wasm-projects', {
  path: 'math-wasm',
  filename: 'main.wasm',
  name: 'math'
});

// Enable silent mode for production
math.call('setSilentMode', true);

// Perform calculations with error handling
const operations = [
  { func: 'add', args: [10, 5] },
  { func: 'divide', args: [10, 0] }, // Will return error
  { func: 'factorial', args: [5] },
  { func: 'power', args: [2, 8] }
];

operations.forEach(({ func, args }) => {
  const result = math.call(func, ...args);
  if (typeof result === 'string' && result.includes('Erreur')) {
    console.error(`${func}(${args.join(', ')}) failed:`, result);
  } else {
    console.log(`${func}(${args.join(', ')}) =`, result);
  }
});
```

#### Crypto Module

```javascript
// Load crypto module
const crypto = await loadFromGitHub('your-org/wasm-projects', {
  path: 'crypto-wasm',
  name: 'crypto'
});

// Hash operations
const hashResult = crypto.call('hashSHA256', 'Hello World');
if (hashResult.error) {
  console.error('Hash error:', hashResult.error);
} else {
  console.log('SHA256:', hashResult.hash);
  console.log('Algorithm:', hashResult.algorithm);
}

// AES encryption
const keyResult = crypto.call('generateAESKey', 32); // 256-bit
if (keyResult.error) {
  console.error('Key generation failed:', keyResult.error);
} else {
  const encryptResult = crypto.call('encryptAES', 'Secret message', keyResult.key);
  if (encryptResult.error) {
    console.error('Encryption failed:', encryptResult.error);
  } else {
    console.log('Encrypted data:', encryptResult.encryptedData);
    
    // Decrypt
    const decryptResult = crypto.call('decryptAES', encryptResult.encryptedData, keyResult.key);
    if (decryptResult.error) {
      console.error('Decryption failed:', decryptResult.error);
    } else {
      console.log('Decrypted:', decryptResult.decryptedData);
    }
  }
}
```

#### Image Module

```javascript
// Load image module
const image = await loadFromGitHub('your-org/wasm-projects', {
  path: 'image-wasm',
  name: 'image'
});

// Process image with error handling
const imageData = /* your image data */;
const compressResult = image.call('compressJPEG', imageData, 0.8);

if (compressResult.error) {
  console.error('Compression failed:', compressResult.error);
} else {
  console.log('Original size:', compressResult.originalSize);
  console.log('Compressed size:', compressResult.compressedSize);
  console.log('Compression ratio:', compressResult.compressionRatio);
  // Access compressed data: compressResult.data
}
```

### React Integration

```jsx
import React, { useState, useEffect } from 'react';
import { useWasmFromGitHub } from 'gowm/hooks/useWasm';

function MathCalculator() {
  const { wasm: math, loading, error } = useWasmFromGitHub('your-org/wasm-projects', {
    path: 'math-wasm',
    name: 'math'
  });
  
  const [result, setResult] = useState(null);
  const [functions, setFunctions] = useState([]);

  useEffect(() => {
    if (math) {
      math.call('setSilentMode', true);
      const availableFunctions = math.call('getAvailableFunctions');
      setFunctions(availableFunctions);
    }
  }, [math]);

  const calculate = (operation, a, b) => {
    if (!math) return;
    
    const result = math.call(operation, a, b);
    if (typeof result === 'string' && result.includes('Erreur')) {
      setResult(`Error: ${result}`);
    } else {
      setResult(result);
    }
  };

  if (loading) return <div>Loading math module...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <h3>Available Functions: {functions.join(', ')}</h3>
      <button onClick={() => calculate('add', 10, 5)}>10 + 5</button>
      <button onClick={() => calculate('factorial', 5)}>5!</button>
      {result !== null && <div>Result: {result}</div>}
    </div>
  );
}
```

### Vue.js Integration

```vue
<template>
  <div>
    <div v-if="loading">Loading crypto module...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
      <h3>Crypto Operations</h3>
      <input v-model="message" placeholder="Message to hash" />
      <button @click="hashMessage">Hash SHA256</button>
      <div v-if="hashResult">
        <strong>Hash:</strong> {{ hashResult.hash }}<br>
        <strong>Algorithm:</strong> {{ hashResult.algorithm }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useWasmFromGitHub } from 'gowm/composables/useWasm';

const { wasm: crypto, loading, error } = useWasmFromGitHub('your-org/wasm-projects', {
  path: 'crypto-wasm',
  name: 'crypto'
});

const message = ref('Hello World');
const hashResult = ref(null);

const hashMessage = () => {
  if (!crypto.value) return;
  
  const result = crypto.value.call('hashSHA256', message.value);
  if (result.error) {
    console.error('Hash error:', result.error);
  } else {
    hashResult.value = result;
  }
};
</script>
```

## Individual Module Build

```bash
# Specific module
cd module-name/
./build.sh
```

## Direct WASM Usage (without GoWM)

For direct usage without GoWM:

```javascript
// 1. Load WASM module manually
const go = new Go();
const result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
go.run(result.instance);

// 2. Wait for module ready signal
await new Promise(resolve => {
  const checkReady = () => {
    if (globalThis.__gowm_ready) {
      resolve();
    } else {
      setTimeout(checkReady, 10);
    }
  };
  checkReady();
});

// 3. List available functions
const functions = getAvailableFunctions();
console.log('Functions:', functions);

// 4. Use functions
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
- `__gowm_ready` global signal - Indicates module is ready

### GoWM Compatibility

All modules are designed for seamless GoWM integration:
- **Automatic Discovery**: GoWM can find `main.wasm` files automatically
- **Error Patterns**: Consistent error handling across modules
- **Function Introspection**: Standard `getAvailableFunctions()` implementation
- **Silent Mode**: Production-ready logging control
- **Ready Signal**: `__gowm_ready` for synchronization

### Types and Metadata

Each `module.json` contains:
- **Functions** : Complete definitions with parameters and return types
- **GoWM Config** : Configuration for optimal GoWM integration
- **Types** : TypeScript structures for auto-completion
- **Security** : Implemented security features
- **Performance** : Applied technical optimizations
- **Compatibility** : Browser and environment support
- **Examples** : Practical usage examples with GoWM
- **BuildInfo** : Compilation configuration
- **WasmConfig** : WebAssembly parameters

### Module Example

```go
//go:build js && wasm
package main

import (
    "fmt"
    "syscall/js"
)

var silentMode = false

func add(this js.Value, args []js.Value) interface{} {
    if len(args) != 2 {
        return js.ValueOf("Erreur: deux arguments requis pour add")
    }
    
    a := args[0].Float()
    b := args[1].Float()
    result := a + b
    
    if !silentMode {
        fmt.Printf("Go WASM: %f + %f = %f\n", a, b, result)
    }
    return js.ValueOf(result)
}

func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
    functions := []interface{}{
        "add", "getAvailableFunctions", "setSilentMode",
    }
    return js.ValueOf(functions)
}

func setSilentMode(this js.Value, args []js.Value) interface{} {
    if len(args) == 1 {
        silentMode = args[0].Bool()
    }
    return js.ValueOf(silentMode)
}

func main() {
    fmt.Println("Go WASM Module initializing...")
    
    // Register functions
    js.Global().Set("add", js.FuncOf(add))
    js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
    js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
    
    // Signal ready for GoWM
    js.Global().Set("__gowm_ready", js.ValueOf(true))
    
    fmt.Println("Go WASM Module ready!")
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
├── crypto-wasm/            # Cryptographic operations module
└── readme.md               # This file
```

## Testing

```bash
# Validate all modules
./wasm-manager.sh validate

# Test function implementations
./wasm-manager.sh test
```

Verifies that each module correctly implements the standard functions and GoWM compatibility.

## Integration with Website

These modules are showcased in the [WASM Manager website](../website) which provides:
- **Module Discovery**: Browse all available modules
- **Integration Examples**: Copy-paste GoWM integration code  
- **GitHub Integration**: Direct links to module repositories
- **Documentation**: Complete API reference and usage guides

---

**Applied optimizations:**
- **GoWM Integration**: Seamless integration with Go WebAssembly Manager
- **Error Handling**: Consistent error patterns across all modules
- **Function Discovery**: Standard introspection capabilities
- **Complete TypeScript types** for auto-completion
- **Optimized compression**: 16MB → 4.1MB (74% reduction)
- **SHA256 integrity hash** for security
- **Standardized build scripts** with optimizations
- **Complete documentation** with practical GoWM examples
- **Consistent structure** across all modules