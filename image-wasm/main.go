//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
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

// applyGrayscale - Convert image to grayscale
func applyGrayscale(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData required",
		})
	}

	imageDataArray := args[0]
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	grayImg := image.NewRGBA(bounds)

	// Apply grayscale conversion
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, a := oldColor.RGBA()
			// Luminosity method: 0.299*R + 0.587*G + 0.114*B
			gray := uint8((299*r + 587*g + 114*b) / 1000 / 256)
			grayImg.Set(x, y, color.RGBA{R: gray, G: gray, B: gray, A: uint8(a / 256)})
		}
	}

	// Encode result
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, grayImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, grayImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Grayscale applied: %dx%d\n", bounds.Dx(), bounds.Dy())
	}

	return dst
}

// adjustBrightness - Adjust image brightness (-100 to +100)
func adjustBrightness(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData and brightness level required",
		})
	}

	imageDataArray := args[0]
	brightness := int(args[1].Float())

	if brightness < -100 || brightness > 100 {
		return js.ValueOf(map[string]interface{}{
			"error": "brightness must be between -100 and +100",
		})
	}

	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	adjustedImg := image.NewRGBA(bounds)

	// Apply brightness adjustment
	brightnessValue := float64(brightness) * 2.55 // Convert to 0-255 range
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, a := oldColor.RGBA()

			newR := clamp(int(float64(r/256)+brightnessValue), 0, 255)
			newG := clamp(int(float64(g/256)+brightnessValue), 0, 255)
			newB := clamp(int(float64(b/256)+brightnessValue), 0, 255)

			adjustedImg.Set(x, y, color.RGBA{R: uint8(newR), G: uint8(newG), B: uint8(newB), A: uint8(a / 256)})
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, adjustedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, adjustedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Brightness adjusted: %+d\n", brightness)
	}

	return dst
}

// adjustContrast - Adjust image contrast (-100 to +100)
func adjustContrast(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData and contrast level required",
		})
	}

	imageDataArray := args[0]
	contrast := args[1].Float()

	if contrast < -100 || contrast > 100 {
		return js.ValueOf(map[string]interface{}{
			"error": "contrast must be between -100 and +100",
		})
	}

	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	adjustedImg := image.NewRGBA(bounds)

	// Calculate contrast factor
	factor := (259.0 * (contrast + 255.0)) / (255.0 * (259.0 - contrast))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, a := oldColor.RGBA()

			newR := clamp(int(factor*float64(r/256-128)+128), 0, 255)
			newG := clamp(int(factor*float64(g/256-128)+128), 0, 255)
			newB := clamp(int(factor*float64(b/256-128)+128), 0, 255)

			adjustedImg.Set(x, y, color.RGBA{R: uint8(newR), G: uint8(newG), B: uint8(newB), A: uint8(a / 256)})
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, adjustedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, adjustedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Contrast adjusted: %+.1f\n", contrast)
	}

	return dst
}

// invertColors - Invert image colors
func invertColors(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData required",
		})
	}

	imageDataArray := args[0]
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	invertedImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, a := oldColor.RGBA()
			invertedImg.Set(x, y, color.RGBA{
				R: uint8(255 - r/256),
				G: uint8(255 - g/256),
				B: uint8(255 - b/256),
				A: uint8(a / 256),
			})
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, invertedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, invertedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Println("Colors inverted")
	}

	return dst
}

// applyBlur - Apply box blur filter (radius: 1-10)
func applyBlur(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData and radius required",
		})
	}

	imageDataArray := args[0]
	radius := int(args[1].Float())

	if radius < 1 || radius > 10 {
		return js.ValueOf(map[string]interface{}{
			"error": "radius must be between 1 and 10",
		})
	}

	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	blurredImg := image.NewRGBA(bounds)

	// Simple box blur
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum, aSum, count uint32

			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx := x + dx
					ny := y + dy

					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						r, g, b, a := img.At(nx, ny).RGBA()
						rSum += r
						gSum += g
						bSum += b
						aSum += a
						count++
					}
				}
			}

			blurredImg.Set(x, y, color.RGBA{
				R: uint8(rSum / count / 256),
				G: uint8(gSum / count / 256),
				B: uint8(bSum / count / 256),
				A: uint8(aSum / count / 256),
			})
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, blurredImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, blurredImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Blur applied: radius=%d\n", radius)
	}

	return dst
}

// applySharpen - Apply sharpen filter
func applySharpen(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData required",
		})
	}

	imageDataArray := args[0]
	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	sharpenedImg := image.NewRGBA(bounds)

	// Sharpen kernel: [-1 -1 -1, -1 9 -1, -1 -1 -1]
	kernel := [][]int{
		{-1, -1, -1},
		{-1, 9, -1},
		{-1, -1, -1},
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum int

			for ky := 0; ky < 3; ky++ {
				for kx := 0; kx < 3; kx++ {
					nx := x + kx - 1
					ny := y + ky - 1

					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						r, g, b, _ := img.At(nx, ny).RGBA()
						weight := kernel[ky][kx]
						rSum += int(r/256) * weight
						gSum += int(g/256) * weight
						bSum += int(b/256) * weight
					}
				}
			}

			_, _, _, a := img.At(x, y).RGBA()
			sharpenedImg.Set(x, y, color.RGBA{
				R: uint8(clamp(rSum, 0, 255)),
				G: uint8(clamp(gSum, 0, 255)),
				B: uint8(clamp(bSum, 0, 255)),
				A: uint8(a / 256),
			})
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, sharpenedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, sharpenedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Println("Sharpen filter applied")
	}

	return dst
}

// rotateImage - Rotate image (angle: 90, 180, 270 degrees)
func rotateImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData and angle required (90, 180, 270)",
		})
	}

	imageDataArray := args[0]
	angle := int(args[1].Float())

	if angle != 90 && angle != 180 && angle != 270 {
		return js.ValueOf(map[string]interface{}{
			"error": "angle must be 90, 180, or 270 degrees",
		})
	}

	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	var rotatedImg *image.RGBA

	switch angle {
	case 90:
		rotatedImg = image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotatedImg.Set(bounds.Max.Y-1-y, x, img.At(x, y))
			}
		}
	case 180:
		rotatedImg = image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotatedImg.Set(bounds.Max.X-1-x, bounds.Max.Y-1-y, img.At(x, y))
			}
		}
	case 270:
		rotatedImg = image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotatedImg.Set(y, bounds.Max.X-1-x, img.At(x, y))
			}
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, rotatedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, rotatedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Image rotated: %d degrees\n", angle)
	}

	return dst
}

// flipImage - Flip image (direction: "horizontal" or "vertical")
func flipImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData and direction required (horizontal/vertical)",
		})
	}

	imageDataArray := args[0]
	direction := args[1].String()

	if direction != "horizontal" && direction != "vertical" {
		return js.ValueOf(map[string]interface{}{
			"error": "direction must be 'horizontal' or 'vertical'",
		})
	}

	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()
	flippedImg := image.NewRGBA(bounds)

	if direction == "horizontal" {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				flippedImg.Set(bounds.Max.X-1-x, y, img.At(x, y))
			}
		}
	} else {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				flippedImg.Set(x, bounds.Max.Y-1-y, img.At(x, y))
			}
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, flippedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, flippedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Image flipped: %s\n", direction)
	}

	return dst
}

// cropImage - Crop image (x, y, width, height)
func cropImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 5 {
		return js.ValueOf(map[string]interface{}{
			"error": "imageData, x, y, width, and height required",
		})
	}

	imageDataArray := args[0]
	cropX := int(args[1].Float())
	cropY := int(args[2].Float())
	cropWidth := int(args[3].Float())
	cropHeight := int(args[4].Float())

	if cropWidth <= 0 || cropHeight <= 0 {
		return js.ValueOf(map[string]interface{}{
			"error": "width and height must be positive",
		})
	}

	imageDataLen := imageDataArray.Get("length").Int()
	imageData := make([]byte, imageDataLen)
	js.CopyBytesToGo(imageData, imageDataArray)

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error decoding image: %v", err),
		})
	}

	bounds := img.Bounds()

	// Validate crop region
	if cropX < 0 || cropY < 0 ||
		cropX+cropWidth > bounds.Dx() ||
		cropY+cropHeight > bounds.Dy() {
		return js.ValueOf(map[string]interface{}{
			"error": "crop region out of bounds",
		})
	}

	croppedImg := image.NewRGBA(image.Rect(0, 0, cropWidth, cropHeight))

	for y := 0; y < cropHeight; y++ {
		for x := 0; x < cropWidth; x++ {
			croppedImg.Set(x, y, img.At(bounds.Min.X+cropX+x, bounds.Min.Y+cropY+y))
		}
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, croppedImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, croppedImg)
	}

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Error encoding image: %v", err),
		})
	}

	result := buf.Bytes()
	dst := js.Global().Get("Uint8Array").New(len(result))
	js.CopyBytesToJS(dst, result)

	if !silentMode {
		fmt.Printf("Image cropped: %dx%d at (%d,%d)\n", cropWidth, cropHeight, cropX, cropY)
	}

	return dst
}

// Helper function to clamp values
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// getAvailableFunctions - Get list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		// Compression & Conversion
		"compressJPEG", "compressPNG", "convertToWebP", "resizeImage",
		// Filters
		"applyGrayscale", "adjustBrightness", "adjustContrast", "invertColors",
		"applyBlur", "applySharpen",
		// Transformations
		"rotateImage", "flipImage", "cropImage",
		// Utilities
		"getImageInfo", "getAvailableFunctions", "setSilentMode",
	}

	// Convert to JS Array (safe pattern)
	arr := js.Global().Get("Array").New(len(functions))
	for i, fn := range functions {
		arr.SetIndex(i, fn)
	}
	return arr
}

func main() {
	fmt.Println("Go WASM Image Processor initializing...")

	// Register compression & conversion functions
	js.Global().Set("compressJPEG", js.FuncOf(compressJPEG))
	js.Global().Set("compressPNG", js.FuncOf(compressPNG))
	js.Global().Set("convertToWebP", js.FuncOf(convertToWebP))
	js.Global().Set("resizeImage", js.FuncOf(resizeImage))

	// Register filter functions
	js.Global().Set("applyGrayscale", js.FuncOf(applyGrayscale))
	js.Global().Set("adjustBrightness", js.FuncOf(adjustBrightness))
	js.Global().Set("adjustContrast", js.FuncOf(adjustContrast))
	js.Global().Set("invertColors", js.FuncOf(invertColors))
	js.Global().Set("applyBlur", js.FuncOf(applyBlur))
	js.Global().Set("applySharpen", js.FuncOf(applySharpen))

	// Register transformation functions
	js.Global().Set("rotateImage", js.FuncOf(rotateImage))
	js.Global().Set("flipImage", js.FuncOf(flipImage))
	js.Global().Set("cropImage", js.FuncOf(cropImage))

	// Register utility functions
	js.Global().Set("getImageInfo", js.FuncOf(getImageInfo))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))

	// Ready signal for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Go WASM Image Processor ready!")
	fmt.Println("Available: Compression, Filters (grayscale, brightness, contrast, invert, blur, sharpen), Transformations (rotate, flip, crop)")

	// Keep the program alive
	select {}
}
