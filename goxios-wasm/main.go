//go:build js && wasm

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall/js"
	"time"
)

var silentMode = false
var globalDefaults = RequestConfig{
	Timeout: 30000, // Default timeout of 30 seconds
	Headers: make(map[string]string),
}

var (
	requestPool sync.Pool
	clientPool  sync.Pool
)

// StructuredError represents a detailed error response
type StructuredError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// RequestConfig structure for request configuration
type RequestConfig struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Data        interface{}       `json:"data"`
	Timeout     int               `json:"timeout"`     // in milliseconds
	MaxBodySize int64             `json:"maxBodySize"` // in bytes
}

// Response structure for responses
type Response struct {
	Data    interface{}       `json:"data"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Config  RequestConfig     `json:"config"`
	Error   *StructuredError  `json:"error,omitempty"`
}

func init() {
	// Initialize request buffer pool
	requestPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 32*1024) // 32KB initial capacity
		},
	}

	// Initialize HTTP client pool
	clientPool = sync.Pool{
		New: func() interface{} {
			return &http.Client{
				Timeout: time.Duration(globalDefaults.Timeout) * time.Millisecond,
			}
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
	return requestPool.Get().([]byte)
}

// putBuffer returns a buffer to the pool
func putBuffer(buf []byte) {
	requestPool.Put(buf[:0])
}

// getClient gets a client from the pool
func getClient(timeout time.Duration) *http.Client {
	client := clientPool.Get().(*http.Client)
	client.Timeout = timeout
	return client
}

// putClient returns a client to the pool
func putClient(client *http.Client) {
	clientPool.Put(client)
}

// Fonction pour activer/désactiver le mode silencieux
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// setDefaults - Set global default configuration
func setDefaults(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "Configuration object required for setDefaults",
		})
	}

	config := parseConfig(args[0])

	// Merge with existing defaults
	if config.Timeout > 0 {
		globalDefaults.Timeout = config.Timeout
	}
	if config.Headers != nil {
		if globalDefaults.Headers == nil {
			globalDefaults.Headers = make(map[string]string)
		}
		for key, value := range config.Headers {
			globalDefaults.Headers[key] = value
		}
	}
	if config.URL != "" {
		globalDefaults.URL = config.URL
	}

	if !silentMode {
		fmt.Printf("Goxios WASM: Global defaults updated\n")
	}

	return js.ValueOf(map[string]interface{}{
		"success": true,
		"config":  convertToJSValue(globalDefaults),
	})
}

// getDefaults - Get current global default configuration
func getDefaults(this js.Value, args []js.Value) interface{} {
	return js.ValueOf(map[string]interface{}{
		"config": convertToJSValue(globalDefaults),
	})
}

// getAvailableFunctions - Get list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		"get", "post", "put", "delete", "patch", "request", "create",
		"setDefaults", "getDefaults", "getAvailableFunctions", "setSilentMode",
	}
	return js.ValueOf(functions)
}

// Fonction GET
func get(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for GET request")
	}

	url := args[0].String()
	var config RequestConfig

	// Start with global defaults
	config = globalDefaults

	// Configuration optionnelle
	if len(args) > 1 && !args[1].IsUndefined() {
		configJS := args[1]
		userConfig := parseConfig(configJS)
		config = mergeConfig(config, userConfig)
	}

	config.Method = "GET"
	config.URL = url

	return makeRequest(config)
}

// Fonction POST
func post(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for POST request")
	}

	url := args[0].String()
	var data interface{}

	// Start with global defaults
	config := globalDefaults

	// Data optionnelle
	if len(args) > 1 && !args[1].IsUndefined() {
		data = parseJSValue(args[1])
	}

	// Configuration optionnelle
	if len(args) > 2 && !args[2].IsUndefined() {
		userConfig := parseConfig(args[2])
		config = mergeConfig(config, userConfig)
	}

	config.Method = "POST"
	config.URL = url
	config.Data = data

	return makeRequest(config)
}

// Fonction PUT
func put(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for PUT request")
	}

	url := args[0].String()
	var config RequestConfig
	var data interface{}

	if len(args) > 1 && !args[1].IsUndefined() {
		data = parseJSValue(args[1])
	}

	if len(args) > 2 && !args[2].IsUndefined() {
		config = parseConfig(args[2])
	}

	config.Method = "PUT"
	config.URL = url
	config.Data = data

	return makeRequest(config)
}

// Fonction DELETE
func delete(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for DELETE request")
	}

	url := args[0].String()
	var config RequestConfig

	if len(args) > 1 && !args[1].IsUndefined() {
		config = parseConfig(args[1])
	}

	config.Method = "DELETE"
	config.URL = url

	return makeRequest(config)
}

// Fonction PATCH
func patch(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for PATCH request")
	}

	url := args[0].String()
	var config RequestConfig
	var data interface{}

	if len(args) > 1 && !args[1].IsUndefined() {
		data = parseJSValue(args[1])
	}

	if len(args) > 2 && !args[2].IsUndefined() {
		config = parseConfig(args[2])
	}

	config.Method = "PATCH"
	config.URL = url
	config.Data = data

	return makeRequest(config)
}

// Fonction générique pour faire des requêtes
func request(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("Configuration is required for request")
	}

	config := parseConfig(args[0])
	return makeRequest(config)
}

// Fonction pour créer une instance avec des valeurs par défaut
func create(this js.Value, args []js.Value) interface{} {
	var defaultConfig RequestConfig

	if len(args) > 0 && !args[0].IsUndefined() {
		defaultConfig = parseConfig(args[0])
	}

	// Créer un objet instance avec les méthodes
	instance := js.Global().Get("Object").New()

	// Ajouter les méthodes à l'instance
	instance.Set("get", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return instanceGet(defaultConfig, args)
	}))

	instance.Set("post", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return instancePost(defaultConfig, args)
	}))

	instance.Set("put", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return instancePut(defaultConfig, args)
	}))

	instance.Set("delete", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return instanceDelete(defaultConfig, args)
	}))

	instance.Set("patch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return instancePatch(defaultConfig, args)
	}))

	instance.Set("request", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return instanceRequest(defaultConfig, args)
	}))

	return instance
}

// Fonctions d'instance qui utilisent la configuration par défaut
func instanceGet(defaultConfig RequestConfig, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for GET request")
	}

	config := mergeConfig(defaultConfig, RequestConfig{
		Method: "GET",
		URL:    args[0].String(),
	})

	if len(args) > 1 && !args[1].IsUndefined() {
		userConfig := parseConfig(args[1])
		config = mergeConfig(config, userConfig)
	}

	return makeRequest(config)
}

func instancePost(defaultConfig RequestConfig, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for POST request")
	}

	config := mergeConfig(defaultConfig, RequestConfig{
		Method: "POST",
		URL:    args[0].String(),
	})

	if len(args) > 1 && !args[1].IsUndefined() {
		config.Data = parseJSValue(args[1])
	}

	if len(args) > 2 && !args[2].IsUndefined() {
		userConfig := parseConfig(args[2])
		config = mergeConfig(config, userConfig)
	}

	return makeRequest(config)
}

func instancePut(defaultConfig RequestConfig, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for PUT request")
	}

	config := mergeConfig(defaultConfig, RequestConfig{
		Method: "PUT",
		URL:    args[0].String(),
	})

	if len(args) > 1 && !args[1].IsUndefined() {
		config.Data = parseJSValue(args[1])
	}

	if len(args) > 2 && !args[2].IsUndefined() {
		userConfig := parseConfig(args[2])
		config = mergeConfig(config, userConfig)
	}

	return makeRequest(config)
}

func instanceDelete(defaultConfig RequestConfig, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for DELETE request")
	}

	config := mergeConfig(defaultConfig, RequestConfig{
		Method: "DELETE",
		URL:    args[0].String(),
	})

	if len(args) > 1 && !args[1].IsUndefined() {
		userConfig := parseConfig(args[1])
		config = mergeConfig(config, userConfig)
	}

	return makeRequest(config)
}

func instancePatch(defaultConfig RequestConfig, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("URL is required for PATCH request")
	}

	config := mergeConfig(defaultConfig, RequestConfig{
		Method: "PATCH",
		URL:    args[0].String(),
	})

	if len(args) > 1 && !args[1].IsUndefined() {
		config.Data = parseJSValue(args[1])
	}

	if len(args) > 2 && !args[2].IsUndefined() {
		userConfig := parseConfig(args[2])
		config = mergeConfig(config, userConfig)
	}

	return makeRequest(config)
}

func instanceRequest(defaultConfig RequestConfig, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorPromise("Configuration is required for request")
	}

	userConfig := parseConfig(args[0])
	config := mergeConfig(defaultConfig, userConfig)

	return makeRequest(config)
}

// Fonction utilitaire pour fusionner les configurations
func mergeConfig(base, override RequestConfig) RequestConfig {
	result := base

	if override.Method != "" {
		result.Method = override.Method
	}
	if override.URL != "" {
		result.URL = override.URL
	}
	if override.Data != nil {
		result.Data = override.Data
	}
	if override.Timeout > 0 {
		result.Timeout = override.Timeout
	}

	// Fusionner les headers
	if result.Headers == nil {
		result.Headers = make(map[string]string)
	}
	for k, v := range override.Headers {
		result.Headers[k] = v
	}

	return result
}

// Fonction utilitaire pour parser la configuration JavaScript
func parseConfig(configJS js.Value) RequestConfig {
	config := RequestConfig{
		Headers: make(map[string]string),
		Timeout: 5000, // timeout par défaut de 5 secondes
	}

	if !configJS.IsUndefined() {
		if method := configJS.Get("method"); !method.IsUndefined() {
			config.Method = strings.ToUpper(method.String())
		}
		if url := configJS.Get("url"); !url.IsUndefined() {
			config.URL = url.String()
		}
		if data := configJS.Get("data"); !data.IsUndefined() {
			config.Data = parseJSValue(data)
		}
		if timeout := configJS.Get("timeout"); !timeout.IsUndefined() {
			config.Timeout = timeout.Int()
		}
		if headers := configJS.Get("headers"); !headers.IsUndefined() {
			parseHeaders(headers, config.Headers)
		}
	}

	return config
}

// Fonction utilitaire pour parser les headers
func parseHeaders(headersJS js.Value, headers map[string]string) {
	if headersJS.Type() == js.TypeObject {
		keys := js.Global().Get("Object").Call("keys", headersJS)
		length := keys.Get("length").Int()

		for i := 0; i < length; i++ {
			key := keys.Index(i).String()
			value := headersJS.Get(key).String()
			headers[key] = value
		}
	}
}

// Fonction utilitaire pour parser une valeur JavaScript
func parseJSValue(value js.Value) interface{} {
	switch value.Type() {
	case js.TypeString:
		return value.String()
	case js.TypeNumber:
		return value.Float()
	case js.TypeBoolean:
		return value.Bool()
	case js.TypeObject:
		if value.Get("constructor").Get("name").String() == "Array" {
			length := value.Get("length").Int()
			arr := make([]interface{}, length)
			for i := 0; i < length; i++ {
				arr[i] = parseJSValue(value.Index(i))
			}
			return arr
		} else {
			obj := make(map[string]interface{})
			keys := js.Global().Get("Object").Call("keys", value)
			length := keys.Get("length").Int()

			for i := 0; i < length; i++ {
				key := keys.Index(i).String()
				obj[key] = parseJSValue(value.Get(key))
			}
			return obj
		}
	default:
		return nil
	}
}

// Fonction principale pour faire la requête HTTP
func makeRequest(config RequestConfig) interface{} {
	// Create promise
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			// Get client from pool
			client := getClient(time.Duration(config.Timeout) * time.Millisecond)
			defer putClient(client)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Millisecond)
			defer cancel()

			// Prepare request body
			var body io.Reader
			if config.Data != nil {
				buf := getBuffer()
				defer putBuffer(buf)

				jsonData, err := json.Marshal(config.Data)
				if err != nil {
					rejectWithError(reject, StructuredError{
						Code:    "INVALID_DATA",
						Message: fmt.Sprintf("Failed to marshal request data: %v", err),
						Details: map[string]interface{}{
							"error": err.Error(),
							"data":  fmt.Sprintf("%v", config.Data),
						},
					})
					return
				}
				body = bytes.NewReader(jsonData)
			}

			// Create request
			req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, body)
			if err != nil {
				rejectWithError(reject, StructuredError{
					Code:    "INVALID_REQUEST",
					Message: fmt.Sprintf("Failed to create request: %v", err),
					Details: map[string]interface{}{
						"error":  err.Error(),
						"method": config.Method,
						"url":    config.URL,
					},
				})
				return
			}

			// Set headers
			for key, value := range config.Headers {
				req.Header.Set(key, value)
			}

			// Make request
			resp, err := client.Do(req)
			if err != nil {
				rejectWithError(reject, StructuredError{
					Code:    "REQUEST_FAILED",
					Message: fmt.Sprintf("Request failed: %v", err),
					Details: map[string]interface{}{
						"error":  err.Error(),
						"method": config.Method,
						"url":    config.URL,
					},
				})
				return
			}
			defer resp.Body.Close()

			// Check response size
			if config.MaxBodySize > 0 && resp.ContentLength > config.MaxBodySize {
				rejectWithError(reject, StructuredError{
					Code:    "RESPONSE_TOO_LARGE",
					Message: "Response body exceeds size limit",
					Details: map[string]interface{}{
						"limit": config.MaxBodySize,
						"size":  resp.ContentLength,
					},
				})
				return
			}

			// Read response body with buffer pool
			buf := getBuffer()
			defer putBuffer(buf)

			bodyReader := io.LimitReader(resp.Body, 1024*1024*10) // 10MB max
			bodyBytes, err := io.ReadAll(bodyReader)
			if err != nil {
				rejectWithError(reject, StructuredError{
					Code:    "READ_ERROR",
					Message: fmt.Sprintf("Failed to read response: %v", err),
					Details: map[string]interface{}{
						"error": err.Error(),
					},
				})
				return
			}

			// Parse response headers
			headers := make(map[string]string)
			for key, values := range resp.Header {
				headers[key] = strings.Join(values, ", ")
			}

			// Create response
			response := Response{
				Status:  resp.StatusCode,
				Headers: headers,
				Config:  config,
			}

			// Try to parse JSON response
			var jsonData interface{}
			if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
				response.Data = jsonData
			} else {
				response.Data = string(bodyBytes)
			}

			// Check for error status codes
			if resp.StatusCode >= 400 {
				response.Error = &StructuredError{
					Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
					Message: fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status),
					Details: map[string]interface{}{
						"status":     resp.StatusCode,
						"statusText": resp.Status,
						"headers":    headers,
					},
				}
			}

			resolve.Invoke(js.ValueOf(response))
			cleanupMemory()
		}()

		return nil
	})

	// Create and return promise
	return js.Global().Get("Promise").New(handler)
}

// rejectWithError - Helper to reject promise with structured error
func rejectWithError(reject js.Value, err StructuredError) {
	reject.Invoke(js.ValueOf(map[string]interface{}{
		"error": err,
	}))
}

// Fonction utilitaire pour créer une promesse d'erreur
func createErrorPromise(message string) interface{} {
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		reject := args[1]
		errorJS := convertToJSValue(StructuredError{
			Code:    "GENERIC_ERROR",
			Message: message,
		})
		reject.Invoke(errorJS)
		return nil
	}))
}

// Fonction utilitaire pour convertir les structures Go en valeurs JavaScript
func convertToJSValue(data interface{}) js.Value {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: %v", err))
	}

	jsonString := string(jsonBytes)
	return js.Global().Get("JSON").Call("parse", jsonString)
}

func main() {
	fmt.Println("Goxios WASM module initializing...")

	// Créer l'objet global goxios
	goxios := js.Global().Get("Object").New()

	// Enregistrer les fonctions principales
	goxios.Set("get", js.FuncOf(get))
	goxios.Set("post", js.FuncOf(post))
	goxios.Set("put", js.FuncOf(put))
	goxios.Set("delete", js.FuncOf(delete))
	goxios.Set("patch", js.FuncOf(patch))
	goxios.Set("request", js.FuncOf(request))
	goxios.Set("create", js.FuncOf(create))
	goxios.Set("setDefaults", js.FuncOf(setDefaults))
	goxios.Set("getDefaults", js.FuncOf(getDefaults))
	goxios.Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	goxios.Set("setSilentMode", js.FuncOf(setSilentMode))

	// Exposer l'objet goxios globalement
	js.Global().Set("goxios", goxios)

	// Exposer aussi les fonctions directement sur l'objet global
	js.Global().Set("get", js.FuncOf(get))
	js.Global().Set("post", js.FuncOf(post))
	js.Global().Set("put", js.FuncOf(put))
	js.Global().Set("delete", js.FuncOf(delete))
	js.Global().Set("patch", js.FuncOf(patch))
	js.Global().Set("request", js.FuncOf(request))
	js.Global().Set("create", js.FuncOf(create))
	js.Global().Set("setDefaults", js.FuncOf(setDefaults))
	js.Global().Set("getDefaults", js.FuncOf(getDefaults))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	// Ready signal for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Goxios WASM module initialized successfully")

	// Garder le programme en vie
	select {}
}
