package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"wasm-manager/internal/config"
)

// Validator handles module validation
type Validator struct {
	config *Config
}

// Config holds validator configuration
type Config struct {
	Strict  bool
	Fix     bool
	Verbose bool
}

// ValidationResult represents the result of validating a module
type ValidationResult struct {
	Module   string          `json:"module"`
	Valid    bool            `json:"valid"`
	Errors   []string        `json:"errors,omitempty"`
	Warnings []string        `json:"warnings,omitempty"`
	Checks   map[string]bool `json:"checks"`
}

// New creates a new Validator instance
func New(cfg *Config) *Validator {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Validator{config: cfg}
}

// ValidateModules validates multiple modules
func (v *Validator) ValidateModules(modules []string) ([]*ValidationResult, error) {
	if len(modules) == 0 {
		// Discover all modules
		discoveredModules, err := v.discoverModules(".")
		if err != nil {
			return nil, fmt.Errorf("failed to discover modules: %w", err)
		}
		modules = discoveredModules
	}

	results := make([]*ValidationResult, len(modules))

	for i, module := range modules {
		results[i] = v.validateModule(module)
	}

	return results, nil
}

// validateModule validates a single module
func (v *Validator) validateModule(module string) *ValidationResult {
	result := &ValidationResult{
		Module: module,
		Checks: make(map[string]bool),
	}

	modulePath := filepath.Join(".", module)

	// Check if module directory exists
	if !v.dirExists(modulePath) {
		result.Errors = append(result.Errors, fmt.Sprintf("module directory %s not found", modulePath))
		return result
	}

	// Check required files
	v.checkRequiredFiles(modulePath, result)

	// Check Go source structure
	v.checkGoSource(modulePath, result)

	// Check module.json structure
	v.checkModuleJson(modulePath, result)

	// Check go.mod
	v.checkGoMod(modulePath, result)

	// Check build artifacts if they exist
	v.checkBuildArtifacts(modulePath, result)

	// Determine if module is valid
	result.Valid = len(result.Errors) == 0
	if v.config.Strict {
		result.Valid = result.Valid && len(result.Warnings) == 0
	}

	return result
}

// checkRequiredFiles verifies that all required files exist
func (v *Validator) checkRequiredFiles(modulePath string, result *ValidationResult) {
	requiredFiles := []string{
		"main.go",
		"module.json",
		"go.mod",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(modulePath, file)
		if v.fileExists(filePath) {
			result.Checks[file] = true
		} else {
			result.Checks[file] = false
			result.Errors = append(result.Errors, fmt.Sprintf("required file %s not found", file))
		}
	}

	// Optional files
	optionalFiles := []string{
		"main.wasm",
		"main.wasm.gz",
		"main.wasm.integrity",
	}

	for _, file := range optionalFiles {
		filePath := filepath.Join(modulePath, file)
		if v.fileExists(filePath) {
			result.Checks[file] = true
		} else {
			result.Checks[file] = false
			if file == "main.wasm" {
				result.Warnings = append(result.Warnings, "WASM binary not built")
			}
		}
	}
}

// checkGoSource validates Go source code structure
func (v *Validator) checkGoSource(modulePath string, result *ValidationResult) {
	mainGoPath := filepath.Join(modulePath, "main.go")
	if !v.fileExists(mainGoPath) {
		return // Already checked in required files
	}

	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read main.go: %v", err))
		return
	}

	source := string(content)

	// Check for required functions
	requiredFunctions := []string{
		"getAvailableFunctions",
		"setSilentMode",
	}

	for _, fn := range requiredFunctions {
		pattern := fmt.Sprintf(`func\s+%s\s*\(`, fn)
		matched, _ := regexp.MatchString(pattern, source)
		result.Checks[fn] = matched

		if !matched {
			result.Errors = append(result.Errors, fmt.Sprintf("required function %s not found", fn))
		} else {
			// Check if function is registered
			regPattern := fmt.Sprintf(`js\.FuncOf\(%s\)`, fn)
			regMatched, _ := regexp.MatchString(regPattern, source)
			if !regMatched {
				result.Warnings = append(result.Warnings, fmt.Sprintf("function %s not registered in main()", fn))
			}
		}
	}

	// Check build constraints
	if !strings.Contains(source, "//go:build js && wasm") {
		result.Warnings = append(result.Warnings, "missing build constraint '//go:build js && wasm'")
	}

	// Check package declaration
	if !strings.Contains(source, "package main") {
		result.Errors = append(result.Errors, "missing 'package main' declaration")
	}

	// Check required imports
	requiredImports := []string{
		"syscall/js",
	}

	for _, imp := range requiredImports {
		if !strings.Contains(source, fmt.Sprintf(`"%s"`, imp)) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("missing import %s", imp))
		}
	}
}

// checkModuleJson validates module.json structure
func (v *Validator) checkModuleJson(modulePath string, result *ValidationResult) {
	moduleJsonPath := filepath.Join(modulePath, "module.json")
	if !v.fileExists(moduleJsonPath) {
		return // Already checked in required files
	}

	content, err := os.ReadFile(moduleJsonPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read module.json: %v", err))
		return
	}

	var moduleInfo config.ModuleInfo
	if err := json.Unmarshal(content, &moduleInfo); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid JSON in module.json: %v", err))
		return
	}

	// Check required fields
	if moduleInfo.Name == "" {
		result.Errors = append(result.Errors, "module.json missing 'name' field")
	}

	if moduleInfo.Description == "" {
		result.Errors = append(result.Errors, "module.json missing 'description' field")
	}

	if moduleInfo.Version == "" {
		result.Warnings = append(result.Warnings, "module.json missing 'version' field")
	}

	if len(moduleInfo.Functions) == 0 {
		result.Errors = append(result.Errors, "module.json missing 'functions' array")
	} else {
		// Check for getAvailableFunctions in functions array
		hasGetAvailableFunctions := false
		for _, fn := range moduleInfo.Functions {
			if fn.Name == "getAvailableFunctions" {
				hasGetAvailableFunctions = true
				break
			}
		}

		if !hasGetAvailableFunctions {
			result.Errors = append(result.Errors, "getAvailableFunctions not documented in module.json")
		}
	}

	result.Checks["module.json_valid"] = true
}

// checkGoMod validates go.mod file
func (v *Validator) checkGoMod(modulePath string, result *ValidationResult) {
	goModPath := filepath.Join(modulePath, "go.mod")
	if !v.fileExists(goModPath) {
		return // Already checked in required files
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read go.mod: %v", err))
		return
	}

	source := string(content)

	// Check module declaration
	if !strings.Contains(source, "module ") {
		result.Errors = append(result.Errors, "go.mod missing module declaration")
	}

	// Check Go version
	if !strings.Contains(source, "go ") {
		result.Warnings = append(result.Warnings, "go.mod missing Go version")
	}

	result.Checks["go.mod_valid"] = true
}

// checkBuildArtifacts validates build artifacts
func (v *Validator) checkBuildArtifacts(modulePath string, result *ValidationResult) {
	wasmPath := filepath.Join(modulePath, "main.wasm")
	if !v.fileExists(wasmPath) {
		return // Not built yet
	}

	// Check file size
	stat, err := os.Stat(wasmPath)
	if err != nil {
		result.Warnings = append(result.Warnings, "failed to stat WASM file")
		return
	}

	if stat.Size() == 0 {
		result.Errors = append(result.Errors, "WASM file is empty")
		return
	}

	// Check for very large files (might indicate missing optimization)
	if stat.Size() > 10*1024*1024 { // 10MB
		result.Warnings = append(result.Warnings, "WASM file is very large, consider optimization")
	}

	// Check for compressed version
	gzipPath := filepath.Join(modulePath, "main.wasm.gz")
	if !v.fileExists(gzipPath) {
		result.Warnings = append(result.Warnings, "compressed version not found")
	}

	// Check for integrity file
	integrityPath := filepath.Join(modulePath, "main.wasm.integrity")
	if !v.fileExists(integrityPath) {
		result.Warnings = append(result.Warnings, "integrity file not found")
	}

	result.Checks["build_artifacts"] = true
}

// discoverModules finds all WASM modules
func (v *Validator) discoverModules(rootDir string) ([]string, error) {
	var modules []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", rootDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modulePath := filepath.Join(rootDir, entry.Name())

		// Check if it's a WASM module
		mainGoPath := filepath.Join(modulePath, "main.go")
		goModPath := filepath.Join(modulePath, "go.mod")

		if v.fileExists(mainGoPath) && v.fileExists(goModPath) {
			modules = append(modules, entry.Name())
		}
	}

	return modules, nil
}

// PrintValidationSummary prints validation results summary
func PrintValidationSummary(results []*ValidationResult) (passed, total int) {
	total = len(results)

	fmt.Println("\n🔍 Validation Summary")
	fmt.Println("====================")

	for _, result := range results {
		if result.Valid {
			passed++
			fmt.Printf("✅ %-15s valid\n", result.Module)
		} else {
			fmt.Printf("❌ %-15s invalid (%d errors", result.Module, len(result.Errors))
			if len(result.Warnings) > 0 {
				fmt.Printf(", %d warnings", len(result.Warnings))
			}
			fmt.Println(")")

			for _, err := range result.Errors {
				fmt.Printf("   • %s\n", err)
			}

			for _, warning := range result.Warnings {
				fmt.Printf("   ⚠ %s\n", warning)
			}
		}
	}

	fmt.Printf("\n📊 Results: %d/%d modules valid\n", passed, total)

	return passed, total
}

// Helper functions
func (v *Validator) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (v *Validator) dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
