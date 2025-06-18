package tester

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Tester handles module testing
type Tester struct {
	config *Config
}

// Config holds tester configuration
type Config struct {
	Integration bool
	Coverage    bool
	Verbose     bool
	Workers     int
}

// TestResult represents the result of testing a module
type TestResult struct {
	Module string          `json:"module"`
	Passed bool            `json:"passed"`
	Errors []string        `json:"errors,omitempty"`
	Tests  map[string]bool `json:"tests"`
}

// New creates a new Tester instance
func New(cfg *Config) *Tester {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Tester{config: cfg}
}

// TestModules tests multiple modules
func (t *Tester) TestModules(modules []string) ([]*TestResult, error) {
	if len(modules) == 0 {
		// Discover all modules
		discoveredModules, err := t.discoverModules(".")
		if err != nil {
			return nil, fmt.Errorf("failed to discover modules: %w", err)
		}
		modules = discoveredModules
	}

	results := make([]*TestResult, len(modules))

	for i, module := range modules {
		results[i] = t.testModule(module)
	}

	return results, nil
}

// testModule tests a single module
func (t *Tester) testModule(module string) *TestResult {
	result := &TestResult{
		Module: module,
		Tests:  make(map[string]bool),
	}

	modulePath := filepath.Join(".", module)

	// Check if module directory exists
	if !t.dirExists(modulePath) {
		result.Errors = append(result.Errors, fmt.Sprintf("module directory %s not found", modulePath))
		return result
	}

	// Test getAvailableFunctions implementation
	t.testGetAvailableFunctions(modulePath, result)

	// Test setSilentMode implementation
	t.testSetSilentMode(modulePath, result)

	// Test function registration
	t.testFunctionRegistration(modulePath, result)

	// Test module.json documentation
	t.testModuleJsonDocumentation(modulePath, result)

	// Determine if all tests passed
	result.Passed = len(result.Errors) == 0

	return result
}

// testGetAvailableFunctions checks if getAvailableFunctions is implemented
func (t *Tester) testGetAvailableFunctions(modulePath string, result *TestResult) {
	mainGoPath := filepath.Join(modulePath, "main.go")
	if !t.fileExists(mainGoPath) {
		result.Errors = append(result.Errors, "main.go not found")
		return
	}

	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read main.go: %v", err))
		return
	}

	source := string(content)

	// Check if function exists
	pattern := `func\s+getAvailableFunctions\s*\(`
	matched, _ := regexp.MatchString(pattern, source)
	result.Tests["getAvailableFunctions_exists"] = matched

	if !matched {
		result.Errors = append(result.Errors, "getAvailableFunctions function not found")
		return
	}

	// Check if it returns proper format
	// This is a basic check - could be enhanced with AST parsing
	if strings.Contains(source, "getAvailableFunctions") {
		result.Tests["getAvailableFunctions_implemented"] = true
	}
}

// testSetSilentMode checks if setSilentMode is implemented
func (t *Tester) testSetSilentMode(modulePath string, result *TestResult) {
	mainGoPath := filepath.Join(modulePath, "main.go")
	if !t.fileExists(mainGoPath) {
		return // Already checked in getAvailableFunctions
	}

	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		return // Already handled
	}

	source := string(content)

	// Check if function exists
	pattern := `func\s+setSilentMode\s*\(`
	matched, _ := regexp.MatchString(pattern, source)
	result.Tests["setSilentMode_exists"] = matched

	if !matched {
		result.Errors = append(result.Errors, "setSilentMode function not found")
	}
}

// testFunctionRegistration checks if functions are properly registered
func (t *Tester) testFunctionRegistration(modulePath string, result *TestResult) {
	mainGoPath := filepath.Join(modulePath, "main.go")
	if !t.fileExists(mainGoPath) {
		return
	}

	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		return
	}

	source := string(content)

	functions := []string{"getAvailableFunctions", "setSilentMode"}

	for _, fn := range functions {
		pattern := fmt.Sprintf(`js\.FuncOf\(%s\)`, fn)
		matched, _ := regexp.MatchString(pattern, source)
		result.Tests[fn+"_registered"] = matched

		if !matched {
			result.Errors = append(result.Errors, fmt.Sprintf("%s not properly registered", fn))
		}
	}
}

// testModuleJsonDocumentation checks if functions are documented in module.json
func (t *Tester) testModuleJsonDocumentation(modulePath string, result *TestResult) {
	moduleJsonPath := filepath.Join(modulePath, "module.json")
	if !t.fileExists(moduleJsonPath) {
		result.Errors = append(result.Errors, "module.json not found")
		return
	}

	content, err := os.ReadFile(moduleJsonPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read module.json: %v", err))
		return
	}

	source := string(content)

	// Check if getAvailableFunctions is documented
	if strings.Contains(source, `"name": "getAvailableFunctions"`) {
		result.Tests["getAvailableFunctions_documented"] = true
	} else {
		result.Errors = append(result.Errors, "getAvailableFunctions not documented in module.json")
	}
}

// discoverModules finds all WASM modules
func (t *Tester) discoverModules(rootDir string) ([]string, error) {
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

		if t.fileExists(mainGoPath) && t.fileExists(goModPath) {
			modules = append(modules, entry.Name())
		}
	}

	return modules, nil
}

// PrintTestSummary prints test results summary
func PrintTestSummary(results []*TestResult) (passed, total int) {
	total = len(results)

	fmt.Println("\n🧪 Test Summary")
	fmt.Println("===============")

	for _, result := range results {
		if result.Passed {
			passed++
			fmt.Printf("✅ %-15s all tests passed\n", result.Module)
		} else {
			fmt.Printf("❌ %-15s %d test(s) failed\n", result.Module, len(result.Errors))

			for _, err := range result.Errors {
				fmt.Printf("   • %s\n", err)
			}
		}
	}

	fmt.Printf("\n📊 Results: %d/%d modules passed\n", passed, total)

	return passed, total
}

// Helper functions
func (t *Tester) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (t *Tester) dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
