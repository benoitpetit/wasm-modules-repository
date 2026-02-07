# WASM Modules Repository 🚀

High-performance WebAssembly modules collection written in Go, designed for use with [GoWM (Go Wasm Manager)](https://github.com/benoitpetit/gowm).

Built with a **Go-based build system** featuring parallel processing, advanced optimizations, and integrated toolchain management.

## Available Modules

| Module           | Description                   | Functions                                                                                       | Size (wasm → gzip) |
| ---------------- | ----------------------------- | ----------------------------------------------------------------------------------------------- | ------------------ |
| **math-wasm**    | Mathematical calculations     | add, subtract, multiply, divide, power, factorial, sqrt, gcd, fibonacci, mean, median... (25)   | 2.5M → 720K        |
| **crypto-wasm**  | Cryptographic operations      | hashSHA256, encryptAES, generateRSAKeyPair, generateJWT, bcryptHash, generateUUID... (19)       | 6.1M → 1.7M        |
| **text-wasm**    | Advanced text processing      | textSimilarity, levenshtein, slugify, camelCase, extractEmails, wordCount, readingTime... (17)  | 3.8M → 1.1M        |
| **image-wasm**   | Image processing              | compressJPEG, compressPNG, convertToWebP, resizeImage, getImageInfo (5)                         | 3.0M → 864K        |
| **qr-wasm**      | QR Codes & Barcodes           | generateQRCode, decodeQRCode, generateBarcode, decodeBarcode, generateVCard, generateWiFiQR (6) | 3.3M → 920K        |
| **pdf-wasm**     | PDF generation & manipulation | createPDF, mergePDFs, splitPDF, generateInvoice, generateReport, htmlToPDF, analyzePDF... (18)  | 5.4M → 1.5M        |
| **jsonxml-wasm** | JSON/XML/CSV/YAML conversion  | parseJSON, validateJSON, parseXML, xmlToJSON, csvToJSON, yamlToJSON, jsonToYAML... (14)         | 7.5M → 2.3M        |
| **goxios-wasm**  | HTTP client (axios-like)      | get, post, put, delete, patch, request, create, setDefaults (8)                                 | 11M → 2.7M         |

Each module also exposes the utility functions `setSilentMode`, `getAvailableFunctions`, and `getModuleInfo`.

## Quick Start

### Prerequisites

- Go 1.21+
- (Optional) [Binaryen](https://github.com/WebAssembly/binaryen) for wasm-opt optimization

### Setup

```bash
# Install dependencies and build the manager
make setup

# Or manually:
go mod tidy
go build -o wasm-manager .
```

### Building Modules

```bash
# Build all modules (parallel)
./wasm-manager build

# Build a specific module
./wasm-manager build math-wasm

# Build multiple modules
./wasm-manager build math-wasm crypto-wasm qr-wasm

# Build without optimization (faster for development)
./wasm-manager build --optimize=false

# Clean build
./wasm-manager build --clean

# Custom worker count
./wasm-manager build --workers 8
```

### Available Commands

| Command         | Description                | Key Options                                                       |
| --------------- | -------------------------- | ----------------------------------------------------------------- |
| `build`         | Build WASM modules         | `--workers`, `--optimize`, `--clean`, `--compress`, `--integrity` |
| `validate`      | Validate module structure  | `--strict`, `--fix`                                               |
| `test`          | Test implementations       | `--integration`, `--coverage`                                     |
| `clean`         | Remove build artifacts     | `--all`, `--cache`                                                |
| `install-tools` | Install optimization tools | `--check`, `--force`, `--binaryen`                                |

```bash
./wasm-manager --help    # Full help
```

## Build System

### Parallel Processing

- **Configurable worker pools** (defaults to CPU core count)
- Simultaneous multi-module builds
- Error isolation: a failed build doesn't stop others

### Optimization Pipeline

- WASM optimization via `wasm-opt` (Binaryen)
- Gzip and brotli compression
- SHA256 hash generation (`.wasm.integrity` files)
- Size reporting and compression ratios

### Example Output

```
🚀 Building 8 modules with 8 workers

✅ math-wasm        2.5M → 720K (30s)
✅ crypto-wasm      6.1M → 1.7M (40s)
✅ image-wasm       3.0M → 864K (35s)
✅ qr-wasm          3.3M → 920K (36s)
✅ text-wasm        3.8M → 1.1M (38s)
✅ pdf-wasm         5.4M → 1.5M (40s)
✅ jsonxml-wasm     7.5M → 2.3M (44s)
✅ goxios-wasm       11M → 2.7M (45s)

📊 Statistics:
   Successful: 8
   Failed: 0
   Compression ratio: 27%
```

## Module Structure

Each WASM module follows the same structure:

```
module-wasm/
├── main.go              # Go source code
├── go.mod               # Go dependencies
├── module.json          # Metadata (functions, examples, config)
├── main.wasm            # Compiled binary
├── main.wasm.gz         # Gzip compressed version
└── main.wasm.integrity  # SHA256 hash (SRI)
```

### module.json

The `module.json` file describes the module for GoWM integration:

- **Metadata**: name, version, description, author, license
- **Functions**: list with parameters, return types, descriptions
- **Examples**: ready-to-use code snippets
- **GoWM config**: readySignal, standard functions, auto-detection
- **Build info**: language, build command, target

## Development Workflow

```bash
# 1. Edit the Go source code of a module
# 2. Build and test
./wasm-manager build math-wasm --verbose
./wasm-manager test math-wasm

# 3. Validate compliance
./wasm-manager validate math-wasm --strict

# 4. Full production build
./wasm-manager clean --all
./wasm-manager install-tools --check
./wasm-manager build --workers 8
./wasm-manager validate --strict
```

## Creating a New Module

1. Create a `my-module-wasm/` directory
2. Add `main.go` with functions exported via `js.Global().Set()`
3. Add `go.mod` with the Go module
4. Add `module.json` with metadata
5. End `main()` with `js.Global().Set("__gowm_ready", js.ValueOf(true))` and a blocking channel
6. Build with `./wasm-manager build my-module-wasm`

### Minimal Template

```go
//go:build js && wasm

package main

import (
    "fmt"
    "syscall/js"
)

func myFunction(this js.Value, args []js.Value) interface{} {
    // Implementation
    return js.ValueOf("result")
}

func main() {
    c := make(chan struct{})

    js.Global().Set("myFunction", js.FuncOf(myFunction))
    js.Global().Set("__gowm_ready", js.ValueOf(true))

    fmt.Println("Module loaded")
    <-c
}
```

## Project Structure

```
wasm-modules-repository/
├── main.go              # Build manager entry point
├── go.mod               # Dependencies
├── Makefile             # Build automation
├── wasm-manager         # Compiled manager binary
├── cmd/                 # CLI commands (build, clean, test, validate, install)
├── internal/            # Internal logic (builder, cleaner, tester, validator)
├── shared/              # Shared resources (wasm_exec.js)
└── *-wasm/              # Individual WASM modules
```

## License

MIT
