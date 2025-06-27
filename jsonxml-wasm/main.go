//go:build js && wasm

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall/js"

	"github.com/antchfx/xmlquery"
	"gopkg.in/yaml.v3"
)

var silentMode = false
var memoryPool sync.Pool

// StructuredError represents a detailed error response
type StructuredError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// JSONResult represents a JSON operation result
type JSONResult struct {
	Data     interface{}      `json:"data"`
	Valid    bool             `json:"valid"`
	Size     int              `json:"size"`
	Format   string           `json:"format"`
	Minified bool             `json:"minified,omitempty"`
	Error    *StructuredError `json:"error,omitempty"`
}

// XMLResult represents an XML operation result
type XMLResult struct {
	Data     interface{}      `json:"data"`
	Valid    bool             `json:"valid"`
	Size     int              `json:"size"`
	Format   string           `json:"format"`
	Root     string           `json:"root,omitempty"`
	Encoding string           `json:"encoding,omitempty"`
	Error    *StructuredError `json:"error,omitempty"`
}

// CSVResult represents a CSV operation result
type CSVResult struct {
	Data    interface{} `json:"data"`
	Rows    int         `json:"rows"`
	Columns int         `json:"columns"`
	Format  string      `json:"format"`
	Error   string      `json:"error,omitempty"`
}

// YAMLResult represents a YAML operation result
type YAMLResult struct {
	Data   interface{} `json:"data"`
	Valid  bool        `json:"valid"`
	Size   int         `json:"size"`
	Format string      `json:"format"`
	Error  string      `json:"error,omitempty"`
}

// ValidationResult represents validation result
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Format   string   `json:"format"`
}

func init() {
	// Initialize memory pool
	memoryPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 64*1024) // 64KB initial capacity
		},
	}
}

// cleanupMemory forces a garbage collection
func cleanupMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

// getBuffer gets a buffer from the pool
func getBuffer() []byte {
	return memoryPool.Get().([]byte)
}

// putBuffer returns a buffer to the pool
func putBuffer(buf []byte) {
	memoryPool.Put(buf[:0])
}

// setSilentMode - Set silent mode for operations
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// parseJSON - Parse JSON string and validate with memory management
func parseJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "parseJSON requires exactly 1 argument (jsonString)",
				Details: map[string]interface{}{
					"expected": 1,
					"received": len(args),
				},
			},
		})
	}

	jsonString := args[0].String()
	buf := getBuffer()
	defer putBuffer(buf)

	// Check memory limit
	if len(jsonString) > 128*1024*1024 { // 128MB limit
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "MEMORY_LIMIT",
				Message: "JSON string exceeds memory limit",
				Details: map[string]interface{}{
					"limit": "128MB",
					"size":  fmt.Sprintf("%.2fMB", float64(len(jsonString))/(1024*1024)),
				},
			},
		})
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Size:   len(jsonString),
			Format: "json",
			Error: &StructuredError{
				Code:    "INVALID_JSON",
				Message: fmt.Sprintf("Invalid JSON: %v", err),
				Details: map[string]interface{}{
					"position": getErrorPosition(err),
					"context":  getErrorContext(jsonString, err),
				},
			},
		})
	}

	defer cleanupMemory()

	if !silentMode {
		fmt.Printf("JSON WASM: Successfully parsed JSON (%d bytes)\n", len(jsonString))
	}

	return js.ValueOf(JSONResult{
		Data:   data,
		Valid:  true,
		Size:   len(jsonString),
		Format: "json",
	})
}

// getErrorPosition extracts position from json.SyntaxError
func getErrorPosition(err error) int {
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		return int(syntaxErr.Offset)
	}
	return 0
}

// getErrorContext returns the context around the error position
func getErrorContext(input string, err error) string {
	pos := getErrorPosition(err)
	if pos == 0 {
		return ""
	}

	start := pos - 10
	if start < 0 {
		start = 0
	}
	end := pos + 10
	if end > len(input) {
		end = len(input)
	}

	return input[start:end]
}

// stringifyJSON - Convert object to JSON string
func stringifyJSON(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "stringifyJSON requires at least 1 argument (data)",
				Details: map[string]interface{}{
					"expected": "at least 1",
					"received": len(args),
				},
			},
		})
	}

	// Parse JS value to Go interface
	data := parseJSValue(args[0])

	// Optional formatting (pretty print)
	pretty := false
	if len(args) > 1 {
		pretty = args[1].Bool()
	}

	var jsonBytes []byte
	var err error

	if pretty {
		jsonBytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(data)
	}

	if err != nil {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "STRINGIFY_ERROR",
				Message: fmt.Sprintf("Failed to stringify JSON: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", data),
				},
			},
		})
	}

	jsonString := string(jsonBytes)

	if !silentMode {
		fmt.Printf("JSON WASM: Generated JSON string (%d bytes, pretty: %v)\n", len(jsonString), pretty)
	}

	return js.ValueOf(JSONResult{
		Data:   jsonString,
		Valid:  true,
		Size:   len(jsonString),
		Format: "json",
	})
}

// validateJSON - Validate JSON syntax
func validateJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(ValidationResult{
			Valid:  false,
			Errors: []string{"validateJSON requires exactly 1 argument (jsonString)"},
			Format: "json",
		})
	}

	jsonString := args[0].String()

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)

	result := ValidationResult{
		Format: "json",
	}

	if err != nil {
		result.Valid = false
		result.Errors = []string{err.Error()}
	} else {
		result.Valid = true
	}

	if !silentMode {
		fmt.Printf("JSON WASM: JSON validation result: %v\n", result.Valid)
	}

	return js.ValueOf(result)
}

// minifyJSON - Minify JSON by removing whitespace
func minifyJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "minifyJSON requires exactly 1 argument (jsonString)",
				Details: map[string]interface{}{
					"expected": 1,
					"received": len(args),
				},
			},
		})
	}

	jsonString := args[0].String()

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Size:   len(jsonString),
			Format: "json",
			Error: &StructuredError{
				Code:    "INVALID_JSON",
				Message: fmt.Sprintf("Invalid JSON: %v", err),
				Details: map[string]interface{}{
					"error":    err.Error(),
					"position": getErrorPosition(err),
					"context":  getErrorContext(jsonString, err),
				},
			},
		})
	}

	minified, err := json.Marshal(data)
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "MINIFY_ERROR",
				Message: fmt.Sprintf("Failed to minify JSON: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", data),
				},
			},
		})
	}

	minifiedString := string(minified)

	if !silentMode {
		fmt.Printf("JSON WASM: Minified JSON (%d bytes)\n", len(minifiedString))
	}

	return js.ValueOf(JSONResult{
		Data:     minifiedString,
		Valid:    true,
		Size:     len(minifiedString),
		Format:   "json",
		Minified: true,
	})
}

// parseXML - Parse XML string and validate
func parseXML(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(XMLResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "parseXML requires exactly 1 argument (xmlString)",
				Details: map[string]interface{}{
					"expected": 1,
					"received": len(args),
				},
			},
		})
	}

	xmlString := args[0].String()
	doc, err := xmlquery.Parse(strings.NewReader(xmlString))

	if err != nil {
		return js.ValueOf(XMLResult{
			Valid:  false,
			Size:   len(xmlString),
			Format: "xml",
			Error: &StructuredError{
				Code:    "INVALID_XML",
				Message: fmt.Sprintf("Invalid XML: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
		})
	}

	rootElement := doc.SelectElement("/*")
	rootName := ""
	if rootElement != nil {
		rootName = rootElement.Data
	}

	if !silentMode {
		fmt.Printf("XML WASM: Successfully parsed XML (%d bytes)\n", len(xmlString))
	}

	return js.ValueOf(XMLResult{
		Data:     xmlNodeToMap(doc),
		Valid:    true,
		Size:     len(xmlString),
		Format:   "xml",
		Root:     rootName,
		Encoding: "UTF-8",
	})
}

// xmlToJSON - Convert XML to JSON
func xmlToJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "xmlToJSON requires exactly 1 argument (xmlString)",
				Details: map[string]interface{}{
					"expected": 1,
					"received": len(args),
				},
			},
		})
	}

	xmlString := args[0].String()
	doc, err := xmlquery.Parse(strings.NewReader(xmlString))

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Size:   len(xmlString),
			Format: "json",
			Error: &StructuredError{
				Code:    "INVALID_XML",
				Message: fmt.Sprintf("Invalid XML: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
		})
	}

	data := xmlNodeToMap(doc)
	jsonBytes, err := json.Marshal(data)

	if err != nil {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "CONVERSION_ERROR",
				Message: fmt.Sprintf("Failed to convert to JSON: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", data),
				},
			},
		})
	}

	jsonString := string(jsonBytes)

	if !silentMode {
		fmt.Printf("XML WASM: Successfully converted XML to JSON (%d bytes)\n", len(jsonString))
	}

	return js.ValueOf(JSONResult{
		Data:   jsonString,
		Valid:  true,
		Size:   len(jsonString),
		Format: "json",
	})
}

// jsonToXML - Convert JSON to XML
func jsonToXML(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(XMLResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "jsonToXML requires at least 1 argument (jsonString)",
				Details: map[string]interface{}{
					"expected": "at least 1",
					"received": len(args),
				},
			},
		})
	}

	jsonString := args[0].String()
	rootName := "root"
	if len(args) > 1 && args[1].Type() == js.TypeString && args[1].String() != "" {
		rootName = args[1].String()
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return js.ValueOf(XMLResult{
			Valid:  false,
			Size:   len(jsonString),
			Format: "xml",
			Error: &StructuredError{
				Code:    "INVALID_JSON",
				Message: fmt.Sprintf("Invalid JSON: %v", err),
				Details: map[string]interface{}{
					"error":    err.Error(),
					"position": getErrorPosition(err),
					"context":  getErrorContext(jsonString, err),
				},
			},
		})
	}

	// Only allow object or array at root
	switch data.(type) {
	case map[string]interface{}, []interface{}:
		// ok
	default:
		return js.ValueOf(XMLResult{
			Valid:  false,
			Size:   len(jsonString),
			Format: "xml",
			Error: &StructuredError{
				Code:    "UNSUPPORTED_TYPE",
				Message: "Root JSON value must be an object or array for XML conversion",
				Details: map[string]interface{}{
					"type": fmt.Sprintf("%T", data),
				},
			},
		})
	}

	xmlString := mapToXML(data, rootName, 0)
	xmlString = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + xmlString

	if xmlString == "" {
		return js.ValueOf(XMLResult{
			Valid:  false,
			Size:   len(jsonString),
			Format: "xml",
			Error: &StructuredError{
				Code:    "CONVERSION_ERROR",
				Message: "Failed to convert JSON to XML: empty result",
			},
		})
	}

	if !silentMode {
		fmt.Printf("XML WASM: Converted JSON to XML (%d → %d bytes)\n",
			len(jsonString), len(xmlString))
	}

	return js.ValueOf(XMLResult{
		Data:     xmlString,
		Valid:    true,
		Size:     len(xmlString),
		Format:   "xml",
		Root:     rootName,
		Encoding: "UTF-8",
	})
}

// validateXML - Validate XML syntax
func validateXML(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(ValidationResult{
			Valid:  false,
			Errors: []string{"validateXML requires exactly 1 argument (xmlString)"},
			Format: "xml",
		})
	}

	xmlString := args[0].String()

	_, err := xmlquery.Parse(strings.NewReader(xmlString))

	result := ValidationResult{
		Format: "xml",
	}

	if err != nil {
		result.Valid = false
		result.Errors = []string{err.Error()}
	} else {
		result.Valid = true
	}

	if !silentMode {
		fmt.Printf("XML WASM: XML validation result: %v\n", result.Valid)
	}

	return js.ValueOf(result)
}

// csvToJSON - Convert CSV to JSON
func csvToJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "csvToJSON requires exactly 1 argument (csvString)",
				Details: map[string]interface{}{
					"expected": 1,
					"received": len(args),
				},
			},
		})
	}

	csvString := args[0].String()
	reader := csv.NewReader(strings.NewReader(csvString))
	records, err := reader.ReadAll()

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Size:   len(csvString),
			Format: "json",
			Error: &StructuredError{
				Code:    "INVALID_CSV",
				Message: fmt.Sprintf("Invalid CSV: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
		})
	}

	if len(records) == 0 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "EMPTY_CSV",
				Message: "Empty CSV data",
				Details: map[string]interface{}{
					"size": len(csvString),
				},
			},
		})
	}

	// Process CSV data
	headers := records[0]
	var result []map[string]interface{}

	for _, record := range records[1:] {
		row := make(map[string]interface{})
		for i, value := range record {
			if i < len(headers) {
				// Try to convert to appropriate type
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					row[headers[i]] = v
				} else if v, err := strconv.ParseBool(value); err == nil {
					row[headers[i]] = v
				} else {
					row[headers[i]] = value
				}
			}
		}
		result = append(result, row)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "CONVERSION_ERROR",
				Message: fmt.Sprintf("Failed to convert to JSON: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", result),
				},
			},
		})
	}

	jsonString := string(jsonBytes)

	if !silentMode {
		fmt.Printf("CSV WASM: Successfully converted CSV to JSON (%d bytes)\n", len(jsonString))
	}

	return js.ValueOf(JSONResult{
		Data:   jsonString,
		Valid:  true,
		Size:   len(jsonString),
		Format: "json",
	})
}

// jsonToCSV - Convert JSON to CSV
func jsonToCSV(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(CSVResult{
			Error: "jsonToCSV requires exactly 1 argument (jsonString)",
		})
	}

	jsonString := args[0].String()

	var data []map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return js.ValueOf(CSVResult{
			Error:  fmt.Sprintf("Invalid JSON: %v", err),
			Format: "csv",
		})
	}

	if len(data) == 0 {
		return js.ValueOf(CSVResult{
			Error:  "Empty JSON array",
			Format: "csv",
		})
	}

	// Extract headers from first object
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	writer.Write(headers)

	// Write data rows
	for _, row := range data {
		var record []string
		for _, header := range headers {
			if value, exists := row[header]; exists {
				record = append(record, fmt.Sprintf("%v", value))
			} else {
				record = append(record, "")
			}
		}
		writer.Write(record)
	}

	writer.Flush()
	csvString := buf.String()

	if !silentMode {
		fmt.Printf("CSV WASM: Converted JSON to CSV (%d rows, %d columns)\n",
			len(data), len(headers))
	}

	return js.ValueOf(CSVResult{
		Data:    csvString,
		Rows:    len(data),
		Columns: len(headers),
		Format:  "csv",
	})
}

// yamlToJSON - Convert YAML to JSON
func yamlToJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "yamlToJSON requires exactly 1 argument (yamlString)",
				Details: map[string]interface{}{
					"expected": 1,
					"received": len(args),
				},
			},
		})
	}

	yamlString := args[0].String()
	var data interface{}
	err := yaml.Unmarshal([]byte(yamlString), &data)

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Size:   len(yamlString),
			Format: "json",
			Error: &StructuredError{
				Code:    "INVALID_YAML",
				Message: fmt.Sprintf("Invalid YAML: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
		})
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "CONVERSION_ERROR",
				Message: fmt.Sprintf("Failed to convert to JSON: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", data),
				},
			},
		})
	}

	jsonString := string(jsonBytes)

	if !silentMode {
		fmt.Printf("YAML WASM: Converted YAML to JSON (%d → %d bytes)\n",
			len(yamlString), len(jsonString))
	}

	return js.ValueOf(JSONResult{
		Data:   jsonString,
		Valid:  true,
		Size:   len(jsonString),
		Format: "json",
	})
}

// jsonToYAML - Convert JSON to YAML
func jsonToYAML(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(YAMLResult{
			Error: "jsonToYAML requires exactly 1 argument (jsonString)",
		})
	}

	jsonString := args[0].String()

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return js.ValueOf(YAMLResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid JSON: %v", err),
			Format: "yaml",
		})
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return js.ValueOf(YAMLResult{
			Error: fmt.Sprintf("Failed to convert to YAML: %v", err),
		})
	}

	yamlString := string(yamlBytes)

	if !silentMode {
		fmt.Printf("YAML WASM: Converted JSON to YAML (%d → %d bytes)\n",
			len(jsonString), len(yamlString))
	}

	return js.ValueOf(YAMLResult{
		Data:   yamlString,
		Valid:  true,
		Size:   len(yamlString),
		Format: "yaml",
	})
}

// extractJSONPath - Extract value from JSON using path
func extractJSONPath(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "INVALID_ARGS",
				Message: "extractJSONPath requires exactly 2 arguments (jsonString, path)",
				Details: map[string]interface{}{
					"expected": 2,
					"received": len(args),
				},
			},
		})
	}

	jsonString := args[0].String()
	path := args[1].String()

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Size:   len(jsonString),
			Format: "json",
			Error: &StructuredError{
				Code:    "INVALID_JSON",
				Message: fmt.Sprintf("Invalid JSON: %v", err),
				Details: map[string]interface{}{
					"error":    err.Error(),
					"position": getErrorPosition(err),
					"context":  getErrorContext(jsonString, err),
				},
			},
		})
	}

	result := extractByPath(data, path)
	resultBytes, err := json.Marshal(result)

	if err != nil {
		return js.ValueOf(JSONResult{
			Error: &StructuredError{
				Code:    "SERIALIZATION_ERROR",
				Message: fmt.Sprintf("Failed to serialize result: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", result),
				},
			},
		})
	}

	resultString := string(resultBytes)

	if !silentMode {
		fmt.Printf("JSON WASM: Extracted JSON path '%s'\n", path)
	}

	return js.ValueOf(JSONResult{
		Data:   resultString,
		Valid:  true,
		Size:   len(resultString),
		Format: "json",
	})
}

// validateJSONSchema - Basic JSON schema validation
func validateJSONSchema(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(ValidationResult{
			Valid:  false,
			Errors: []string{"validateJSONSchema requires exactly 2 arguments (jsonString, schemaString)"},
			Format: "json",
		})
	}

	jsonString := args[0].String()
	schemaString := args[1].String()

	var data interface{}
	var schema interface{}

	// Validate JSON data
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return js.ValueOf(ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Invalid JSON data: %v", err)},
			Format: "json",
		})
	}

	// Validate schema
	err = json.Unmarshal([]byte(schemaString), &schema)
	if err != nil {
		return js.ValueOf(ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Invalid JSON schema: %v", err)},
			Format: "json",
		})
	}

	// Basic validation (simplified)
	errors := performBasicSchemaValidation(data, schema)

	result := ValidationResult{
		Valid:  len(errors) == 0,
		Format: "json",
	}

	if len(errors) > 0 {
		result.Errors = errors
	}

	if !silentMode {
		fmt.Printf("JSON WASM: Schema validation result: %v (%d errors)\n",
			result.Valid, len(errors))
	}

	return js.ValueOf(result)
}

// getAvailableFunctions - Return list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []interface{}{
		"parseJSON",
		"stringifyJSON",
		"validateJSON",
		"minifyJSON",
		"parseXML",
		"xmlToJSON",
		"jsonToXML",
		"validateXML",
		"csvToJSON",
		"jsonToCSV",
		"yamlToJSON",
		"jsonToYAML",
		"extractJSONPath",
		"validateJSONSchema",
		"getAvailableFunctions",
		"setSilentMode",
	}
	return js.ValueOf(functions)
}

// Helper functions

func parseJSValue(value js.Value) interface{} {
	switch value.Type() {
	case js.TypeBoolean:
		return value.Bool()
	case js.TypeNumber:
		return value.Float()
	case js.TypeString:
		return value.String()
	case js.TypeObject:
		if value.Get("constructor").Get("name").String() == "Array" {
			length := value.Get("length").Int()
			result := make([]interface{}, length)
			for i := 0; i < length; i++ {
				result[i] = parseJSValue(value.Index(i))
			}
			return result
		} else {
			result := make(map[string]interface{})
			keys := js.Global().Get("Object").Call("keys", value)
			for i := 0; i < keys.Get("length").Int(); i++ {
				key := keys.Index(i).String()
				result[key] = parseJSValue(value.Get(key))
			}
			return result
		}
	case js.TypeNull:
		return nil
	default:
		return value.String()
	}
}

func xmlNodeToMap(node *xmlquery.Node) interface{} {
	if node == nil {
		return nil
	}

	switch node.Type {
	case xmlquery.TextNode:
		return strings.TrimSpace(node.Data)
	case xmlquery.ElementNode:
		result := make(map[string]interface{})

		// Add attributes
		if len(node.Attr) > 0 {
			attrs := make(map[string]string)
			for _, attr := range node.Attr {
				attrs[attr.Name.Local] = attr.Value
			}
			result["@attributes"] = attrs
		}

		// Process children
		var children []interface{}
		var textContent string

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			switch child.Type {
			case xmlquery.TextNode:
				text := strings.TrimSpace(child.Data)
				if text != "" {
					textContent += text
				}
			case xmlquery.ElementNode:
				children = append(children, map[string]interface{}{
					child.Data: xmlNodeToMap(child),
				})
			}
		}

		if textContent != "" {
			result["#text"] = textContent
		}

		if len(children) > 0 {
			result["children"] = children
		}

		return result
	}

	return nil
}

func mapToXML(data interface{}, tagName string, indent int) string {
	// Sanitize tag name to prevent XML errors
	if tagName == "" {
		tagName = "item"
	}
	// Remove invalid XML tag characters
	tagName = strings.ReplaceAll(tagName, " ", "_")
	tagName = strings.ReplaceAll(tagName, "\n", "_")
	tagName = strings.ReplaceAll(tagName, "\t", "_")

	indentStr := strings.Repeat("  ", indent)

	switch v := data.(type) {
	case nil:
		return fmt.Sprintf("%s<%s></%s>\n", indentStr, tagName, tagName)
	case map[string]interface{}:
		var result strings.Builder
		result.WriteString(fmt.Sprintf("%s<%s>\n", indentStr, tagName))

		for key, value := range v {
			result.WriteString(mapToXML(value, key, indent+1))
		}

		result.WriteString(fmt.Sprintf("%s</%s>\n", indentStr, tagName))
		return result.String()

	case []interface{}:
		var result strings.Builder
		for _, item := range v {
			result.WriteString(mapToXML(item, tagName, indent))
		}
		return result.String()

	case string:
		// Escape XML special characters
		escaped := strings.ReplaceAll(v, "&", "&amp;")
		escaped = strings.ReplaceAll(escaped, "<", "&lt;")
		escaped = strings.ReplaceAll(escaped, ">", "&gt;")
		escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
		escaped = strings.ReplaceAll(escaped, "'", "&apos;")
		return fmt.Sprintf("%s<%s>%s</%s>\n", indentStr, tagName, escaped, tagName)

	case bool:
		return fmt.Sprintf("%s<%s>%t</%s>\n", indentStr, tagName, v, tagName)

	case float64:
		return fmt.Sprintf("%s<%s>%g</%s>\n", indentStr, tagName, v, tagName)

	case int:
		return fmt.Sprintf("%s<%s>%d</%s>\n", indentStr, tagName, v, tagName)

	default:
		// Convert to string safely
		value := fmt.Sprintf("%v", v)
		// Escape XML special characters
		escaped := strings.ReplaceAll(value, "&", "&amp;")
		escaped = strings.ReplaceAll(escaped, "<", "&lt;")
		escaped = strings.ReplaceAll(escaped, ">", "&gt;")
		escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
		escaped = strings.ReplaceAll(escaped, "'", "&apos;")
		return fmt.Sprintf("%s<%s>%s</%s>\n", indentStr, tagName, escaped, tagName)
	}
}

func extractByPath(data interface{}, path string) interface{} {
	if path == "" || path == "." {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			if idx, err := strconv.Atoi(part); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

func performBasicSchemaValidation(data interface{}, schema interface{}) []string {
	var errors []string

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		errors = append(errors, "Schema must be an object")
		return errors
	}

	// Check type
	if expectedType, exists := schemaMap["type"]; exists {
		actualType := getJSONType(data)
		if expectedType.(string) != actualType {
			errors = append(errors, fmt.Sprintf("Expected type %s, got %s",
				expectedType, actualType))
		}
	}

	// Check required properties for objects
	if required, exists := schemaMap["required"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if reqArray, ok := required.([]interface{}); ok {
				for _, req := range reqArray {
					if reqStr, ok := req.(string); ok {
						if _, exists := dataMap[reqStr]; !exists {
							errors = append(errors, fmt.Sprintf("Required property '%s' is missing", reqStr))
						}
					}
				}
			}
		}
	}

	return errors
}

func getJSONType(data interface{}) string {
	switch data.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64, int:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func main() {
	done := make(chan struct{})

	// Register all functions
	js.Global().Set("parseJSON", js.FuncOf(parseJSON))
	js.Global().Set("stringifyJSON", js.FuncOf(stringifyJSON))
	js.Global().Set("validateJSON", js.FuncOf(validateJSON))
	js.Global().Set("minifyJSON", js.FuncOf(minifyJSON))
	js.Global().Set("parseXML", js.FuncOf(parseXML))
	js.Global().Set("xmlToJSON", js.FuncOf(xmlToJSON))
	js.Global().Set("jsonToXML", js.FuncOf(jsonToXML))
	js.Global().Set("validateXML", js.FuncOf(validateXML))
	js.Global().Set("csvToJSON", js.FuncOf(csvToJSON))
	js.Global().Set("jsonToCSV", js.FuncOf(jsonToCSV))
	js.Global().Set("yamlToJSON", js.FuncOf(yamlToJSON))
	js.Global().Set("jsonToYAML", js.FuncOf(jsonToYAML))
	js.Global().Set("extractJSONPath", js.FuncOf(extractJSONPath))
	js.Global().Set("validateJSONSchema", js.FuncOf(validateJSONSchema))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	fmt.Println("JSONXML WASM: Module loaded successfully with comprehensive data processing capabilities")
	fmt.Println("Available functions:")
	fmt.Println("- JSON: parseJSON, stringifyJSON, validateJSON, minifyJSON")
	fmt.Println("- XML: parseXML, xmlToJSON, jsonToXML, validateXML")
	fmt.Println("- CSV: csvToJSON, jsonToCSV")
	fmt.Println("- YAML: yamlToJSON, jsonToYAML")
	fmt.Println("- Advanced: extractJSONPath, validateJSONSchema")
	fmt.Println("- Utility: getAvailableFunctions, setSilentMode")

	<-done
}
