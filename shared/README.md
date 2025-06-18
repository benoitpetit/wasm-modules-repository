# Shared Resources Directory

This directory contains shared resources used by all WASM modules in this project.

## Contents

### `wasm_exec.js`
- **Purpose**: JavaScript runtime support for WebAssembly modules compiled with Go
- **Source**: Copied from Go installation (`$(go env GOROOT)/lib/wasm/wasm_exec.js`)
- **Usage**: All WASM modules reference this single file instead of maintaining individual copies

## Benefits of Shared Approach

- **Single Source of Truth**: One file to maintain and update
- **Reduced Disk Usage**: No duplicate files across modules
- **Consistency**: Guaranteed same version across all modules
- **Easier Maintenance**: Updates only need to be made in one place

## Management Scripts

### Setup
```bash
# Initialize or update the shared wasm_exec.js
./scripts/setup-shared-wasm-exec.sh
```

### Automatic Management
The build scripts automatically handle the shared `wasm_exec.js`:
- Check if the shared file exists
- Create it from Go installation if missing
- Reference the shared location during builds

## Usage in Web Projects

When deploying WASM modules, include both files:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>WASM Module</title>
</head>
<body>
    <script src="path/to/shared/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("module.wasm"), go.importObject)
            .then((result) => {
                go.run(result.instance);
            });
    </script>
</body>
</html>
```

## File Structure

```
wasm-projects/
├── shared/
│   ├── wasm_exec.js     # Shared JavaScript runtime
│   └── README.md        # This file
├── crypto-wasm/
│   ├── main.wasm        # Module binary
│   └── build.sh         # References ../shared/wasm_exec.js
├── math-wasm/
│   ├── main.wasm        # Module binary  
│   └── build.sh         # References ../shared/wasm_exec.js
└── ...
```

## Version Management

The shared `wasm_exec.js` is automatically updated when:
- The Go installation is newer than the shared file
- The setup script is run manually
- A build script detects the file is missing

This ensures compatibility with the Go version used for compilation. 