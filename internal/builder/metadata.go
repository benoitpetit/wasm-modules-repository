package builder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ModuleMetadata represents the structure of module.json files
type ModuleMetadata struct {
	Name          string                 `json:"name"`
	Author        string                 `json:"author"`
	Version       string                 `json:"version"`
	Description   string                 `json:"description"`
	BuildTime     int64                  `json:"buildTime"`
	Size          int64                  `json:"size"`
	GzipSize      int64                  `json:"gzipSize"`
	License       string                 `json:"license"`
	Tags          []string               `json:"tags"`
	Changelog     map[string]interface{} `json:"changelog"`
	BuildInfo     map[string]interface{} `json:"buildInfo"`
	Compatibility map[string]interface{} `json:"compatibility"`
	Functions     []interface{}          `json:"functions"`
	FileInfo      map[string]interface{} `json:"fileInfo"`
}

// ValidateModuleMetadata validates a module's metadata and returns issues
func ValidateModuleMetadata(module string) (bool, []string) {
	var issues []string

	modulePath := filepath.Join(".", module)
	metadataPath := filepath.Join(modulePath, "module.json")

	// Check if module.json exists
	if !fileExists(metadataPath) {
		issues = append(issues, "module.json file not found")
		return false, issues
	}

	// Parse metadata
	metadata, err := parseModuleMetadata(metadataPath)
	if err != nil {
		issues = append(issues, fmt.Sprintf("failed to parse module.json: %v", err))
		return false, issues
	}

	// Validate required fields
	if metadata.Name == "" {
		issues = append(issues, "missing required field: name")
	}

	if metadata.Author == "" {
		issues = append(issues, "missing required field: author")
	}

	if metadata.Version == "" {
		issues = append(issues, "missing required field: version")
	}

	if metadata.Description == "" {
		issues = append(issues, "missing required field: description")
	}

	// Check if WASM file exists and validate file info
	wasmPath := filepath.Join(modulePath, "main.wasm")
	if fileExists(wasmPath) {
		if stat, err := os.Stat(wasmPath); err == nil {
			actualSize := stat.Size()
			if metadata.Size > 0 && metadata.Size != actualSize {
				issues = append(issues, fmt.Sprintf("size mismatch: reported %d, actual %d", metadata.Size, actualSize))
			}
		}
	} else {
		issues = append(issues, "main.wasm file not found")
	}

	// Check if gzip file exists and validate gzip size
	gzipPath := filepath.Join(modulePath, "main.wasm.gz")
	if fileExists(gzipPath) {
		if stat, err := os.Stat(gzipPath); err == nil {
			actualGzipSize := stat.Size()
			if metadata.GzipSize > 0 && metadata.GzipSize != actualGzipSize {
				issues = append(issues, fmt.Sprintf("gzip size mismatch: reported %d, actual %d", metadata.GzipSize, actualGzipSize))
			}
		}
	}

	// Validate source line count if provided
	if metadata.FileInfo != nil {
		if sourceLines, ok := metadata.FileInfo["sourceLines"]; ok {
			if lines, ok := sourceLines.(float64); ok {
				actualLines := countSourceLines(filepath.Join(modulePath, "main.go"))
				if actualLines > 0 && int(lines) != actualLines {
					issues = append(issues, fmt.Sprintf("source lines mismatch: reported %d, actual %d", int(lines), actualLines))
				}
			}
		}
	}

	// Validate build time consistency
	if metadata.BuildTime > 0 {
		buildTime := time.Unix(metadata.BuildTime, 0)
		if buildTime.After(time.Now()) {
			issues = append(issues, "build time is in the future")
		}
	}

	// Check for build info completeness
	if metadata.BuildInfo == nil {
		issues = append(issues, "missing build information")
	} else {
		requiredBuildFields := []string{"goVersion", "buildTime", "target", "outputFile"}
		for _, field := range requiredBuildFields {
			if _, exists := metadata.BuildInfo[field]; !exists {
				issues = append(issues, fmt.Sprintf("missing build info field: %s", field))
			}
		}
	}

	return len(issues) == 0, issues
}

// GenerateMetadataReport generates a detailed report for multiple modules
func GenerateMetadataReport(modules []string) {
	fmt.Println("\n📊 Metadata Report")
	fmt.Println("==================")

	totalModules := len(modules)
	validModules := 0
	totalIssues := 0

	for _, module := range modules {
		fmt.Printf("\n🔍 Module: %s\n", module)
		fmt.Println(strings.Repeat("-", len(module)+11))

		valid, issues := ValidateModuleMetadata(module)

		if valid {
			validModules++
			fmt.Printf("✅ Status: Valid\n")

			// Show metadata summary
			showMetadataSummary(module)
		} else {
			totalIssues += len(issues)
			fmt.Printf("❌ Status: %d issues found\n", len(issues))

			for i, issue := range issues {
				fmt.Printf("   %d. %s\n", i+1, issue)
			}
		}
	}

	// Summary
	fmt.Printf("\n📋 Summary\n")
	fmt.Printf("==========\n")
	fmt.Printf("Total modules: %d\n", totalModules)
	fmt.Printf("Valid modules: %d\n", validModules)
	fmt.Printf("Invalid modules: %d\n", totalModules-validModules)
	fmt.Printf("Total issues: %d\n", totalIssues)

	if validModules == totalModules {
		fmt.Printf("\n🎉 All modules have valid metadata!\n")
	} else {
		fmt.Printf("\n⚠️  %d modules need attention\n", totalModules-validModules)
	}
}

// showMetadataSummary displays a summary of module metadata
func showMetadataSummary(module string) {
	metadataPath := filepath.Join(".", module, "module.json")
	metadata, err := parseModuleMetadata(metadataPath)
	if err != nil {
		fmt.Printf("   Error reading metadata: %v\n", err)
		return
	}

	fmt.Printf("   Name: %s\n", metadata.Name)
	fmt.Printf("   Author: %s\n", metadata.Author)
	fmt.Printf("   Version: %s\n", metadata.Version)
	fmt.Printf("   Description: %s\n", truncateString(metadata.Description, 80))

	if metadata.Size > 0 {
		fmt.Printf("   WASM Size: %s\n", formatBytes(metadata.Size))
	}

	if metadata.GzipSize > 0 {
		fmt.Printf("   Compressed: %s\n", formatBytes(metadata.GzipSize))
		if metadata.Size > 0 {
			ratio := (float64(metadata.GzipSize) / float64(metadata.Size)) * 100
			fmt.Printf("   Compression: %.1f%%\n", ratio)
		}
	}

	if metadata.BuildTime > 0 {
		buildTime := time.Unix(metadata.BuildTime, 0)
		fmt.Printf("   Last Build: %s\n", buildTime.Format("2006-01-02 15:04:05"))
	}

	if len(metadata.Tags) > 0 {
		fmt.Printf("   Tags: %s\n", strings.Join(metadata.Tags, ", "))
	}

	if len(metadata.Functions) > 0 {
		fmt.Printf("   Functions: %d\n", len(metadata.Functions))
	}
}

// parseModuleMetadata parses module.json file
func parseModuleMetadata(path string) (*ModuleMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var metadata ModuleMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &metadata, nil
}

// countSourceLines counts the number of lines in main.go
func countSourceLines(path string) int {
	if !fileExists(path) {
		return 0
	}

	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Count non-empty lines that are not just comments
		if line != "" && !strings.HasPrefix(line, "//") {
			lineCount++
		}
	}

	return lineCount
}

// truncateString truncates a string to the specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
