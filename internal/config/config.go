package config

import (
	"time"
)

// BuildConfig holds configuration for build operations
type BuildConfig struct {
	Workers           int
	Optimize          bool
	Compress          bool
	GenerateIntegrity bool
	Clean             bool
	Verbose           bool
	Timeout           time.Duration
}

// DefaultBuildConfig returns default build configuration
func DefaultBuildConfig() *BuildConfig {
	return &BuildConfig{
		Workers:           4,
		Optimize:          true,
		Compress:          true,
		GenerateIntegrity: true,
		Clean:             false,
		Verbose:           false,
		Timeout:           10 * time.Minute,
	}
}

// ModuleInfo represents information about a WASM module
type ModuleInfo struct {
	Name        string         `json:"name"`
	Path        string         `json:"path"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Author      string         `json:"author"`
	License     string         `json:"license"`
	Tags        []string       `json:"tags"`
	Functions   []FunctionInfo `json:"functions"`
	BuildInfo   BuildInfo      `json:"buildInfo,omitempty"`
	Security    SecurityInfo   `json:"security,omitempty"`
}

// FunctionInfo represents a WASM function
type FunctionInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	ReturnType  string      `json:"returnType"`
	Example     string      `json:"example,omitempty"`
}

// Parameter represents a function parameter
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// BuildInfo contains build-related information
type BuildInfo struct {
	GoVersion string    `json:"goVersion,omitempty"`
	BuildTime time.Time `json:"buildTime,omitempty"`
	Size      int64     `json:"size,omitempty"`
	GzipSize  int64     `json:"gzipSize,omitempty"`
	Optimized bool      `json:"optimized,omitempty"`
	Integrity string    `json:"integrity,omitempty"`
}

// SecurityInfo contains security-related information
type SecurityInfo struct {
	SandboxMode    bool     `json:"sandboxMode"`
	AllowedDomains []string `json:"allowedDomains,omitempty"`
	CORS           bool     `json:"cors"`
	CSP            string   `json:"csp,omitempty"`
}
