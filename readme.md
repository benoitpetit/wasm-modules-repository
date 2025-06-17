# WASM Modules

Collection de modules WebAssembly haute performance écrits en Go.

## Modules Disponibles

| Module | Description | Fonctions |
|--------|-------------|-----------|
| **goxios-wasm** | Client HTTP (axios-like) | get, post, put, delete, patch, request, create |
| **math-wasm** | Calculs mathématiques | add, subtract, multiply, divide, power, factorial |
| **image-wasm** | Traitement d'images | compressJPEG, compressPNG, convertToWebP, resizeImage |

## Build

```bash
# Tous les modules
./build-all.sh

# Module spécifique
cd module-name/
./build.sh
```

## Utilisation

```javascript
// 1. Charger le module WASM
const go = new Go();
const result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
go.run(result.instance);

// 2. Lister les fonctions disponibles
const functions = getAvailableFunctions();
console.log('Fonctions:', functions);

// 3. Utiliser les fonctions
setSilentMode(true); // Mode silencieux
const result = add(5, 3); // Exemple math-wasm
```

## Structure Standard

Chaque module suit cette architecture :

```
module-name/
├── main.go         # Code source Go
├── module.json     # Métadonnées et documentation
├── build.sh        # Script de build
├── main.wasm       # Binaire WebAssembly
└── go.mod          # Dépendances
```

### Fonctions Obligatoires

Tous les modules implémentent :
- `getAvailableFunctions()` → Array<string>
- `setSilentMode(boolean)` → boolean

### Exemple de Module

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

## Tests

```bash
# Valider tous les modules
./test-functions.sh
```

Vérifie que chaque module implémente correctement les fonctions standard.

---

**Total size:** 16MB → 4.1MB (73% compression)