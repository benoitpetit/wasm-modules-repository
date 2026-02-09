//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"
)

var silentMode = false

// Database represents a simple key-value database stored in localStorage
type Database struct {
	name string
	data map[string]interface{}
}

// Global database instances
var databases = make(map[string]*Database)

// setSilentMode enables/disables silent mode for console logs
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// createDatabase creates a new database or opens an existing one
func createDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	// Check if database already exists
	if _, exists := databases[dbName]; exists {
		if !silentMode {
			fmt.Printf("Go WASM DB: Database '%s' already exists, opening existing database\n", dbName)
		}
		return successResponse("Database opened", map[string]interface{}{
			"name":   dbName,
			"exists": true,
		})
	}

	// Create new database
	db := &Database{
		name: dbName,
		data: make(map[string]interface{}),
	}

	// Load from localStorage if exists
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		storageKey := "gowm_db_" + dbName
		stored := localStorage.Call("getItem", storageKey)
		if !stored.IsNull() {
			var loadedData map[string]interface{}
			if err := json.Unmarshal([]byte(stored.String()), &loadedData); err == nil {
				db.data = loadedData
				if !silentMode {
					fmt.Printf("Go WASM DB: Loaded database '%s' from storage with %d entries\n", dbName, len(loadedData))
				}
			}
		}
	}

	databases[dbName] = db

	if !silentMode {
		fmt.Printf("Go WASM DB: Created/opened database '%s'\n", dbName)
	}

	return successResponse("Database created successfully", map[string]interface{}{
		"name":    dbName,
		"entries": len(db.data),
	})
}

// insertData inserts or updates a key-value pair in the database
func insertData(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return errorResponse("Error: database name, key, and value required")
	}

	dbName := args[0].String()
	key := args[1].String()
	value := args[2]

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found. Create it first with createDatabase()")
	}

	// Convert JS value to Go interface{}
	var goValue interface{}
	switch value.Type() {
	case js.TypeString:
		goValue = value.String()
	case js.TypeNumber:
		goValue = value.Float()
	case js.TypeBoolean:
		goValue = value.Bool()
	case js.TypeObject:
		// Try to parse as JSON
		jsonStr := js.Global().Get("JSON").Call("stringify", value).String()
		if err := json.Unmarshal([]byte(jsonStr), &goValue); err != nil {
			goValue = jsonStr
		}
	default:
		goValue = value.String()
	}

	db.data[key] = goValue

	// Save to localStorage
	if err := saveDatabase(db); err != nil {
		return errorResponse("Error: failed to save database - " + err.Error())
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Inserted key '%s' in database '%s'\n", key, dbName)
	}

	return successResponse("Data inserted successfully", map[string]interface{}{
		"key":   key,
		"value": goValue,
	})
}

// getData retrieves a value by key from the database
func getData(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return errorResponse("Error: database name and key required")
	}

	dbName := args[0].String()
	key := args[1].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	value, exists := db.data[key]
	if !exists {
		return errorResponse("Error: key not found")
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Retrieved key '%s' from database '%s'\n", key, dbName)
	}

	return successResponse("Data retrieved successfully", map[string]interface{}{
		"key":   key,
		"value": value,
	})
}

// deleteData removes a key-value pair from the database
func deleteData(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return errorResponse("Error: database name and key required")
	}

	dbName := args[0].String()
	key := args[1].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	if _, exists := db.data[key]; !exists {
		return errorResponse("Error: key not found")
	}

	delete(db.data, key)

	// Save to localStorage
	if err := saveDatabase(db); err != nil {
		return errorResponse("Error: failed to save database - " + err.Error())
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Deleted key '%s' from database '%s'\n", key, dbName)
	}

	return successResponse("Data deleted successfully", map[string]interface{}{
		"key": key,
	})
}

// listKeys returns all keys in the database
func listKeys(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	keys := make([]string, 0, len(db.data))
	for k := range db.data {
		keys = append(keys, k)
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Listed %d keys from database '%s'\n", len(keys), dbName)
	}

	return successResponse("Keys retrieved successfully", map[string]interface{}{
		"keys":  keys,
		"count": len(keys),
	})
}

// getAllData returns all data in the database
func getAllData(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Retrieved all data from database '%s' (%d entries)\n", dbName, len(db.data))
	}

	return successResponse("All data retrieved successfully", map[string]interface{}{
		"data":  db.data,
		"count": len(db.data),
	})
}

// clearDatabase removes all data from the database
func clearDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	previousCount := len(db.data)
	db.data = make(map[string]interface{})

	// Save to localStorage
	if err := saveDatabase(db); err != nil {
		return errorResponse("Error: failed to save database - " + err.Error())
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Cleared database '%s', removed %d entries\n", dbName, previousCount)
	}

	return successResponse("Database cleared successfully", map[string]interface{}{
		"deletedCount": previousCount,
	})
}

// deleteDatabase removes the database entirely
func deleteDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	_, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	// Remove from localStorage
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		storageKey := "gowm_db_" + dbName
		localStorage.Call("removeItem", storageKey)
	}

	delete(databases, dbName)

	if !silentMode {
		fmt.Printf("Go WASM DB: Deleted database '%s'\n", dbName)
	}

	return successResponse("Database deleted successfully", map[string]interface{}{
		"name": dbName,
	})
}

// exportDatabase exports the database as JSON
func exportDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	jsonData, err := json.Marshal(db.data)
	if err != nil {
		return errorResponse("Error: failed to export database - " + err.Error())
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Exported database '%s'\n", dbName)
	}

	return successResponse("Database exported successfully", map[string]interface{}{
		"name": dbName,
		"data": string(jsonData),
	})
}

// importDatabase imports data from JSON into the database
func importDatabase(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return errorResponse("Error: database name and JSON data required")
	}

	dbName := args[0].String()
	jsonData := args[1].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found. Create it first with createDatabase()")
	}

	var importedData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &importedData); err != nil {
		return errorResponse("Error: invalid JSON data - " + err.Error())
	}

	// Merge or replace data
	replace := false
	if len(args) > 2 {
		replace = args[2].Bool()
	}

	if replace {
		db.data = importedData
	} else {
		for k, v := range importedData {
			db.data[k] = v
		}
	}

	// Save to localStorage
	if err := saveDatabase(db); err != nil {
		return errorResponse("Error: failed to save database - " + err.Error())
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Imported %d entries into database '%s'\n", len(importedData), dbName)
	}

	return successResponse("Database imported successfully", map[string]interface{}{
		"name":          dbName,
		"importedCount": len(importedData),
		"totalCount":    len(db.data),
	})
}

// getStats returns statistics about the database
func getStats(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("Error: database name required")
	}

	dbName := args[0].String()

	db, exists := databases[dbName]
	if !exists {
		return errorResponse("Error: database not found")
	}

	// Calculate storage size
	jsonData, _ := json.Marshal(db.data)
	storageSize := len(jsonData)

	stats := map[string]interface{}{
		"name":          dbName,
		"entries":       len(db.data),
		"storageSize":   storageSize,
		"storageSizeKB": float64(storageSize) / 1024,
	}

	if !silentMode {
		fmt.Printf("Go WASM DB: Retrieved stats for database '%s'\n", dbName)
	}

	return successResponse("Stats retrieved successfully", stats)
}

// Helper functions
func saveDatabase(db *Database) error {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return fmt.Errorf("localStorage not available")
	}

	jsonData, err := json.Marshal(db.data)
	if err != nil {
		return err
	}

	storageKey := "gowm_db_" + db.name
	localStorage.Call("setItem", storageKey, string(jsonData))

	return nil
}

func successResponse(message string, data map[string]interface{}) js.Value {
	response := map[string]interface{}{
		"success":   true,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}
	return mapToJSValue(response)
}

func errorResponse(message string) js.Value {
	response := map[string]interface{}{
		"success":   false,
		"error":     message,
		"timestamp": time.Now().Unix(),
	}
	return mapToJSValue(response)
}

func mapToJSValue(m map[string]interface{}) js.Value {
	obj := js.Global().Get("Object").New()
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			obj.Set(k, mapToJSValue(val))
		case []interface{}:
			arr := js.Global().Get("Array").New(len(val))
			for i, item := range val {
				arr.SetIndex(i, item)
			}
			obj.Set(k, arr)
		case []string:
			arr := js.Global().Get("Array").New(len(val))
			for i, item := range val {
				arr.SetIndex(i, item)
			}
			obj.Set(k, arr)
		default:
			obj.Set(k, v)
		}
	}
	return obj
}

// getAvailableFunctions returns the list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		"createDatabase", "deleteDatabase", "clearDatabase",
		"insertData", "getData", "deleteData",
		"getAllData", "listKeys",
		"exportDatabase", "importDatabase",
		"getStats",
		"getAvailableFunctions", "setSilentMode",
	}

	// Convert to JS Array
	arr := js.Global().Get("Array").New(len(functions))
	for i, fn := range functions {
		arr.SetIndex(i, fn)
	}
	return arr
}

func main() {
	fmt.Println("🚀 DB-WASM v0.1.0 starting...")

	// Register database management functions
	js.Global().Set("createDatabase", js.FuncOf(createDatabase))
	js.Global().Set("deleteDatabase", js.FuncOf(deleteDatabase))
	js.Global().Set("clearDatabase", js.FuncOf(clearDatabase))

	// Register CRUD operations
	js.Global().Set("insertData", js.FuncOf(insertData))
	js.Global().Set("getData", js.FuncOf(getData))
	js.Global().Set("deleteData", js.FuncOf(deleteData))
	js.Global().Set("getAllData", js.FuncOf(getAllData))
	js.Global().Set("listKeys", js.FuncOf(listKeys))

	// Register export/import functions
	js.Global().Set("exportDatabase", js.FuncOf(exportDatabase))
	js.Global().Set("importDatabase", js.FuncOf(importDatabase))

	// Register utility functions
	js.Global().Set("getStats", js.FuncOf(getStats))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	// Signal readiness for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("✅ DB-WASM ready!")
	fmt.Println("Available functions: createDatabase, insertData, getData, deleteData, getAllData, listKeys, clearDatabase, deleteDatabase, exportDatabase, importDatabase, getStats, setSilentMode")

	// Keep the program alive
	select {}
}
