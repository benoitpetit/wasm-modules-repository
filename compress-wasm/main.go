//go:build js && wasm

package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"syscall/js"
)

var silentMode = false

// setSilentMode enables/disables silent mode for console logs
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// ============================================================================
// GZIP Compression/Decompression
// ============================================================================

// gzipCompress compresses data using gzip algorithm
func gzipCompress(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for gzipCompress (data as string)")
	}

	data := []byte(args[0].String())

	// Optional compression level (1-9, default 6)
	level := gzip.DefaultCompression
	if len(args) >= 2 {
		l := args[1].Int()
		if l < 1 || l > 9 {
			return js.ValueOf("Error: compression level must be between 1 and 9")
		}
		level = l
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to create gzip writer: %v", err))
	}

	_, err = writer.Write(data)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to write gzip data: %v", err))
	}

	if err := writer.Close(); err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to close gzip writer: %v", err))
	}

	compressed := buf.Bytes()
	encoded := base64.StdEncoding.EncodeToString(compressed)

	ratio := float64(len(compressed)) / float64(len(data)) * 100

	result := map[string]interface{}{
		"data":             encoded,
		"originalSize":     len(data),
		"compressedSize":   len(compressed),
		"compressionRatio": fmt.Sprintf("%.2f%%", ratio),
		"algorithm":        "gzip",
		"level":            level,
	}

	if !silentMode {
		fmt.Printf("Go WASM: gzip compressed %d bytes → %d bytes (%.2f%%)\n", len(data), len(compressed), ratio)
	}

	return js.ValueOf(result)
}

// gzipDecompress decompresses gzip data
func gzipDecompress(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for gzipDecompress (base64-encoded compressed data)")
	}

	compressed, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to create gzip reader: %v", err))
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to decompress gzip data: %v", err))
	}

	if !silentMode {
		fmt.Printf("Go WASM: gzip decompressed %d bytes → %d bytes\n", len(compressed), len(decompressed))
	}

	return js.ValueOf(string(decompressed))
}

// ============================================================================
// Deflate Compression/Decompression
// ============================================================================

// deflateCompress compresses data using deflate algorithm
func deflateCompress(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for deflateCompress (data as string)")
	}

	data := []byte(args[0].String())

	level := flate.DefaultCompression
	if len(args) >= 2 {
		l := args[1].Int()
		if l < 1 || l > 9 {
			return js.ValueOf("Error: compression level must be between 1 and 9")
		}
		level = l
	}

	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, level)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to create deflate writer: %v", err))
	}

	_, err = writer.Write(data)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to write deflate data: %v", err))
	}

	if err := writer.Close(); err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to close deflate writer: %v", err))
	}

	compressed := buf.Bytes()
	encoded := base64.StdEncoding.EncodeToString(compressed)

	ratio := float64(len(compressed)) / float64(len(data)) * 100

	result := map[string]interface{}{
		"data":             encoded,
		"originalSize":     len(data),
		"compressedSize":   len(compressed),
		"compressionRatio": fmt.Sprintf("%.2f%%", ratio),
		"algorithm":        "deflate",
		"level":            level,
	}

	if !silentMode {
		fmt.Printf("Go WASM: deflate compressed %d bytes → %d bytes (%.2f%%)\n", len(data), len(compressed), ratio)
	}

	return js.ValueOf(result)
}

// deflateDecompress decompresses deflate data
func deflateDecompress(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for deflateDecompress (base64-encoded compressed data)")
	}

	compressed, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	reader := flate.NewReader(bytes.NewReader(compressed))
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to decompress deflate data: %v", err))
	}

	if !silentMode {
		fmt.Printf("Go WASM: deflate decompressed %d bytes → %d bytes\n", len(compressed), len(decompressed))
	}

	return js.ValueOf(string(decompressed))
}

// ============================================================================
// LZ4 Compression/Decompression (simplified block format)
// ============================================================================

// lz4Compress compresses data using a simplified LZ4-like block compression
func lz4Compress(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for lz4Compress (data as string)")
	}

	data := []byte(args[0].String())
	compressed := lz4CompressBlock(data)
	encoded := base64.StdEncoding.EncodeToString(compressed)

	ratio := float64(len(compressed)) / float64(len(data)) * 100

	result := map[string]interface{}{
		"data":             encoded,
		"originalSize":     len(data),
		"compressedSize":   len(compressed),
		"compressionRatio": fmt.Sprintf("%.2f%%", ratio),
		"algorithm":        "lz4",
	}

	if !silentMode {
		fmt.Printf("Go WASM: LZ4 compressed %d bytes → %d bytes (%.2f%%)\n", len(data), len(compressed), ratio)
	}

	return js.ValueOf(result)
}

// lz4CompressBlock implements a simplified LZ4-like block compression
func lz4CompressBlock(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	var result bytes.Buffer
	// Write original length as 4 bytes (big endian)
	origLen := len(data)
	result.WriteByte(byte(origLen >> 24))
	result.WriteByte(byte(origLen >> 16))
	result.WriteByte(byte(origLen >> 8))
	result.WriteByte(byte(origLen))

	hashTable := make(map[uint32]int)
	i := 0

	for i < len(data) {
		bestLen := 0
		bestOffset := 0

		if i+4 <= len(data) {
			h := hash4(data[i:])
			if prev, ok := hashTable[h]; ok && i-prev < 65536 && i-prev > 0 {
				// Try to find match
				matchLen := 0
				for i+matchLen < len(data) && prev+matchLen < i && data[prev+matchLen] == data[i+matchLen] {
					matchLen++
					if matchLen >= 255 {
						break
					}
				}
				if matchLen >= 4 {
					bestLen = matchLen
					bestOffset = i - prev
				}
			}
			hashTable[h] = i
		}

		if bestLen >= 4 {
			// Write literal count = 0, then match
			result.WriteByte(0) // no literals before match
			result.WriteByte(byte(bestOffset >> 8))
			result.WriteByte(byte(bestOffset))
			result.WriteByte(byte(bestLen))
			i += bestLen
		} else {
			// Write literal
			result.WriteByte(1) // literal marker
			result.WriteByte(data[i])
			i++
		}
	}

	return result.Bytes()
}

func hash4(data []byte) uint32 {
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}

// lz4Decompress decompresses LZ4-like block data
func lz4Decompress(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for lz4Decompress (base64-encoded compressed data)")
	}

	compressed, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	decompressed, err := lz4DecompressBlock(compressed)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to decompress LZ4 data: %v", err))
	}

	if !silentMode {
		fmt.Printf("Go WASM: LZ4 decompressed %d bytes → %d bytes\n", len(compressed), len(decompressed))
	}

	return js.ValueOf(string(decompressed))
}

func lz4DecompressBlock(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	origLen := int(data[0])<<24 | int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	result := make([]byte, 0, origLen)
	i := 4

	for i < len(data) {
		marker := data[i]
		i++

		if marker == 1 {
			// Literal
			if i >= len(data) {
				return nil, fmt.Errorf("unexpected end of data during literal")
			}
			result = append(result, data[i])
			i++
		} else if marker == 0 {
			// Match reference
			if i+2 >= len(data) {
				return nil, fmt.Errorf("unexpected end of data during match")
			}
			offset := int(data[i])<<8 | int(data[i+1])
			matchLen := int(data[i+2])
			i += 3

			start := len(result) - offset
			if start < 0 || start >= len(result) {
				return nil, fmt.Errorf("invalid match offset")
			}

			for j := 0; j < matchLen; j++ {
				result = append(result, result[start+j])
			}
		} else {
			return nil, fmt.Errorf("unknown marker: %d", marker)
		}
	}

	return result, nil
}

// ============================================================================
// Snappy Compression/Decompression (simplified)
// ============================================================================

// snappyCompress compresses data using a simplified Snappy-like algorithm
func snappyCompress(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for snappyCompress (data as string)")
	}

	data := []byte(args[0].String())
	compressed := snappyCompressBlock(data)
	encoded := base64.StdEncoding.EncodeToString(compressed)

	ratio := float64(len(compressed)) / float64(len(data)) * 100

	result := map[string]interface{}{
		"data":             encoded,
		"originalSize":     len(data),
		"compressedSize":   len(compressed),
		"compressionRatio": fmt.Sprintf("%.2f%%", ratio),
		"algorithm":        "snappy",
	}

	if !silentMode {
		fmt.Printf("Go WASM: Snappy compressed %d bytes → %d bytes (%.2f%%)\n", len(data), len(compressed), ratio)
	}

	return js.ValueOf(result)
}

// snappyCompressBlock implements a simplified Snappy-like compression
func snappyCompressBlock(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	var result bytes.Buffer

	// Write "SNPY" magic + original length
	result.Write([]byte("SNPY"))
	origLen := len(data)
	result.WriteByte(byte(origLen >> 24))
	result.WriteByte(byte(origLen >> 16))
	result.WriteByte(byte(origLen >> 8))
	result.WriteByte(byte(origLen))

	// Run-length + back-reference encoding
	i := 0
	for i < len(data) {
		// Try run-length encoding
		runLen := 1
		for i+runLen < len(data) && data[i+runLen] == data[i] && runLen < 255 {
			runLen++
		}

		if runLen >= 4 {
			// RLE: marker 0xFF, byte, count
			result.WriteByte(0xFF)
			result.WriteByte(data[i])
			result.WriteByte(byte(runLen))
			i += runLen
		} else {
			// Literal
			if data[i] == 0xFF {
				result.WriteByte(0xFF)
				result.WriteByte(data[i])
				result.WriteByte(1)
			} else {
				result.WriteByte(data[i])
			}
			i++
		}
	}

	return result.Bytes()
}

// snappyDecompress decompresses Snappy-like data
func snappyDecompress(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for snappyDecompress (base64-encoded compressed data)")
	}

	compressed, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	decompressed, err := snappyDecompressBlock(compressed)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: failed to decompress Snappy data: %v", err))
	}

	if !silentMode {
		fmt.Printf("Go WASM: Snappy decompressed %d bytes → %d bytes\n", len(compressed), len(decompressed))
	}

	return js.ValueOf(string(decompressed))
}

func snappyDecompressBlock(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}

	// Verify magic
	if string(data[:4]) != "SNPY" {
		return nil, fmt.Errorf("invalid snappy magic header")
	}

	origLen := int(data[4])<<24 | int(data[5])<<16 | int(data[6])<<8 | int(data[7])
	result := make([]byte, 0, origLen)
	i := 8

	for i < len(data) {
		if data[i] == 0xFF {
			if i+2 >= len(data) {
				return nil, fmt.Errorf("unexpected end of data during RLE")
			}
			b := data[i+1]
			count := int(data[i+2])
			for j := 0; j < count; j++ {
				result = append(result, b)
			}
			i += 3
		} else {
			result = append(result, data[i])
			i++
		}
	}

	return result, nil
}

// ============================================================================
// TAR Archive Creation/Extraction (simplified)
// ============================================================================

// tarCreate creates a TAR-like archive from multiple named data entries
func tarCreate(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for tarCreate (JSON object with filename:content pairs)")
	}

	input := args[0]
	if input.Type() != js.TypeObject {
		return js.ValueOf("Error: argument must be an object with filename:content pairs")
	}

	// Get all keys from the object
	keys := js.Global().Get("Object").Call("keys", input)
	numKeys := keys.Length()

	if numKeys == 0 {
		return js.ValueOf("Error: object must contain at least one file entry")
	}

	var archive bytes.Buffer
	// Magic header
	archive.Write([]byte("GTAR"))
	// Number of files
	archive.WriteByte(byte(numKeys >> 8))
	archive.WriteByte(byte(numKeys))

	totalSize := 0
	fileList := make([]string, 0, numKeys)

	for i := 0; i < numKeys; i++ {
		key := keys.Index(i).String()
		content := []byte(input.Get(key).String())

		fileList = append(fileList, key)

		// Write filename length + filename
		nameBytes := []byte(key)
		archive.WriteByte(byte(len(nameBytes) >> 8))
		archive.WriteByte(byte(len(nameBytes)))
		archive.Write(nameBytes)

		// Write content length + content
		archive.WriteByte(byte(len(content) >> 24))
		archive.WriteByte(byte(len(content) >> 16))
		archive.WriteByte(byte(len(content) >> 8))
		archive.WriteByte(byte(len(content)))
		archive.Write(content)

		totalSize += len(content)
	}

	encoded := base64.StdEncoding.EncodeToString(archive.Bytes())

	result := map[string]interface{}{
		"data":        encoded,
		"archiveSize": archive.Len(),
		"files":       numKeys,
		"fileList":    fileList,
		"totalSize":   totalSize,
		"format":      "tar",
	}

	if !silentMode {
		fmt.Printf("Go WASM: TAR archive created with %d files (%d bytes)\n", numKeys, archive.Len())
	}

	return js.ValueOf(result)
}

// tarExtract extracts files from a TAR-like archive
func tarExtract(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for tarExtract (base64-encoded archive data)")
	}

	data, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	if len(data) < 6 || string(data[:4]) != "GTAR" {
		return js.ValueOf("Error: invalid TAR archive format")
	}

	numFiles := int(data[4])<<8 | int(data[5])
	i := 6

	files := map[string]interface{}{}
	fileList := make([]string, 0)

	for f := 0; f < numFiles && i < len(data); f++ {
		if i+2 > len(data) {
			return js.ValueOf("Error: corrupted archive - unexpected end of data")
		}

		nameLen := int(data[i])<<8 | int(data[i+1])
		i += 2

		if i+nameLen > len(data) {
			return js.ValueOf("Error: corrupted archive - filename truncated")
		}

		name := string(data[i : i+nameLen])
		i += nameLen
		fileList = append(fileList, name)

		if i+4 > len(data) {
			return js.ValueOf("Error: corrupted archive - content length missing")
		}

		contentLen := int(data[i])<<24 | int(data[i+1])<<16 | int(data[i+2])<<8 | int(data[i+3])
		i += 4

		if i+contentLen > len(data) {
			return js.ValueOf("Error: corrupted archive - content truncated")
		}

		content := string(data[i : i+contentLen])
		i += contentLen

		files[name] = content
	}

	result := map[string]interface{}{
		"files":    files,
		"fileList": fileList,
		"count":    len(fileList),
	}

	if !silentMode {
		fmt.Printf("Go WASM: TAR archive extracted %d files\n", len(fileList))
	}

	return js.ValueOf(result)
}

// tarList lists files in a TAR archive without extracting
func tarList(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for tarList (base64-encoded archive data)")
	}

	data, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	if len(data) < 6 || string(data[:4]) != "GTAR" {
		return js.ValueOf("Error: invalid TAR archive format")
	}

	numFiles := int(data[4])<<8 | int(data[5])
	i := 6

	entries := make([]interface{}, 0)

	for f := 0; f < numFiles && i < len(data); f++ {
		if i+2 > len(data) {
			break
		}
		nameLen := int(data[i])<<8 | int(data[i+1])
		i += 2

		if i+nameLen > len(data) {
			break
		}
		name := string(data[i : i+nameLen])
		i += nameLen

		if i+4 > len(data) {
			break
		}
		contentLen := int(data[i])<<24 | int(data[i+1])<<16 | int(data[i+2])<<8 | int(data[i+3])
		i += 4

		entries = append(entries, map[string]interface{}{
			"name": name,
			"size": contentLen,
		})

		i += contentLen
	}

	result := map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	}

	if !silentMode {
		fmt.Printf("Go WASM: TAR archive contains %d files\n", len(entries))
	}

	return js.ValueOf(result)
}

// ============================================================================
// ZIP Archive Creation/Extraction (simplified with compression)
// ============================================================================

// zipCreate creates a ZIP-like archive with optional compression
func zipCreate(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for zipCreate (JSON object with filename:content pairs)")
	}

	input := args[0]
	if input.Type() != js.TypeObject {
		return js.ValueOf("Error: argument must be an object with filename:content pairs")
	}

	compress := true
	if len(args) >= 2 {
		compress = args[1].Bool()
	}

	keys := js.Global().Get("Object").Call("keys", input)
	numKeys := keys.Length()

	if numKeys == 0 {
		return js.ValueOf("Error: object must contain at least one file entry")
	}

	var archive bytes.Buffer
	// Magic header "GZIP"
	archive.Write([]byte("GZIP"))
	// Compression flag
	if compress {
		archive.WriteByte(1)
	} else {
		archive.WriteByte(0)
	}
	// Number of files
	archive.WriteByte(byte(numKeys >> 8))
	archive.WriteByte(byte(numKeys))

	totalOriginal := 0
	totalCompressed := 0
	fileList := make([]string, 0, numKeys)

	for i := 0; i < numKeys; i++ {
		key := keys.Index(i).String()
		content := []byte(input.Get(key).String())
		fileList = append(fileList, key)
		totalOriginal += len(content)

		nameBytes := []byte(key)
		archive.WriteByte(byte(len(nameBytes) >> 8))
		archive.WriteByte(byte(len(nameBytes)))
		archive.Write(nameBytes)

		// Write original size
		archive.WriteByte(byte(len(content) >> 24))
		archive.WriteByte(byte(len(content) >> 16))
		archive.WriteByte(byte(len(content) >> 8))
		archive.WriteByte(byte(len(content)))

		var stored []byte
		if compress && len(content) > 0 {
			var cbuf bytes.Buffer
			w, _ := flate.NewWriter(&cbuf, flate.DefaultCompression)
			w.Write(content)
			w.Close()
			stored = cbuf.Bytes()
		} else {
			stored = content
		}

		totalCompressed += len(stored)

		// Write stored size
		archive.WriteByte(byte(len(stored) >> 24))
		archive.WriteByte(byte(len(stored) >> 16))
		archive.WriteByte(byte(len(stored) >> 8))
		archive.WriteByte(byte(len(stored)))

		archive.Write(stored)
	}

	encoded := base64.StdEncoding.EncodeToString(archive.Bytes())

	ratio := float64(totalCompressed) / float64(totalOriginal) * 100
	if totalOriginal == 0 {
		ratio = 100.0
	}

	result := map[string]interface{}{
		"data":             encoded,
		"archiveSize":      archive.Len(),
		"files":            numKeys,
		"fileList":         fileList,
		"totalOriginal":    totalOriginal,
		"totalCompressed":  totalCompressed,
		"compressionRatio": fmt.Sprintf("%.2f%%", ratio),
		"compressed":       compress,
		"format":           "zip",
	}

	if !silentMode {
		fmt.Printf("Go WASM: ZIP archive created with %d files (%.2f%% ratio)\n", numKeys, ratio)
	}

	return js.ValueOf(result)
}

// zipExtract extracts files from a ZIP-like archive
func zipExtract(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for zipExtract (base64-encoded archive data)")
	}

	data, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	if len(data) < 7 || string(data[:4]) != "GZIP" {
		return js.ValueOf("Error: invalid ZIP archive format")
	}

	compress := data[4] == 1
	numFiles := int(data[5])<<8 | int(data[6])
	idx := 7

	files := map[string]interface{}{}
	fileList := make([]string, 0)

	for f := 0; f < numFiles && idx < len(data); f++ {
		if idx+2 > len(data) {
			return js.ValueOf("Error: corrupted archive")
		}
		nameLen := int(data[idx])<<8 | int(data[idx+1])
		idx += 2

		if idx+nameLen > len(data) {
			return js.ValueOf("Error: corrupted archive")
		}
		name := string(data[idx : idx+nameLen])
		idx += nameLen
		fileList = append(fileList, name)

		if idx+4 > len(data) {
			return js.ValueOf("Error: corrupted archive")
		}
		// Original size (skip, just for info)
		idx += 4

		if idx+4 > len(data) {
			return js.ValueOf("Error: corrupted archive")
		}
		storedLen := int(data[idx])<<24 | int(data[idx+1])<<16 | int(data[idx+2])<<8 | int(data[idx+3])
		idx += 4

		if idx+storedLen > len(data) {
			return js.ValueOf("Error: corrupted archive")
		}
		stored := data[idx : idx+storedLen]
		idx += storedLen

		var content []byte
		if compress && len(stored) > 0 {
			reader := flate.NewReader(bytes.NewReader(stored))
			content, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				return js.ValueOf(fmt.Sprintf("Error: failed to decompress file '%s': %v", name, err))
			}
		} else {
			content = stored
		}

		files[name] = string(content)
	}

	result := map[string]interface{}{
		"files":    files,
		"fileList": fileList,
		"count":    len(fileList),
	}

	if !silentMode {
		fmt.Printf("Go WASM: ZIP archive extracted %d files\n", len(fileList))
	}

	return js.ValueOf(result)
}

// zipList lists files in a ZIP archive without extracting
func zipList(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for zipList (base64-encoded archive data)")
	}

	data, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid base64 data: %v", err))
	}

	if len(data) < 7 || string(data[:4]) != "GZIP" {
		return js.ValueOf("Error: invalid ZIP archive format")
	}

	numFiles := int(data[5])<<8 | int(data[6])
	idx := 7

	entries := make([]interface{}, 0)

	for f := 0; f < numFiles && idx < len(data); f++ {
		if idx+2 > len(data) {
			break
		}
		nameLen := int(data[idx])<<8 | int(data[idx+1])
		idx += 2

		if idx+nameLen > len(data) {
			break
		}
		name := string(data[idx : idx+nameLen])
		idx += nameLen

		if idx+4 > len(data) {
			break
		}
		origSize := int(data[idx])<<24 | int(data[idx+1])<<16 | int(data[idx+2])<<8 | int(data[idx+3])
		idx += 4

		if idx+4 > len(data) {
			break
		}
		storedSize := int(data[idx])<<24 | int(data[idx+1])<<16 | int(data[idx+2])<<8 | int(data[idx+3])
		idx += 4

		entries = append(entries, map[string]interface{}{
			"name":         name,
			"originalSize": origSize,
			"storedSize":   storedSize,
		})

		idx += storedSize
	}

	result := map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	}

	if !silentMode {
		fmt.Printf("Go WASM: ZIP archive contains %d files\n", len(entries))
	}

	return js.ValueOf(result)
}

// ============================================================================
// Compression Analysis & Benchmarking
// ============================================================================

// analyzeCompression analyses compression ratio across all available algorithms
func analyzeCompression(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for analyzeCompression (data as string)")
	}

	data := []byte(args[0].String())
	originalSize := len(data)

	results := make([]interface{}, 0)

	// Gzip
	{
		var buf bytes.Buffer
		w, _ := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
		w.Write(data)
		w.Close()
		compSize := buf.Len()
		ratio := float64(compSize) / float64(originalSize) * 100
		results = append(results, map[string]interface{}{
			"algorithm":      "gzip",
			"compressedSize": compSize,
			"ratio":          fmt.Sprintf("%.2f%%", ratio),
			"savings":        fmt.Sprintf("%.2f%%", 100-ratio),
		})
	}

	// Deflate
	{
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flate.DefaultCompression)
		w.Write(data)
		w.Close()
		compSize := buf.Len()
		ratio := float64(compSize) / float64(originalSize) * 100
		results = append(results, map[string]interface{}{
			"algorithm":      "deflate",
			"compressedSize": compSize,
			"ratio":          fmt.Sprintf("%.2f%%", ratio),
			"savings":        fmt.Sprintf("%.2f%%", 100-ratio),
		})
	}

	// LZ4
	{
		comp := lz4CompressBlock(data)
		compSize := len(comp)
		ratio := float64(compSize) / float64(originalSize) * 100
		results = append(results, map[string]interface{}{
			"algorithm":      "lz4",
			"compressedSize": compSize,
			"ratio":          fmt.Sprintf("%.2f%%", ratio),
			"savings":        fmt.Sprintf("%.2f%%", 100-ratio),
		})
	}

	// Snappy
	{
		comp := snappyCompressBlock(data)
		compSize := len(comp)
		ratio := float64(compSize) / float64(originalSize) * 100
		results = append(results, map[string]interface{}{
			"algorithm":      "snappy",
			"compressedSize": compSize,
			"ratio":          fmt.Sprintf("%.2f%%", ratio),
			"savings":        fmt.Sprintf("%.2f%%", 100-ratio),
		})
	}

	// Gzip levels comparison
	gzipLevels := make([]interface{}, 0)
	for level := 1; level <= 9; level++ {
		var buf bytes.Buffer
		w, _ := gzip.NewWriterLevel(&buf, level)
		w.Write(data)
		w.Close()
		compSize := buf.Len()
		ratio := float64(compSize) / float64(originalSize) * 100
		gzipLevels = append(gzipLevels, map[string]interface{}{
			"level":          level,
			"compressedSize": compSize,
			"ratio":          fmt.Sprintf("%.2f%%", ratio),
		})
	}

	// Find best algorithm
	bestAlgo := "gzip"
	bestSize := originalSize
	for _, r := range results {
		rm := r.(map[string]interface{})
		cs := rm["compressedSize"].(int)
		if cs < bestSize {
			bestSize = cs
			bestAlgo = rm["algorithm"].(string)
		}
	}

	analysis := map[string]interface{}{
		"originalSize":    originalSize,
		"algorithms":      results,
		"gzipLevels":      gzipLevels,
		"bestAlgorithm":   bestAlgo,
		"bestSize":        bestSize,
		"bestRatio":       fmt.Sprintf("%.2f%%", float64(bestSize)/float64(originalSize)*100),
		"bestSavings":     fmt.Sprintf("%.2f%%", 100-float64(bestSize)/float64(originalSize)*100),
		"recommendation":  fmt.Sprintf("Use %s for best compression ratio", bestAlgo),
		"dataType":        detectDataType(data),
		"entropy":         calculateEntropy(data),
		"compressibility": classifyCompressibility(data),
	}

	if !silentMode {
		fmt.Printf("Go WASM: Compression analysis complete. Best: %s (%.2f%%)\n", bestAlgo, float64(bestSize)/float64(originalSize)*100)
	}

	return js.ValueOf(analysis)
}

// detectDataType tries to detect the type of data
func detectDataType(data []byte) string {
	if json.Valid(data) {
		return "json"
	}

	// Check if it looks like text
	textChars := 0
	for _, b := range data {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			textChars++
		}
	}
	if float64(textChars)/float64(len(data)) > 0.85 {
		return "text"
	}

	return "binary"
}

// calculateEntropy calculates Shannon entropy of data
func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	entropy := 0.0
	total := float64(len(data))
	for _, count := range freq {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * logBase2(p)
		}
	}

	return entropy
}

func logBase2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// ln(x) / ln(2)
	result := 0.0
	v := x
	for v < 1 {
		v *= 2
		result--
	}
	for v >= 2 {
		v /= 2
		result++
	}
	// Taylor series for ln(v) where 1 <= v < 2
	t := (v - 1) / (v + 1)
	t2 := t * t
	ln := t
	term := t
	for i := 0; i < 20; i++ {
		term *= t2
		ln += term / float64(2*i+3)
	}
	ln *= 2
	result += ln / 0.6931471805599453 // ln(2)
	return result
}

// classifyCompressibility provides a classification of the compressibility
func classifyCompressibility(data []byte) string {
	entropy := calculateEntropy(data)
	if entropy < 3.0 {
		return "highly-compressible"
	} else if entropy < 5.0 {
		return "moderately-compressible"
	} else if entropy < 7.0 {
		return "slightly-compressible"
	}
	return "barely-compressible"
}

// estimateCompression provides a quick estimate of compression potential
func estimateCompression(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: one argument required for estimateCompression (data as string)")
	}

	data := []byte(args[0].String())
	entropy := calculateEntropy(data)
	compressibility := classifyCompressibility(data)
	dataType := detectDataType(data)

	// Unique byte ratio
	uniqueBytes := make(map[byte]bool)
	for _, b := range data {
		uniqueBytes[b] = true
	}
	uniqueRatio := float64(len(uniqueBytes)) / 256.0

	// Estimate compression ratio based on entropy
	estimatedRatio := entropy / 8.0 * 100 // 8 bits max entropy

	result := map[string]interface{}{
		"originalSize":     len(data),
		"entropy":          fmt.Sprintf("%.4f", entropy),
		"maxEntropy":       8.0,
		"uniqueBytes":      len(uniqueBytes),
		"uniqueRatio":      fmt.Sprintf("%.2f%%", uniqueRatio*100),
		"compressibility":  compressibility,
		"estimatedRatio":   fmt.Sprintf("%.2f%%", estimatedRatio),
		"estimatedSavings": fmt.Sprintf("%.2f%%", 100-estimatedRatio),
		"dataType":         dataType,
		"recommendation":   getCompressionRecommendation(dataType, compressibility),
	}

	if !silentMode {
		fmt.Printf("Go WASM: Compression estimate - %s, entropy: %.4f\n", compressibility, entropy)
	}

	return js.ValueOf(result)
}

func getCompressionRecommendation(dataType string, compressibility string) string {
	switch {
	case compressibility == "highly-compressible":
		return "Any compression algorithm will work well. Use gzip for best ratio or snappy/lz4 for speed."
	case dataType == "json" || dataType == "text":
		return "Text data compresses well with gzip or deflate. Use level 6-9 for best results."
	case compressibility == "barely-compressible":
		return "Data has high entropy. Compression may not be beneficial. Consider if data is already compressed."
	default:
		return "Use gzip with default settings for a good balance of speed and compression ratio."
	}
}

// ============================================================================
// System Functions
// ============================================================================

// getModuleInfo returns comprehensive module information
func getModuleInfo(this js.Value, args []js.Value) interface{} {
	info := map[string]interface{}{
		"name":        "compress-wasm",
		"version":     "0.1.0",
		"description": "Comprehensive compression and decompression module supporting multiple algorithms",
		"author":      "Ben",
		"language":    "Go",
		"target":      "WebAssembly",
		"functions":   19,
		"categories": []string{
			"Gzip Compression",
			"Deflate Compression",
			"LZ4 Compression",
			"Snappy Compression",
			"TAR Archive",
			"ZIP Archive",
			"Compression Analysis",
		},
		"features": []string{
			"Multi-algorithm compression and decompression",
			"Configurable compression levels",
			"TAR and ZIP archive creation and extraction",
			"Compression ratio analysis across algorithms",
			"Shannon entropy calculation",
			"Compressibility classification",
			"Data type detection",
			"Base64 encoded I/O for WASM compatibility",
		},
		"buildInfo": map[string]interface{}{
			"goVersion":    "1.21+",
			"dependencies": []string{},
			"optimized":    true,
			"compressed":   true,
		},
	}

	if !silentMode {
		fmt.Printf("Go WASM: Module info retrieved for compress-wasm v0.1.0\n")
	}

	// Convert to JSON and back to avoid nested map issues with js.ValueOf
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Sprintf("Error: Failed to marshal module info: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return fmt.Sprintf("Error: Failed to unmarshal module info: %v", err)
	}

	return js.ValueOf(result)
}

// getAvailableFunctions returns list of all available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		// Gzip
		"gzipCompress", "gzipDecompress",
		// Deflate
		"deflateCompress", "deflateDecompress",
		// LZ4
		"lz4Compress", "lz4Decompress",
		// Snappy
		"snappyCompress", "snappyDecompress",
		// TAR
		"tarCreate", "tarExtract", "tarList",
		// ZIP
		"zipCreate", "zipExtract", "zipList",
		// Analysis
		"analyzeCompression", "estimateCompression",
		// System
		"setSilentMode", "getAvailableFunctions", "getModuleInfo",
	}

	if !silentMode {
		fmt.Printf("Go WASM: Available functions: %d\n", len(functions))
	}

	return js.ValueOf(functions)
}

func main() {
	c := make(chan struct{})

	fmt.Println("Go WASM Compress Module initializing...")

	// Register Gzip functions
	js.Global().Set("gzipCompress", js.FuncOf(gzipCompress))
	js.Global().Set("gzipDecompress", js.FuncOf(gzipDecompress))

	// Register Deflate functions
	js.Global().Set("deflateCompress", js.FuncOf(deflateCompress))
	js.Global().Set("deflateDecompress", js.FuncOf(deflateDecompress))

	// Register LZ4 functions
	js.Global().Set("lz4Compress", js.FuncOf(lz4Compress))
	js.Global().Set("lz4Decompress", js.FuncOf(lz4Decompress))

	// Register Snappy functions
	js.Global().Set("snappyCompress", js.FuncOf(snappyCompress))
	js.Global().Set("snappyDecompress", js.FuncOf(snappyDecompress))

	// Register TAR functions
	js.Global().Set("tarCreate", js.FuncOf(tarCreate))
	js.Global().Set("tarExtract", js.FuncOf(tarExtract))
	js.Global().Set("tarList", js.FuncOf(tarList))

	// Register ZIP functions
	js.Global().Set("zipCreate", js.FuncOf(zipCreate))
	js.Global().Set("zipExtract", js.FuncOf(zipExtract))
	js.Global().Set("zipList", js.FuncOf(zipList))

	// Register Analysis functions
	js.Global().Set("analyzeCompression", js.FuncOf(analyzeCompression))
	js.Global().Set("estimateCompression", js.FuncOf(estimateCompression))

	// Register System functions
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("getModuleInfo", js.FuncOf(getModuleInfo))

	// Signal readiness for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Go WASM Compress Module ready!")
	fmt.Println("Available: Gzip, Deflate, LZ4, Snappy, TAR, ZIP, Compression Analysis")

	<-c
}
