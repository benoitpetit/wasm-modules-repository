//go:build js && wasm

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/antchfx/xmlquery"
	"gopkg.in/yaml.v3"
)

var silentMode = false

// JSONResult represents a JSON operation result
type JSONResult struct {
	Data     interface{} `json:"data"`
	Valid    bool        `json:"valid"`
	Size     int         `json:"size"`
	Format   string      `json:"format"`
	Minified bool        `json:"minified,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// XMLResult represents an XML operation result
type XMLResult struct {
	Data     interface{} `json:"data"`
	Valid    bool        `json:"valid"`
	Size     int         `json:"size"`
	Format   string      `json:"format"`
	Root     string      `json:"root,omitempty"`
	Encoding string      `json:"encoding,omitempty"`
	Error    string      `json:"error,omitempty"`
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

// setSilentMode - Set silent mode for operations
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// parseJSON - Parse JSON string and validate
func parseJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: "parseJSON requires exactly 1 argument (jsonString)",
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
			Error:  fmt.Sprintf("Invalid JSON: %v", err),
		})
	}

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

// stringifyJSON - Convert object to JSON string
func stringifyJSON(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(JSONResult{
			Error: "stringifyJSON requires at least 1 argument (data)",
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
			Error: fmt.Sprintf("Failed to stringify JSON: %v", err),
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
			Error: "minifyJSON requires exactly 1 argument (jsonString)",
		})
	}

	jsonString := args[0].String()

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid JSON: %v", err),
			Format: "json",
		})
	}

	minifiedBytes, err := json.Marshal(data)
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: fmt.Sprintf("Failed to minify JSON: %v", err),
		})
	}

	minified := string(minifiedBytes)
	originalSize := len(jsonString)
	newSize := len(minified)

	if !silentMode {
		reduction := float64(originalSize-newSize) / float64(originalSize) * 100
		fmt.Printf("JSON WASM: Minified JSON - %d → %d bytes (%.1f%% reduction)\n",
			originalSize, newSize, reduction)
	}

	return js.ValueOf(JSONResult{
		Data:     minified,
		Valid:    true,
		Size:     newSize,
		Format:   "json",
		Minified: true,
	})
}

// parseXML - Parse XML string and validate
func parseXML(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(XMLResult{
			Error: "parseXML requires exactly 1 argument (xmlString)",
		})
	}

	xmlString := args[0].String()

	// Parse with xmlquery for better handling
	doc, err := xmlquery.Parse(strings.NewReader(xmlString))
	if err != nil {
		return js.ValueOf(XMLResult{
			Valid:  false,
			Size:   len(xmlString),
			Format: "xml",
			Error:  fmt.Sprintf("Invalid XML: %v", err),
		})
	}

	// Convert to map for JS consumption
	data := xmlNodeToMap(doc)

	rootElement := ""
	if doc.FirstChild != nil {
		rootElement = doc.FirstChild.Data
	}

	if !silentMode {
		fmt.Printf("XML WASM: Successfully parsed XML (%d bytes)\n", len(xmlString))
	}

	return js.ValueOf(XMLResult{
		Data:     data,
		Valid:    true,
		Size:     len(xmlString),
		Format:   "xml",
		Root:     rootElement,
		Encoding: "UTF-8",
	})
}

// xmlToJSON - Convert XML to JSON
func xmlToJSON(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(JSONResult{
			Error: "xmlToJSON requires exactly 1 argument (xmlString)",
		})
	}

	xmlString := args[0].String()

	doc, err := xmlquery.Parse(strings.NewReader(xmlString))
	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid XML: %v", err),
			Format: "json",
		})
	}

	// Convert XML to map structure
	data := xmlNodeToMap(doc)

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: fmt.Sprintf("Failed to convert to JSON: %v", err),
		})
	}

	jsonString := string(jsonBytes)

	if !silentMode {
		fmt.Printf("XML WASM: Converted XML to JSON (%d → %d bytes)\n",
			len(xmlString), len(jsonString))
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
			Error: "jsonToXML requires at least 1 argument (jsonString)",
		})
	}

	jsonString := args[0].String()
	rootElement := "root"

	if len(args) > 1 {
		rootElement = args[1].String()
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return js.ValueOf(XMLResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid JSON: %v", err),
			Format: "xml",
		})
	}

	// Convert to XML
	xmlString := mapToXML(data, rootElement, 0)
	xmlString = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + xmlString

	if !silentMode {
		fmt.Printf("XML WASM: Converted JSON to XML (%d → %d bytes)\n",
			len(jsonString), len(xmlString))
	}

	return js.ValueOf(XMLResult{
		Data:     xmlString,
		Valid:    true,
		Size:     len(xmlString),
		Format:   "xml",
		Root:     rootElement,
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
			Error: "csvToJSON requires exactly 1 argument (csvString)",
		})
	}

	csvString := args[0].String()

	reader := csv.NewReader(strings.NewReader(csvString))
	records, err := reader.ReadAll()

	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid CSV: %v", err),
			Format: "json",
		})
	}

	if len(records) == 0 {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Error:  "Empty CSV data",
			Format: "json",
		})
	}

	// Use first row as headers
	headers := records[0]
	var jsonData []map[string]interface{}

	for i := 1; i < len(records); i++ {
		row := make(map[string]interface{})
		for j, value := range records[i] {
			if j < len(headers) {
				// Try to convert numbers
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					row[headers[j]] = num
				} else if value == "true" || value == "false" {
					row[headers[j]] = value == "true"
				} else {
					row[headers[j]] = value
				}
			}
		}
		jsonData = append(jsonData, row)
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: fmt.Sprintf("Failed to convert to JSON: %v", err),
		})
	}

	jsonString := string(jsonBytes)

	if !silentMode {
		fmt.Printf("CSV WASM: Converted CSV to JSON (%d rows → %d bytes)\n",
			len(records)-1, len(jsonString))
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
			Error: "yamlToJSON requires exactly 1 argument (yamlString)",
		})
	}

	yamlString := args[0].String()

	var data interface{}
	err := yaml.Unmarshal([]byte(yamlString), &data)
	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid YAML: %v", err),
			Format: "json",
		})
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: fmt.Sprintf("Failed to convert to JSON: %v", err),
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

// extractJSONPath - Extract value using JSON path
func extractJSONPath(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(JSONResult{
			Error: "extractJSONPath requires exactly 2 arguments (jsonString, path)",
		})
	}

	jsonString := args[0].String()
	path := args[1].String()

	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return js.ValueOf(JSONResult{
			Valid:  false,
			Error:  fmt.Sprintf("Invalid JSON: %v", err),
			Format: "json",
		})
	}

	// Simple path extraction (supports basic dot notation)
	result := extractByPath(data, path)

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return js.ValueOf(JSONResult{
			Error: fmt.Sprintf("Failed to serialize result: %v", err),
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
	indentStr := strings.Repeat("  ", indent)

	switch v := data.(type) {
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

	default:
		return fmt.Sprintf("%s<%s>%v</%s>\n", indentStr, tagName, v, tagName)
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
