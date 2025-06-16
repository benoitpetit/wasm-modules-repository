# WASM Modules

Go WebAssembly modules repository.

## Structure

```
├── module1/
├── module2/
```

Each module contains:
- `main.wasm` - compiled binary
- `module.json` - metadata
- `main.go` - source code
- `build.sh` - build script

## Build

```bash
cd module1/
./build.sh
```