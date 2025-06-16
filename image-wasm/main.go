//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"syscall/js"
)

var silentMode = false

// setSilentMode - Set silent mode for operations
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// compressJPEG - Compress JPEG image with specified quality
func compressJPEG(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: imageData and quality required")
	}

	// Get image data as Uint8Array
	imageDataArray := args[0]
	quality := int(args[1].Float())

	if quality < 1 || quality > 100 {
		return js.ValueOf("Error: quality must be between 1 and 100")
	}

	// Convert JS Uint8Array to Go []byte
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error decoding image: %v", err))
	}

	if !silentMode {
		fmt.Printf("Image decoded: format=%s, size=%dx%d\n", format, img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Compress as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error encoding JPEG: %v", err))
	}

	// Convert to Uint8Array for JavaScript
	compressedData := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(compressedData))
	js.CopyBytesToJS(dst, compressedData)

	if !silentMode {
		fmt.Printf("JPEG compressed: original=%d bytes, compressed=%d bytes, reduction=%.1f%%\n",
			len(imageData), len(compressedData),
			100.0*(1.0-float64(len(compressedData))/float64(len(imageData))))
	}

	return dst
}

// compressPNG - Process PNG image
func compressPNG(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: imageData required")
	}

	// Get image data as Uint8Array
	imageDataArray := args[0]

	// Convert JS Uint8Array to Go []byte
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error decoding image: %v", err))
	}

	if !silentMode {
		fmt.Printf("Image decoded: format=%s, size=%dx%d\n", format, img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Re-encode as PNG (this provides some optimization)
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error encoding PNG: %v", err))
	}

	// Convert to Uint8Array for JavaScript
	compressedData := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(compressedData))
	js.CopyBytesToJS(dst, compressedData)

	if !silentMode {
		fmt.Printf("PNG processed: original=%d bytes, result=%d bytes\n",
			len(imageData), len(compressedData))
	}

	return dst
}

// Simple bilinear resize implementation
func simpleResize(src image.Image, newWidth, newHeight int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	xRatio := float64(srcWidth) / float64(newWidth)
	yRatio := float64(srcHeight) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)

			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}

			pixel := src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY)
			dst.Set(x, y, pixel)
		}
	}

	return dst
}

// resizeImage - Resize image to specified dimensions
func resizeImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf("Error: imageData, width, and height required")
	}

	// Get parameters
	imageDataArray := args[0]
	width := int(args[1].Float())
	height := int(args[2].Float())

	if width <= 0 || height <= 0 {
		return js.ValueOf("Error: width and height must be positive")
	}

	// Convert JS Uint8Array to Go []byte
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error decoding image: %v", err))
	}

	originalBounds := img.Bounds()
	if !silentMode {
		fmt.Printf("Resizing image: format=%s, from %dx%d to %dx%d\n",
			format, originalBounds.Dx(), originalBounds.Dy(), width, height)
	}

	// Resize the image using simple algorithm
	resizedImg := simpleResize(img, width, height)

	// Encode back to original format
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, resizedImg)
	default:
		// Default to PNG for unknown formats
		err = png.Encode(&buf, resizedImg)
	}

	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error encoding resized image: %v", err))
	}

	// Convert to Uint8Array for JavaScript
	resizedData := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(resizedData))
	js.CopyBytesToJS(dst, resizedData)

	if !silentMode {
		fmt.Printf("Image resized: original=%d bytes, resized=%d bytes\n",
			len(imageData), len(resizedData))
	}

	return dst
}

// convertToWebP - Convert image to optimized format (simulated WebP as JPEG with high compression)
func convertToWebP(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: imageData required")
	}

	// Get image data as Uint8Array
	imageDataArray := args[0]
	quality := 75 // Default quality for "WebP simulation"

	if len(args) >= 2 {
		quality = int(args[1].Float())
	}

	if quality < 1 || quality > 100 {
		return js.ValueOf("Error: quality must be between 1 and 100")
	}

	// Convert JS Uint8Array to Go []byte
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error decoding image: %v", err))
	}

	if !silentMode {
		fmt.Printf("Converting to optimized format: format=%s, size=%dx%d\n", format, img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Encode as JPEG with specified quality (simulating WebP compression)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error encoding optimized image: %v", err))
	}

	// Convert to Uint8Array for JavaScript
	optimizedData := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(optimizedData))
	js.CopyBytesToJS(dst, optimizedData)

	if !silentMode {
		fmt.Printf("Image optimized: original=%d bytes, optimized=%d bytes, reduction=%.1f%%\n",
			len(imageData), len(optimizedData),
			100.0*(1.0-float64(len(optimizedData))/float64(len(imageData))))
	}

	return dst
}

// getImageInfo - Get information about an image
func getImageInfo(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: imageData required")
	}

	// Get image data as Uint8Array
	imageDataArray := args[0]

	// Convert JS Uint8Array to Go []byte
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error decoding image: %v", err))
	}

	bounds := img.Bounds()

	// Create JavaScript object directly
	jsInfo := js.Global().Get("Object").New()
	jsInfo.Set("format", js.ValueOf(format))
	jsInfo.Set("width", js.ValueOf(bounds.Dx()))
	jsInfo.Set("height", js.ValueOf(bounds.Dy()))
	jsInfo.Set("size", js.ValueOf(len(imageData)))

	return jsInfo
}

// getAvailableFunctions - Get list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []interface{}{
		"compressJPEG", "compressPNG", "convertToWebP", "resizeImage",
		"getImageInfo", "getAvailableFunctions", "setSilentMode",
	}
	return js.ValueOf(functions)
}

func main() {
	fmt.Println("Go WASM Image Processor initializing...")

	// Register all functions
	js.Global().Set("compressJPEG", js.FuncOf(compressJPEG))
	js.Global().Set("compressPNG", js.FuncOf(compressPNG))
	js.Global().Set("convertToWebP", js.FuncOf(convertToWebP))
	js.Global().Set("resizeImage", js.FuncOf(resizeImage))
	js.Global().Set("getImageInfo", js.FuncOf(getImageInfo))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	// Ready signal for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Go WASM Image Processor ready! Available functions: compressJPEG, compressPNG, convertToWebP, resizeImage, getImageInfo")

	// Keep the program alive
	select {}
}
