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
	requestPool          sync.Pool
	clientPool           sync.Pool
	requestInterceptors  []js.Value
	responseInterceptors []js.Value
	interceptorsMutex    sync.RWMutex
	circuitBreaker       *CircuitBreaker
	circuitBreakerMutex  sync.RWMutex
)

// CircuitBreaker manages circuit breaker state
type CircuitBreaker struct {
	State            string // "closed", "open", "half-open"
	FailureCount     int
	SuccessCount     int
	Threshold        int
	Timeout          int64 // milliseconds
	LastFailureTime  int64 // timestamp
	HalfOpenAttempts int
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	Enabled      bool
	MaxRetries   int
	InitialDelay int // milliseconds
	MaxDelay     int // milliseconds
	Multiplier   float64
	RetryOn      []int // HTTP status codes to retry on
}

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
	Retry       *RetryConfig      `json:"retry,omitempty"`
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

	// Initialize interceptors
	requestInterceptors = make([]js.Value, 0)
	responseInterceptors = make([]js.Value, 0)

	// Initialize circuit breaker with defaults
	circuitBreaker = &CircuitBreaker{
		State:     "closed",
		Threshold: 5,
		Timeout:   60000, // 60 seconds
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

// addRequestInterceptor - Add request interceptor function
func addRequestInterceptor(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 || args[0].Type() != js.TypeFunction {
		return js.ValueOf(map[string]interface{}{
			"error": "Function required for addRequestInterceptor",
		})
	}

	interceptorsMutex.Lock()
	defer interceptorsMutex.Unlock()

	requestInterceptors = append(requestInterceptors, args[0])

	if !silentMode {
		fmt.Printf("Goxios WASM: Request interceptor added (total: %d)\n", len(requestInterceptors))
	}

	return js.ValueOf(map[string]interface{}{
		"success": true,
		"count":   len(requestInterceptors),
	})
}

// addResponseInterceptor - Add response interceptor function
func addResponseInterceptor(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 || args[0].Type() != js.TypeFunction {
		return js.ValueOf(map[string]interface{}{
			"error": "Function required for addResponseInterceptor",
		})
	}

	interceptorsMutex.Lock()
	defer interceptorsMutex.Unlock()

	responseInterceptors = append(responseInterceptors, args[0])

	if !silentMode {
		fmt.Printf("Goxios WASM: Response interceptor added (total: %d)\n", len(responseInterceptors))
	}

	return js.ValueOf(map[string]interface{}{
		"success": true,
		"count":   len(responseInterceptors),
	})
}

// clearInterceptors - Clear all interceptors
func clearInterceptors(this js.Value, args []js.Value) interface{} {
	interceptorsMutex.Lock()
	defer interceptorsMutex.Unlock()

	requestCount := len(requestInterceptors)
	responseCount := len(responseInterceptors)

	requestInterceptors = make([]js.Value, 0)
	responseInterceptors = make([]js.Value, 0)

	if !silentMode {
		fmt.Printf("Goxios WASM: Interceptors cleared (req: %d, res: %d)\n", requestCount, responseCount)
	}

	return js.ValueOf(map[string]interface{}{
		"success":         true,
		"requestCleared":  requestCount,
		"responseCleared": responseCount,
	})
}

// getCircuitBreakerState - Get circuit breaker state
func getCircuitBreakerState(this js.Value, args []js.Value) interface{} {
	circuitBreakerMutex.RLock()
	defer circuitBreakerMutex.RUnlock()

	return js.ValueOf(map[string]interface{}{
		"state":        circuitBreaker.State,
		"failureCount": circuitBreaker.FailureCount,
		"successCount": circuitBreaker.SuccessCount,
		"threshold":    circuitBreaker.Threshold,
		"timeout":      circuitBreaker.Timeout,
	})
}

// resetCircuitBreaker - Manually reset circuit breaker
func resetCircuitBreaker(this js.Value, args []js.Value) interface{} {
	circuitBreakerMutex.Lock()
	defer circuitBreakerMutex.Unlock()

	oldState := circuitBreaker.State
	circuitBreaker.State = "closed"
	circuitBreaker.FailureCount = 0
	circuitBreaker.SuccessCount = 0
	circuitBreaker.LastFailureTime = 0
	circuitBreaker.HalfOpenAttempts = 0

	if !silentMode {
		fmt.Printf("Goxios WASM: Circuit breaker reset (was: %s)\n", oldState)
	}

	return js.ValueOf(map[string]interface{}{
		"success":  true,
		"oldState": oldState,
		"newState": "closed",
	})
}

// configureCircuitBreaker - Configure circuit breaker settings
func configureCircuitBreaker(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "Configuration object required",
		})
	}

	circuitBreakerMutex.Lock()
	defer circuitBreakerMutex.Unlock()

	config := args[0]

	if threshold := config.Get("threshold"); !threshold.IsUndefined() {
		circuitBreaker.Threshold = threshold.Int()
	}
	if timeout := config.Get("timeout"); !timeout.IsUndefined() {
		circuitBreaker.Timeout = int64(timeout.Int())
	}

	if !silentMode {
		fmt.Printf("Goxios WASM: Circuit breaker configured (threshold: %d, timeout: %dms)\n",
			circuitBreaker.Threshold, circuitBreaker.Timeout)
	}

	return js.ValueOf(map[string]interface{}{
		"success":   true,
		"threshold": circuitBreaker.Threshold,
		"timeout":   circuitBreaker.Timeout,
	})
}

// checkCircuitBreaker - Check if circuit breaker allows request
func checkCircuitBreaker() (bool, string) {
	circuitBreakerMutex.Lock()
	defer circuitBreakerMutex.Unlock()

	now := time.Now().UnixMilli()

	switch circuitBreaker.State {
	case "open":
		// Check if timeout has passed
		if now-circuitBreaker.LastFailureTime > circuitBreaker.Timeout {
			circuitBreaker.State = "half-open"
			circuitBreaker.HalfOpenAttempts = 0
			if !silentMode {
				fmt.Println("Goxios WASM: Circuit breaker transitioning to half-open")
			}
			return true, "half-open"
		}
		return false, "open"
	case "half-open":
		// Allow limited attempts
		if circuitBreaker.HalfOpenAttempts < 3 {
			circuitBreaker.HalfOpenAttempts++
			return true, "half-open"
		}
		return false, "half-open"
	default: // closed
		return true, "closed"
	}
}

// recordCircuitBreakerSuccess - Record successful request
func recordCircuitBreakerSuccess() {
	circuitBreakerMutex.Lock()
	defer circuitBreakerMutex.Unlock()

	circuitBreaker.SuccessCount++

	if circuitBreaker.State == "half-open" {
		// After 3 successes in half-open, close the circuit
		if circuitBreaker.SuccessCount >= 3 {
			circuitBreaker.State = "closed"
			circuitBreaker.FailureCount = 0
			circuitBreaker.SuccessCount = 0
			circuitBreaker.HalfOpenAttempts = 0
			if !silentMode {
				fmt.Println("Goxios WASM: Circuit breaker closed after successful recovery")
			}
		}
	}
}

// recordCircuitBreakerFailure - Record failed request
func recordCircuitBreakerFailure() {
	circuitBreakerMutex.Lock()
	defer circuitBreakerMutex.Unlock()

	circuitBreaker.FailureCount++
	circuitBreaker.LastFailureTime = time.Now().UnixMilli()

	if circuitBreaker.State == "half-open" {
		// Failure in half-open returns to open
		circuitBreaker.State = "open"
		circuitBreaker.SuccessCount = 0
		if !silentMode {
			fmt.Println("Goxios WASM: Circuit breaker reopened after failure")
		}
	} else if circuitBreaker.State == "closed" && circuitBreaker.FailureCount >= circuitBreaker.Threshold {
		// Too many failures, open the circuit
		circuitBreaker.State = "open"
		if !silentMode {
			fmt.Printf("Goxios WASM: Circuit breaker opened after %d failures\n", circuitBreaker.FailureCount)
		}
	}
}

// applyRequestInterceptors - Apply all request interceptors
func applyRequestInterceptors(config js.Value) js.Value {
	interceptorsMutex.RLock()
	defer interceptorsMutex.RUnlock()

	result := config
	for _, interceptor := range requestInterceptors {
		result = interceptor.Invoke(result)
	}
	return result
}

// applyResponseInterceptors - Apply all response interceptors
func applyResponseInterceptors(response js.Value) js.Value {
	interceptorsMutex.RLock()
	defer interceptorsMutex.RUnlock()

	result := response
	for _, interceptor := range responseInterceptors {
		result = interceptor.Invoke(result)
	}
	return result
}

// calculateBackoff - Calculate exponential backoff delay
func calculateBackoff(attempt int, config *RetryConfig) int {
	if config == nil {
		return 1000 // default 1 second
	}

	delay := float64(config.InitialDelay) * pow(config.Multiplier, float64(attempt))

	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return int(delay)
}

// pow - Simple power function for integers
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// shouldRetry - Check if request should be retried
func shouldRetry(statusCode int, config *RetryConfig) bool {
	if config == nil || !config.Enabled {
		return false
	}

	// Retry on network errors (status 0)
	if statusCode == 0 {
		return true
	}

	// Check if status code is in retry list
	for _, code := range config.RetryOn {
		if code == statusCode {
			return true
		}
	}

	return false
}

// getDefaults - Get current global default configuration
func getDefaults(this js.Value, args []js.Value) interface{} {
	return js.ValueOf(map[string]interface{}{
		"config": convertToJSValue(globalDefaults),
	})
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
	// Apply request interceptors
	configJS := convertToJSValue(config)
	configJS = applyRequestInterceptors(configJS)

	// Re-parse config after interceptors
	config = parseConfig(configJS)

	// Check circuit breaker
	allowed, state := checkCircuitBreaker()
	if !allowed {
		return createErrorPromise(fmt.Sprintf("Circuit breaker is %s - too many failures", state))
	}

	// Create promise
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			// Setup retry configuration
			retryConfig := config.Retry
			if retryConfig == nil {
				retryConfig = &RetryConfig{
					Enabled:      false,
					MaxRetries:   3,
					InitialDelay: 1000,
					MaxDelay:     30000,
					Multiplier:   2.0,
					RetryOn:      []int{408, 429, 500, 502, 503, 504},
				}
			}

			var lastError *StructuredError
			var response *Response
			attempts := 0
			maxAttempts := 1
			if retryConfig.Enabled {
				maxAttempts = retryConfig.MaxRetries + 1
			}

			// Retry loop
			for attempts < maxAttempts {
				// Make HTTP request
				resp, err := performHTTPRequest(config)

				if err == nil {
					response = resp
					// Check if we should retry based on status code
					if !shouldRetry(resp.Status, retryConfig) || attempts >= maxAttempts-1 {
						break
					}
					lastError = &StructuredError{
						Code:    fmt.Sprintf("HTTP_%d", resp.Status),
						Message: fmt.Sprintf("HTTP error: %d", resp.Status),
					}
				} else {
					lastError = err
					// Network errors are always retryable if retry is enabled
					if !retryConfig.Enabled || attempts >= maxAttempts-1 {
						break
					}
				}

				attempts++

				if attempts < maxAttempts {
					delay := calculateBackoff(attempts-1, retryConfig)
					if !silentMode {
						fmt.Printf("Goxios WASM: Retry attempt %d/%d after %dms\n", attempts, maxAttempts-1, delay)
					}
					time.Sleep(time.Duration(delay) * time.Millisecond)
				}
			}

			// Handle final result
			if lastError != nil {
				recordCircuitBreakerFailure()
				rejectWithError(reject, *lastError)
				return
			}

			// Record success
			recordCircuitBreakerSuccess()

			// Apply response interceptors
			responseJS := convertToJSValue(response)
			responseJS = applyResponseInterceptors(responseJS)

			resolve.Invoke(responseJS)
			cleanupMemory()
		}()

		return nil
	})

	// Create and return promise
	return js.Global().Get("Promise").New(handler)
}

// performHTTPRequest - Execute single HTTP request
func performHTTPRequest(config RequestConfig) (*Response, *StructuredError) {
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
			return nil, &StructuredError{
				Code:    "INVALID_DATA",
				Message: fmt.Sprintf("Failed to marshal request data: %v", err),
				Details: map[string]interface{}{
					"error": err.Error(),
					"data":  fmt.Sprintf("%v", config.Data),
				},
			}
		}
		body = bytes.NewReader(jsonData)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, body)
	if err != nil {
		return nil, &StructuredError{
			Code:    "INVALID_REQUEST",
			Message: fmt.Sprintf("Failed to create request: %v", err),
			Details: map[string]interface{}{
				"error":  err.Error(),
				"method": config.Method,
				"url":    config.URL,
			},
		}
	}

	// Set headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, &StructuredError{
			Code:    "REQUEST_FAILED",
			Message: fmt.Sprintf("Request failed: %v", err),
			Details: map[string]interface{}{
				"error":  err.Error(),
				"method": config.Method,
				"url":    config.URL,
			},
		}
	}
	defer resp.Body.Close()

	// Check response size
	if config.MaxBodySize > 0 && resp.ContentLength > config.MaxBodySize {
		return nil, &StructuredError{
			Code:    "RESPONSE_TOO_LARGE",
			Message: "Response body exceeds size limit",
			Details: map[string]interface{}{
				"limit": config.MaxBodySize,
				"size":  resp.ContentLength,
			},
		}
	}

	// Read response body with buffer pool
	buf := getBuffer()
	defer putBuffer(buf)

	bodyReader := io.LimitReader(resp.Body, 1024*1024*10) // 10MB max
	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, &StructuredError{
			Code:    "READ_ERROR",
			Message: fmt.Sprintf("Failed to read response: %v", err),
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
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

	return &response, nil
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

// getAvailableFunctions - Return list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		"get", "post", "put", "delete", "patch",
		"request", "create", "setDefaults", "getDefaults",
		"addRequestInterceptor", "addResponseInterceptor", "clearInterceptors",
		"getCircuitBreakerState", "resetCircuitBreaker", "configureCircuitBreaker",
		"getAvailableFunctions", "setSilentMode",
	}

	// Convert to JS Array (safe pattern)
	arr := js.Global().Get("Array").New(len(functions))
	for i, fn := range functions {
		arr.SetIndex(i, fn)
	}
	return arr
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

	// Interceptors
	goxios.Set("addRequestInterceptor", js.FuncOf(addRequestInterceptor))
	goxios.Set("addResponseInterceptor", js.FuncOf(addResponseInterceptor))
	goxios.Set("clearInterceptors", js.FuncOf(clearInterceptors))

	// Circuit Breaker
	goxios.Set("getCircuitBreakerState", js.FuncOf(getCircuitBreakerState))
	goxios.Set("resetCircuitBreaker", js.FuncOf(resetCircuitBreaker))
	goxios.Set("configureCircuitBreaker", js.FuncOf(configureCircuitBreaker))

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
	js.Global().Set("addRequestInterceptor", js.FuncOf(addRequestInterceptor))
	js.Global().Set("addResponseInterceptor", js.FuncOf(addResponseInterceptor))
	js.Global().Set("clearInterceptors", js.FuncOf(clearInterceptors))
	js.Global().Set("getCircuitBreakerState", js.FuncOf(getCircuitBreakerState))
	js.Global().Set("resetCircuitBreaker", js.FuncOf(resetCircuitBreaker))
	js.Global().Set("configureCircuitBreaker", js.FuncOf(configureCircuitBreaker))

	// Ready signal for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Goxios WASM module initialized successfully")
	fmt.Println("Features: HTTP methods, Interceptors, Retry with backoff, Circuit breaker")

	// Garder le programme en vie
	select {}
}
