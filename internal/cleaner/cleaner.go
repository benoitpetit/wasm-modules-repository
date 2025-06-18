package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
)

// Cleaner handles cleaning of build artifacts
type Cleaner struct {
	config *Config
}

// Config holds cleaner configuration
type Config struct {
	All     bool
	Cache   bool
	Verbose bool
}

// New creates a new Cleaner instance
func New(cfg *Config) *Cleaner {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Cleaner{config: cfg}
}

// CleanModules cleans build artifacts from modules
func (c *Cleaner) CleanModules(modules []string) (int, error) {
	if len(modules) == 0 {
		// Discover all modules
		discoveredModules, err := c.discoverModules(".")
		if err != nil {
			return 0, fmt.Errorf("failed to discover modules: %w", err)
		}
		modules = discoveredModules
	}

	cleaned := 0

	for _, module := range modules {
		if err := c.cleanModule(module); err != nil {
			if c.config.Verbose {
				fmt.Printf("⚠️ Failed to clean %s: %v\n", module, err)
			}
		} else {
			cleaned++
			if c.config.Verbose {
				fmt.Printf("🧹 Cleaned %s\n", module)
			}
		}
	}

	return cleaned, nil
}

// cleanModule cleans a single module
func (c *Cleaner) cleanModule(module string) error {
	modulePath := filepath.Join(".", module)

	if !c.dirExists(modulePath) {
		return fmt.Errorf("module directory %s not found", modulePath)
	}

	patterns := []string{
		"*.wasm",
		"*.wasm.gz",
		"*.wasm.br",
		"*.wasm.integrity",
		"*.backup",
	}

	if c.config.All {
		patterns = append(patterns,
			"*.tmp",
			"*.temp",
			".build",
		)
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(modulePath, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				if c.config.Verbose {
					fmt.Printf("⚠️ Failed to remove %s: %v\n", match, err)
				}
			} else if c.config.Verbose {
				fmt.Printf("🗑️  Removed %s\n", match)
			}
		}
	}

	return nil
}

// discoverModules finds all WASM modules
func (c *Cleaner) discoverModules(rootDir string) ([]string, error) {
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

		if c.fileExists(mainGoPath) && c.fileExists(goModPath) {
			modules = append(modules, entry.Name())
		}
	}

	return modules, nil
}

// Helper functions
func (c *Cleaner) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *Cleaner) dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
