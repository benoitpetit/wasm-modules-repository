# WASM Modules 🚀

High-performance WebAssembly modules collection written in Go, designed for seamless integration with [GoWM (Go Wasm Manager)](https://github.com/benoitpetit/gowm).

Built with a **Go-based build system** featuring parallel processing, advanced optimizations, and integrated toolchain management.

## Available Modules

| Module | Description | Functions | Original → Optimized → Compressed |
|--------|-------------|-----------|-----------------------------------|
| **goxios-wasm** | HTTP Client (axios-like) | get, post, put, delete, patch, request, create | 9.9M → 2.7M → 891K |
| **math-wasm** | Mathematical calculations | add, subtract, multiply, divide, power, factorial | 2.4M → 688K → 201K |
| **image-wasm** | Image processing | compressJPEG, compressPNG, convertToWebP, resizeImage | 3.0M → 852K → 298K |
| **crypto-wasm** | Cryptographic operations | hashSHA256, encryptAES, generateRSA, JWT, bcrypt, UUID | 6.1M → 1.7M → 487K |
| **qr-wasm** | QR Codes & Barcodes | generateQRCode, generateBarcode, generateVCard, generateWiFiQR | 3.1M → 800K → 267K |
| **text-wasm** | Advanced text processing | textSimilarity, levenshteinDistance, soundex, slugify, camelCase, extractEmails | 3.7M → 3.5M → 1.0M |

## Quick Start

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
# Build all modules (parallel processing)
./wasm-manager build

# Build specific module
./wasm-manager build math-wasm

# Build multiple specific modules
./wasm-manager build math-wasm crypto-wasm qr-wasm

# Build with custom worker count
./wasm-manager build --workers 8

# Build without optimization (faster for development)
./wasm-manager build --optimize=false

# Clean build (removes artifacts first)
./wasm-manager build --clean

# Build with compression disabled
./wasm-manager build --compress=false

# Build without integrity hashes
./wasm-manager build --integrity=false
```

### Available Commands

```bash
# Main commands
./wasm-manager --help                    # Show all available commands
./wasm-manager build                     # Build all modules
./wasm-manager validate                  # Validate all modules  
./wasm-manager test                      # Test all modules
./wasm-manager clean                     # Clean build artifacts
./wasm-manager install-tools             # Install optimization tools
```

| Command | Description | Key Options | Examples |
|---------|-------------|-------------|----------|
| **build** | Build WASM modules with optimizations | `--workers`, `--no-optimize`, `--clean` | `./wasm-manager build math-wasm --workers 8` |
| **validate** | Validate module structure and compliance | `--strict`, `--fix` | `./wasm-manager validate --strict` |
| **test** | Test function implementations | `--integration`, `--coverage` | `./wasm-manager test --integration` |
| **clean** | Clean build artifacts and caches | `--all`, `--cache` | `./wasm-manager clean --all` |
| **install-tools** | Install WASM optimization tools | `--check`, `--force`, `--binaryen` | `./wasm-manager install-tools --check` |

## Build System Features

### ⚡ Parallel Processing
- **Worker Pools**: Configurable number of parallel builds (default: CPU cores)
- **Concurrent Operations**: Multiple modules build simultaneously
- **Smart Scheduling**: Optimal resource utilization
- **Error Isolation**: Failed builds don't stop others

### 🛡️ Advanced Optimizations
- **WASM optimization** using wasm-opt with intelligent flags
- **Compression pipeline** with gzip and brotli
- **Integrity verification** with SHA256 hashes
- **Size analysis** and compression reporting

### 📊 Performance
- **5-10x faster builds** compared to sequential processing
- **Full CPU utilization** with worker pools
- **Type-safe operations** with Go's type system
- **Robust error handling** and recovery

## Usage with GoWM

### Installation

```bash
npm install gowm
```

### Basic Usage

```javascript
import { loadFromGitHub } from 'gowm';

// Load math module
const math = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  branch: 'master',
  name: 'math-wasm'
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
const math = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  branch: 'master',
  name: 'math-wasm'
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
const crypto = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  branch: 'master',
  name: 'crypto-wasm'
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

#### Text Processing Module

```javascript
// Load text processing module
const text = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  branch: 'master',
  name: 'text-wasm'
});

// Configure module
text.call('setSilentMode', true);

// Text similarity analysis
const similarity = text.call('textSimilarity', 'hello world', 'hello earth');
console.log('Similarity:', similarity); // ~0.75

// String case conversions
console.log('camelCase:', text.call('camelCase', 'hello world test')); // helloWorldTest
console.log('kebab-case:', text.call('kebabCase', 'HelloWorldTest')); // hello-world-test
console.log('snake_case:', text.call('snakeCase', 'HelloWorldTest')); // hello_world_test

// Extract information from text
const sampleText = 'Contact us at support@example.com or visit https://example.com';
const emails = text.call('extractEmails', sampleText);
const urls = text.call('extractURLs', sampleText);

console.log('Emails found:', emails); // ['support@example.com']
console.log('URLs found:', urls); // ['https://example.com']

// Text analysis
const wordCount = text.call('wordCount', sampleText);
const readingTime = text.call('readingTime', sampleText, 200);
console.log('Words:', wordCount, 'Reading time:', readingTime.minutes, 'minutes');

// Advanced text processing
const withAccents = 'Café naïve résumé';
const withoutAccents = text.call('removeDiacritics', withAccents);
const slug = text.call('slugify', withAccents);
console.log('Without accents:', withoutAccents); // Cafe naive resume
console.log('Slug:', slug); // cafe-naive-resume
```

#### QR Module

```javascript
// Load QR module
const qr = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  branch: 'master',
  name: 'qr-wasm'
});

// Enable silent mode for production
qr.call('setSilentMode', true);

// Generate basic QR code
const qrResult = qr.call('generateQRCode', 'Hello QR World!', 256, 'HIGH');
if (qrResult.error) {
  console.error('QR generation failed:', qrResult.error);
} else {
  console.log('QR Base64:', qrResult.base64Image);
  // Display QR code
  const img = document.createElement('img');
  img.src = 'data:image/png;base64,' + qrResult.base64Image;
  document.body.appendChild(img);
}

// Generate vCard contact QR
const contact = {
  name: 'John Doe',
  organization: 'Tech Corp',
  phone: '+1234567890',
  email: 'john@example.com',
  url: 'https://johndoe.com'
};
const vCardResult = qr.call('generateVCard', contact, 300);

// Generate WiFi connection QR
const wifi = {
  ssid: 'MyNetwork',
  password: 'mypassword',
  security: 'WPA',
  hidden: false
};
const wifiResult = qr.call('generateWiFiQR', wifi, 256);

// Generate barcode
const barcodeResult = qr.call('generateBarcode', '1234567890128', 'ean13', 400, 200);
if (barcodeResult.error) {
  console.error('Barcode generation failed:', barcodeResult.error);
} else {
  console.log('Barcode type:', barcodeResult.type);
  console.log('Dimensions:', barcodeResult.width + 'x' + barcodeResult.height);
}
```

#### Image Processing

```javascript
// Load image module
const image = await loadFromGitHub('benoitpetit/wasm-modules-repository', {
  branch: 'master',
  name: 'image-wasm'
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

function QRGenerator() {
  const { wasm: qr, loading, error } = useWasmFromGitHub('benoitpetit/wasm-modules-repository', {
    branch: 'master',
    name: 'qr-wasm'
  });
  
  const [qrImage, setQrImage] = useState('');
  const [text, setText] = useState('Hello QR World!');

  useEffect(() => {
    if (qr) {
      qr.call('setSilentMode', true);
    }
  }, [qr]);

  const generateQR = () => {
    if (!qr) return;
    
    const result = qr.call('generateQRCode', text, 256, 'HIGH');
    if (result.error) {
      console.error('QR generation error:', result.error);
    } else {
      setQrImage('data:image/png;base64,' + result.base64Image);
    }
  };

  return (
    <div>
      <input 
        value={text} 
        onChange={(e) => setText(e.target.value)}
        placeholder="Enter text for QR code" 
      />
      <button onClick={generateQR} disabled={!qr}>
        Generate QR Code
      </button>
      {qrImage && <img src={qrImage} alt="Generated QR Code" />}
    </div>
  );
}
```

## Build Configuration

### Environment Variables
```bash
# Number of worker goroutines (default: CPU cores)
export WASM_WORKERS=8

# Enable verbose output
export WASM_VERBOSE=true

# Optimization level
export WASM_OPTIMIZE=true
```

### Configuration File (.wasm-manager.yaml)
```yaml
workers: 8
optimize: true
compress: true
generateIntegrity: true
verbose: false
timeout: 10m
```

## Development Workflow

### Complete Development Workflow

#### Setup and Build Manager
```bash
# Initial setup
make setup                               # Install dependencies and build manager
# OR manually:
go mod tidy && go build -o wasm-manager .

# Rebuild manager after changes
make build                               # Quick rebuild
```

#### Development Cycle
```bash
# 1. Build and test specific module
./wasm-manager build math-wasm --verbose
./wasm-manager test math-wasm

# 2. Validate module compliance
./wasm-manager validate math-wasm --strict

# 3. Clean artifacts when needed
./wasm-manager clean math-wasm

# 4. Full rebuild after changes
./wasm-manager build math-wasm --clean
```

#### Production Build
```bash
# Build all modules for production
./wasm-manager clean --all               # Clean everything
./wasm-manager install-tools --check     # Verify tools
./wasm-manager build --workers 8         # Parallel build
./wasm-manager validate --strict         # Final validation
./wasm-manager test --integration        # Integration tests
```

### Advanced Usage

#### Validation Commands
```bash
# Validate all modules
./wasm-manager validate

# Validate specific module
./wasm-manager validate math-wasm

# Strict validation with enhanced checks
./wasm-manager validate --strict

# Auto-fix validation issues
./wasm-manager validate --fix
```

#### Testing Commands
```bash
# Test all modules
./wasm-manager test

# Test specific module
./wasm-manager test crypto-wasm

# Run integration tests
./wasm-manager test --integration

# Generate test coverage report
./wasm-manager test --coverage
```

#### Cleaning Commands
```bash
# Clean build artifacts for all modules
./wasm-manager clean

# Clean specific module
./wasm-manager clean math-wasm

# Clean everything including caches
./wasm-manager clean --all

# Clean only caches
./wasm-manager clean --cache
```

#### Tool Management
```bash
# Install all optimization tools
./wasm-manager install-tools

# Check current tool installations
./wasm-manager install-tools --check

# Force reinstall tools
./wasm-manager install-tools --force

# Install only Binaryen (wasm-opt)
./wasm-manager install-tools --binaryen

# Install only WABT toolkit
./wasm-manager install-tools --wabt
```

#### Global Options
```bash
# Verbose output for debugging
./wasm-manager build --verbose

# Custom worker count
./wasm-manager build --workers 12

# Use configuration file
./wasm-manager build --config custom-config.yaml

# Show version
./wasm-manager --version
```

## Performance & Optimization

### Build Performance
- **Parallel Processing**: 5-10x faster than sequential builds
- **Resource Utilization**: Full CPU core utilization
- **Memory Efficiency**: Optimized goroutine pools
- **Smart Caching**: Avoid redundant operations

### WASM Optimization
- **Size Reduction**: 50-70% reduction through optimization
- **Compression**: Additional 60-80% reduction with gzip/brotli
- **Integrity**: SHA256 verification for security
- **Compatibility**: Maintained across all optimizations

### Build Output Example
```
🚀 Building 5 modules with 8 workers

✅ math-wasm        2.4M → 688K → 201K (1.2s)
✅ crypto-wasm      6.1M → 1.7M → 487K (2.8s)  
✅ image-wasm       3.0M → 852K → 298K (1.9s)
✅ qr-wasm          3.1M → 800K → 267K (1.7s)
✅ goxios-wasm      9.9M → 2.7M → 891K (3.1s)

📊 Statistics:
   Successful: 5
   Failed: 0
   Total time: 3.1s
   Size reduction: 18.3M → 6.7M (63.4%)
   Compression ratio: 26.8%
```

## Contributing

The Go-based build system makes it easy to contribute:

1. **Add new commands**: Implement in `cmd/` directory
2. **Extend functionality**: Add to `internal/` packages
3. **Custom optimizations**: Enhance the build pipeline
4. **New modules**: Follow the existing structure

### Project Structure
```
wasm-projects/
├── main.go                  # Entry point
├── go.mod                   # Dependencies
├── Makefile                 # Build automation
├── cmd/                     # CLI commands
├── internal/                # Core functionality
├── [module-name]/           # WASM modules
│   ├── main.go
│   ├── go.mod
│   └── module.json
└── shared/                  # Shared resources
```

## License

MIT License

---

**Built with Go for Performance and Reliability** 🚀

This project leverages Go's concurrency model with goroutines and worker pools to deliver high-performance, parallel WASM builds that are 5-10x faster than traditional sequential approaches.