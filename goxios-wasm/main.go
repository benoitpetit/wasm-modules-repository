//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"syscall/js"
	"time"
)

var silentMode = false

// RequestConfig structure pour la configuration des requêtes
type RequestConfig struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Data    interface{}       `json:"data"`
	Timeout int               `json:"timeout"` // en millisecondes
}

// Response structure pour les réponses
type Response struct {
	Data    interface{}       `json:"data"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Config  RequestConfig     `json:"config"`
}

// Error structure pour les erreurs
type HTTPError struct {
	Message    string        `json:"message"`
	Status     int           `json:"status"`
	Response   *Response     `json:"response,omitempty"`
	Config     RequestConfig `json:"config"`
}

// Fonction pour activer/désactiver le mode silencieux
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// getAvailableFunctions - Get list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		"get", "post", "put", "delete", "patch", "request", "create", "getAvailableFunctions", "setSilentMode",
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

	// Configuration optionnelle
	if len(args) > 1 && !args[1].IsUndefined() {
		configJS := args[1]
		config = parseConfig(configJS)
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
	var config RequestConfig
	var data interface{}

	// Data optionnelle
	if len(args) > 1 && !args[1].IsUndefined() {
		data = parseJSValue(args[1])
	}

	// Configuration optionnelle
	if len(args) > 2 && !args[2].IsUndefined() {
		config = parseConfig(args[2])
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
	// Créer une Promise JavaScript
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			// Validation de l'URL
			if config.URL == "" {
				rejectWithError(reject, HTTPError{
					Message: "URL is required",
					Status:  0,
					Config:  config,
				})
				return
			}

			// Validation de la méthode
			if config.Method == "" {
				config.Method = "GET"
			}

			// Préparation des données
			var dataString string
			if config.Data != nil {
				if config.Headers == nil {
					config.Headers = make(map[string]string)
				}

				// Si les données sont un objet, les convertir en JSON
				if _, ok := config.Data.(map[string]interface{}); ok {
					dataBytes, err := json.Marshal(config.Data)
					if err != nil {
						rejectWithError(reject, HTTPError{
							Message: fmt.Sprintf("Failed to marshal request data: %v", err),
							Status:  0,
							Config:  config,
						})
						return
					}
					dataString = string(dataBytes)
					if config.Headers["Content-Type"] == "" {
						config.Headers["Content-Type"] = "application/json"
					}
				} else if str, ok := config.Data.(string); ok {
					dataString = str
				}
			}

			// Créer la requête HTTP
			var req *http.Request
			var err error

			if dataString != "" {
				req, err = http.NewRequest(config.Method, config.URL, strings.NewReader(dataString))
			} else {
				req, err = http.NewRequest(config.Method, config.URL, nil)
			}

			if err != nil {
				rejectWithError(reject, HTTPError{
					Message: fmt.Sprintf("Failed to create request: %v", err),
					Status:  0,
					Config:  config,
				})
				return
			}

			// Ajouter les headers
			for key, value := range config.Headers {
				req.Header.Set(key, value)
			}

			// Créer le client HTTP avec timeout
			client := &http.Client{
				Timeout: time.Duration(config.Timeout) * time.Millisecond,
			}

			if !silentMode {
				fmt.Printf("Goxios WASM: %s %s\n", config.Method, config.URL)
			}

			// Faire la requête
			resp, err := client.Do(req)
			if err != nil {
				rejectWithError(reject, HTTPError{
					Message: fmt.Sprintf("Request failed: %v", err),
					Status:  0,
					Config:  config,
				})
				return
			}
			defer resp.Body.Close()

			// Lire la réponse
			var responseData interface{}
			contentType := resp.Header.Get("Content-Type")

			if strings.Contains(contentType, "application/json") {
				var jsonData interface{}
				decoder := json.NewDecoder(resp.Body)
				if err := decoder.Decode(&jsonData); err == nil {
					responseData = jsonData
				}
			} else {
				// Pour les autres types de contenu, lire comme string
				bodyBytes := make([]byte, 0)
				buffer := make([]byte, 1024)
				for {
					n, err := resp.Body.Read(buffer)
					if n > 0 {
						bodyBytes = append(bodyBytes, buffer[:n]...)
					}
					if err != nil {
						break
					}
				}
				responseData = string(bodyBytes)
			}

			// Créer la réponse
			response := Response{
				Data:   responseData,
				Status: resp.StatusCode,
				Headers: make(map[string]string),
				Config: config,
			}

			// Copier les headers de réponse
			for key, values := range resp.Header {
				if len(values) > 0 {
					response.Headers[key] = values[0]
				}
			}

			// Vérifier le status code
			if resp.StatusCode >= 400 {
				rejectWithError(reject, HTTPError{
					Message:  fmt.Sprintf("Request failed with status %d", resp.StatusCode),
					Status:   resp.StatusCode,
					Response: &response,
					Config:   config,
				})
				return
			}

			// Convertir la réponse en objet JavaScript
			responseJS := convertToJSValue(response)
			resolve.Invoke(responseJS)

			if !silentMode {
				fmt.Printf("Goxios WASM: Response %d from %s\n", resp.StatusCode, config.URL)
			}
		}()

		return nil
	}))
}

// Fonction utilitaire pour rejeter une promesse avec une erreur
func rejectWithError(reject js.Value, err HTTPError) {
	errorJS := convertToJSValue(err)
	reject.Invoke(errorJS)
}

// Fonction utilitaire pour créer une promesse d'erreur
func createErrorPromise(message string) interface{} {
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		reject := args[1]
		errorJS := convertToJSValue(HTTPError{
			Message: message,
			Status:  0,
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
	// Signaler que le module est prêt
	fmt.Println("Goxios WASM module initialized successfully")

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
	goxios.Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	goxios.Set("setSilentMode", js.FuncOf(setSilentMode))

	// Exposer l'objet goxios globalement
	js.Global().Set("goxios", goxios)

	// Garder le programme en vie
	select {}
}
