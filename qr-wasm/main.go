//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/ean"
	"github.com/skip2/go-qrcode"
)

var silentMode = false

// QRResult represents QR code generation result
type QRResult struct {
	Data         string `json:"data"`
	Size         int    `json:"size"`
	Base64Image  string `json:"base64Image"`
	ErrorLevel   string `json:"errorLevel"`
	ContentType  string `json:"contentType"`
	OriginalData string `json:"originalData"`
	Error        string `json:"error,omitempty"`
}

// BarcodeResult represents barcode generation result
type BarcodeResult struct {
	Data         string `json:"data"`
	Type         string `json:"type"`
	Base64Image  string `json:"base64Image"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ContentType  string `json:"contentType"`
	OriginalData string `json:"originalData"`
	Error        string `json:"error,omitempty"`
}

// DecodeResult represents decode operation result
type DecodeResult struct {
	Success    bool   `json:"success"`
	Data       string `json:"data"`
	Type       string `json:"type"`
	Confidence int    `json:"confidence"`
	Error      string `json:"error,omitempty"`
}

// VCardData represents vCard contact information
type VCardData struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	URL          string `json:"url"`
	Address      string `json:"address"`
}

// WiFiData represents WiFi network information
type WiFiData struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
	Security string `json:"security"` // WPA, WEP, or empty for open
	Hidden   bool   `json:"hidden"`
}

// setSilentMode - Set silent mode for operations
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	if !silentMode {
		fmt.Printf("QR WASM: Silent mode set to %v\n", silentMode)
	}
	return js.ValueOf(silentMode)
}

// getAvailableFunctions - Return list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []interface{}{
		"generateQRCode",
		"decodeQRCode",
		"generateBarcode",
		"decodeBarcode",
		"generateVCard",
		"generateWiFiQR",
		"getAvailableFunctions",
		"setSilentMode",
	}
	return js.ValueOf(functions)
}

// generateQRCode - Generate QR code from text data
func generateQRCode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(QRResult{
			Error: "Erreur: au moins un argument requis (data)",
		})
	}

	data := args[0].String()
	size := 256 // default size
	errorLevel := qrcode.Medium

	if len(args) >= 2 {
		if sizeArg := args[1].Int(); sizeArg > 0 {
			size = sizeArg
		}
	}

	if len(args) >= 3 {
		level := strings.ToUpper(args[2].String())
		switch level {
		case "LOW":
			errorLevel = qrcode.Low
		case "MEDIUM":
			errorLevel = qrcode.Medium
		case "HIGH":
			errorLevel = qrcode.High
		case "HIGHEST":
			errorLevel = qrcode.Highest
		}
	}

	if !silentMode {
		fmt.Printf("QR WASM: Generating QR code for data: %s (size: %d)\n", data, size)
	}

	// Generate QR code
	qrBytes, err := qrcode.Encode(data, errorLevel, size)
	if err != nil {
		return js.ValueOf(QRResult{
			Error: fmt.Sprintf("Erreur lors de la génération du QR code: %v", err),
		})
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(qrBytes)

	result := QRResult{
		Data:         data,
		Size:         size,
		Base64Image:  base64Image,
		ErrorLevel:   getErrorLevelString(errorLevel),
		ContentType:  "image/png",
		OriginalData: data,
	}

	if !silentMode {
		fmt.Printf("QR WASM: QR code generated successfully (size: %d bytes)\n", len(qrBytes))
	}

	return js.ValueOf(map[string]interface{}{
		"data":         result.Data,
		"size":         result.Size,
		"base64Image":  result.Base64Image,
		"errorLevel":   result.ErrorLevel,
		"contentType":  result.ContentType,
		"originalData": result.OriginalData,
	})
}

// generateBarcode - Generate barcode from data
func generateBarcode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(BarcodeResult{
			Error: "Erreur: au moins un argument requis (data)",
		})
	}

	data := args[0].String()
	barcodeType := "code128" // default
	width := 200
	height := 100

	if len(args) >= 2 {
		barcodeType = strings.ToLower(args[1].String())
	}

	if len(args) >= 3 {
		if w := args[2].Int(); w > 0 {
			width = w
		}
	}

	if len(args) >= 4 {
		if h := args[3].Int(); h > 0 {
			height = h
		}
	}

	if !silentMode {
		fmt.Printf("QR WASM: Generating %s barcode for data: %s\n", barcodeType, data)
	}

	var barcodeObj barcode.Barcode
	var err error

	switch barcodeType {
	case "code128":
		barcodeObj, err = code128.Encode(data)
	case "code39":
		barcodeObj, err = code39.Encode(data, true, true)
	case "ean13":
		barcodeObj, err = ean.Encode(data)
	case "ean8":
		barcodeObj, err = ean.Encode(data)
	default:
		return js.ValueOf(BarcodeResult{
			Error: fmt.Sprintf("Type de code-barres non supporté: %s", barcodeType),
		})
	}

	if err != nil {
		return js.ValueOf(BarcodeResult{
			Error: fmt.Sprintf("Erreur lors de la génération du code-barres: %v", err),
		})
	}

	// Scale barcode
	scaledBarcode, err := barcode.Scale(barcodeObj, width, height)
	if err != nil {
		return js.ValueOf(BarcodeResult{
			Error: fmt.Sprintf("Erreur lors du redimensionnement: %v", err),
		})
	}

	// Convert to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, scaledBarcode)
	if err != nil {
		return js.ValueOf(BarcodeResult{
			Error: fmt.Sprintf("Erreur lors de l'encodage PNG: %v", err),
		})
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())

	result := BarcodeResult{
		Data:         data,
		Type:         barcodeType,
		Base64Image:  base64Image,
		Width:        width,
		Height:       height,
		ContentType:  "image/png",
		OriginalData: data,
	}

	if !silentMode {
		fmt.Printf("QR WASM: Barcode generated successfully (%dx%d)\n", width, height)
	}

	return js.ValueOf(map[string]interface{}{
		"data":         result.Data,
		"type":         result.Type,
		"base64Image":  result.Base64Image,
		"width":        result.Width,
		"height":       result.Height,
		"contentType":  result.ContentType,
		"originalData": result.OriginalData,
	})
}

// generateVCard - Generate QR code with vCard contact information
func generateVCard(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(QRResult{
			Error: "Erreur: objet vCard requis",
		})
	}

	// Parse vCard data from JavaScript object
	vCardObj := args[0]
	var vCard VCardData

	if vCardObj.Get("name").Type() != js.TypeUndefined {
		vCard.Name = vCardObj.Get("name").String()
	}
	if vCardObj.Get("organization").Type() != js.TypeUndefined {
		vCard.Organization = vCardObj.Get("organization").String()
	}
	if vCardObj.Get("phone").Type() != js.TypeUndefined {
		vCard.Phone = vCardObj.Get("phone").String()
	}
	if vCardObj.Get("email").Type() != js.TypeUndefined {
		vCard.Email = vCardObj.Get("email").String()
	}
	if vCardObj.Get("url").Type() != js.TypeUndefined {
		vCard.URL = vCardObj.Get("url").String()
	}
	if vCardObj.Get("address").Type() != js.TypeUndefined {
		vCard.Address = vCardObj.Get("address").String()
	}

	// Build vCard format
	var vCardData strings.Builder
	vCardData.WriteString("BEGIN:VCARD\n")
	vCardData.WriteString("VERSION:3.0\n")

	if vCard.Name != "" {
		vCardData.WriteString(fmt.Sprintf("FN:%s\n", vCard.Name))
	}
	if vCard.Organization != "" {
		vCardData.WriteString(fmt.Sprintf("ORG:%s\n", vCard.Organization))
	}
	if vCard.Phone != "" {
		vCardData.WriteString(fmt.Sprintf("TEL:%s\n", vCard.Phone))
	}
	if vCard.Email != "" {
		vCardData.WriteString(fmt.Sprintf("EMAIL:%s\n", vCard.Email))
	}
	if vCard.URL != "" {
		vCardData.WriteString(fmt.Sprintf("URL:%s\n", vCard.URL))
	}
	if vCard.Address != "" {
		vCardData.WriteString(fmt.Sprintf("ADR:%s\n", vCard.Address))
	}

	vCardData.WriteString("END:VCARD")

	vCardString := vCardData.String()

	if !silentMode {
		fmt.Printf("QR WASM: Generating vCard QR code for: %s\n", vCard.Name)
	}

	size := 256
	if len(args) >= 2 {
		if sizeArg := args[1].Int(); sizeArg > 0 {
			size = sizeArg
		}
	}

	// Generate QR code
	qrBytes, err := qrcode.Encode(vCardString, qrcode.Medium, size)
	if err != nil {
		return js.ValueOf(QRResult{
			Error: fmt.Sprintf("Erreur lors de la génération du QR vCard: %v", err),
		})
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(qrBytes)

	result := QRResult{
		Data:         "vCard Contact",
		Size:         size,
		Base64Image:  base64Image,
		ErrorLevel:   "Medium",
		ContentType:  "image/png",
		OriginalData: vCardString,
	}

	if !silentMode {
		fmt.Printf("QR WASM: vCard QR code generated successfully\n")
	}

	return js.ValueOf(map[string]interface{}{
		"data":         result.Data,
		"size":         result.Size,
		"base64Image":  result.Base64Image,
		"errorLevel":   result.ErrorLevel,
		"contentType":  result.ContentType,
		"originalData": result.OriginalData,
	})
}

// generateWiFiQR - Generate QR code for WiFi network connection
func generateWiFiQR(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(QRResult{
			Error: "Erreur: objet WiFi requis",
		})
	}

	// Parse WiFi data from JavaScript object
	wifiObj := args[0]
	var wifi WiFiData

	if wifiObj.Get("ssid").Type() != js.TypeUndefined {
		wifi.SSID = wifiObj.Get("ssid").String()
	}
	if wifiObj.Get("password").Type() != js.TypeUndefined {
		wifi.Password = wifiObj.Get("password").String()
	}
	if wifiObj.Get("security").Type() != js.TypeUndefined {
		wifi.Security = strings.ToUpper(wifiObj.Get("security").String())
	}
	if wifiObj.Get("hidden").Type() != js.TypeUndefined {
		wifi.Hidden = wifiObj.Get("hidden").Bool()
	}

	if wifi.SSID == "" {
		return js.ValueOf(QRResult{
			Error: "Erreur: SSID requis pour le WiFi QR",
		})
	}

	// Build WiFi QR format: WIFI:T:WPA;S:mynetwork;P:mypass;H:false;;
	var wifiData strings.Builder
	wifiData.WriteString("WIFI:")

	// Security type
	if wifi.Security != "" {
		wifiData.WriteString(fmt.Sprintf("T:%s;", wifi.Security))
	} else {
		wifiData.WriteString("T:nopass;")
	}

	// SSID
	wifiData.WriteString(fmt.Sprintf("S:%s;", wifi.SSID))

	// Password
	if wifi.Password != "" {
		wifiData.WriteString(fmt.Sprintf("P:%s;", wifi.Password))
	}

	// Hidden network
	wifiData.WriteString(fmt.Sprintf("H:%s;;", strconv.FormatBool(wifi.Hidden)))

	wifiString := wifiData.String()

	if !silentMode {
		fmt.Printf("QR WASM: Generating WiFi QR code for network: %s\n", wifi.SSID)
	}

	size := 256
	if len(args) >= 2 {
		if sizeArg := args[1].Int(); sizeArg > 0 {
			size = sizeArg
		}
	}

	// Generate QR code
	qrBytes, err := qrcode.Encode(wifiString, qrcode.Medium, size)
	if err != nil {
		return js.ValueOf(QRResult{
			Error: fmt.Sprintf("Erreur lors de la génération du QR WiFi: %v", err),
		})
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(qrBytes)

	result := QRResult{
		Data:         fmt.Sprintf("WiFi Network: %s", wifi.SSID),
		Size:         size,
		Base64Image:  base64Image,
		ErrorLevel:   "Medium",
		ContentType:  "image/png",
		OriginalData: wifiString,
	}

	if !silentMode {
		fmt.Printf("QR WASM: WiFi QR code generated successfully\n")
	}

	return js.ValueOf(map[string]interface{}{
		"data":         result.Data,
		"size":         result.Size,
		"base64Image":  result.Base64Image,
		"errorLevel":   result.ErrorLevel,
		"contentType":  result.ContentType,
		"originalData": result.OriginalData,
	})
}

// decodeQRCode - Decode QR code from base64 image data
func decodeQRCode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(DecodeResult{
			Success: false,
			Error:   "Erreur: données d'image base64 requises",
		})
	}

	if !silentMode {
		fmt.Println("QR WASM: QR code decoding not fully implemented in this version")
	}

	return js.ValueOf(DecodeResult{
		Success: false,
		Error:   "Décodage QR non implémenté dans cette version",
		Type:    "qrcode",
	})
}

// decodeBarcode - Decode barcode from base64 image data
func decodeBarcode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(DecodeResult{
			Success: false,
			Error:   "Erreur: données d'image base64 requises",
		})
	}

	if !silentMode {
		fmt.Println("QR WASM: Barcode decoding not fully implemented in this version")
	}

	return js.ValueOf(DecodeResult{
		Success: false,
		Error:   "Décodage code-barres non implémenté dans cette version",
		Type:    "barcode",
	})
}

// Helper function to convert error level to string
func getErrorLevelString(level qrcode.RecoveryLevel) string {
	switch level {
	case qrcode.Low:
		return "Low"
	case qrcode.Medium:
		return "Medium"
	case qrcode.High:
		return "High"
	case qrcode.Highest:
		return "Highest"
	default:
		return "Medium"
	}
}

func main() {
	fmt.Println("QR WASM Module initializing...")

	// Register functions globally
	js.Global().Set("generateQRCode", js.FuncOf(generateQRCode))
	js.Global().Set("decodeQRCode", js.FuncOf(decodeQRCode))
	js.Global().Set("generateBarcode", js.FuncOf(generateBarcode))
	js.Global().Set("decodeBarcode", js.FuncOf(decodeBarcode))
	js.Global().Set("generateVCard", js.FuncOf(generateVCard))
	js.Global().Set("generateWiFiQR", js.FuncOf(generateWiFiQR))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	// Signal ready for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("QR WASM Module ready!")
	fmt.Println("Available functions:", "generateQRCode, decodeQRCode, generateBarcode, decodeBarcode, generateVCard, generateWiFiQR")

	// Keep the program running
	select {}
}
