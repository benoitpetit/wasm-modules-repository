package builder

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"wasm-manager/internal/config"

	"golang.org/x/sync/errgroup"
)

// Builder handles WASM module building with parallel processing
type Builder struct {
	config *config.BuildConfig
}

// BuildResult represents the result of building a module
type BuildResult struct {
	Module         string        `json:"module"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
	BuildTime      time.Duration `json:"buildTime"`
	OriginalSize   int64         `json:"originalSize"`
	OptimizedSize  int64         `json:"optimizedSize"`
	CompressedSize int64         `json:"compressedSize"`
	Integrity      string        `json:"integrity,omitempty"`
}

// New creates a new Builder instance
func New(cfg *config.BuildConfig) *Builder {
	if cfg == nil {
		cfg = config.DefaultBuildConfig()
	}
	return &Builder{
		config: cfg,
	}
}

// DiscoverModules finds all WASM modules in the given directory
func DiscoverModules(rootDir string) ([]string, error) {
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

		// Check if it's a WASM module (has main.go and go.mod)
		mainGoPath := filepath.Join(modulePath, "main.go")
		goModPath := filepath.Join(modulePath, "go.mod")

		if fileExists(mainGoPath) && fileExists(goModPath) {
			modules = append(modules, entry.Name())
		}
	}

	return modules, nil
}

// BuildModules builds multiple modules in parallel
func (b *Builder) BuildModules(modules []string) ([]*BuildResult, error) {
	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules to build")
	}

	// Limit workers to avoid overwhelming the system
	maxWorkers := b.config.Workers
	if maxWorkers > len(modules) {
		maxWorkers = len(modules)
	}

	results := make([]*BuildResult, len(modules))
	resultsMu := sync.Mutex{}

	// Use errgroup for worker management
	g := new(errgroup.Group)
	g.SetLimit(maxWorkers)

	for i, module := range modules {
		i, module := i, module // capture loop variables
		g.Go(func() error {
			result := b.buildModule(module)

			resultsMu.Lock()
			results[i] = result
			resultsMu.Unlock()

			if b.config.Verbose {
				if result.Success {
					fmt.Printf("✅ %s built successfully in %v\n", module, result.BuildTime)
				} else {
					fmt.Printf("❌ %s build failed: %s\n", module, result.Error)
				}
			}

			return nil // Don't stop other builds if one fails
		})
	}

	// Wait for all builds to complete
	if err := g.Wait(); err != nil {
		return results, err
	}

	return results, nil
}

// buildModule builds a single WASM module
func (b *Builder) buildModule(module string) *BuildResult {
	startTime := time.Now()
	result := &BuildResult{
		Module: module,
	}

	// Clean first if requested
	if b.config.Clean {
		if err := b.cleanModule(module); err != nil {
			result.Error = fmt.Sprintf("clean failed: %v", err)
			return result
		}
	}

	// Check if module directory exists
	modulePath := filepath.Join(".", module)
	if !dirExists(modulePath) {
		result.Error = fmt.Sprintf("module directory %s not found", modulePath)
		return result
	}

	// Build the WASM module
	wasmPath := filepath.Join(modulePath, "main.wasm")
	if err := b.compileWasm(modulePath, wasmPath); err != nil {
		result.Error = fmt.Sprintf("compilation failed: %v", err)
		return result
	}

	// Move WASM from subdirectory to root if Go created a subdirectory
	b.moveWasmFromSubdir(modulePath)

	// Get file size
	if stat, err := os.Stat(wasmPath); err == nil {
		result.OriginalSize = stat.Size()
		result.OptimizedSize = stat.Size()
	}

	// Optimize if enabled
	if b.config.Optimize {
		if err := b.optimizeWasm(wasmPath); err != nil {
			if b.config.Verbose {
				fmt.Printf("⚠️ Optimization failed for %s: %v\n", module, err)
			}
		} else {
			// Update optimized size
			if stat, err := os.Stat(wasmPath); err == nil {
				result.OptimizedSize = stat.Size()
			}
		}
	}

	// Compress if enabled
	if b.config.Compress {
		if err := b.compressWasm(wasmPath); err != nil {
			if b.config.Verbose {
				fmt.Printf("⚠️ Compression failed for %s: %v\n", module, err)
			}
		} else {
			// Get compressed size
			gzipPath := wasmPath + ".gz"
			if stat, err := os.Stat(gzipPath); err == nil {
				result.CompressedSize = stat.Size()
			}
		}
	}

	// Generate integrity hash if enabled
	if b.config.GenerateIntegrity {
		integrity, err := b.generateIntegrity(wasmPath)
		if err != nil {
			if b.config.Verbose {
				fmt.Printf("⚠️ Integrity generation failed for %s: %v\n", module, err)
			}
		} else {
			result.Integrity = integrity
		}
	}

	result.Success = true
	result.BuildTime = time.Since(startTime)

	return result
}

// compileWasm compiles Go source to WASM
func (b *Builder) compileWasm(modulePath, outputPath string) error {
	cmd := exec.Command("go", "build",
		"-ldflags", "-s -w -buildid=",
		"-trimpath",
		"-buildmode=default",
		"-tags", "netgo,osusergo",
		"-a",
		"-gcflags", "-l=4 -B",
		"-o", outputPath,
		"main.go")

	cmd.Dir = modulePath
	cmd.Env = append(os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
		"CGO_ENABLED=0",
	)

	if b.config.Verbose {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("go build failed: %w\nOutput: %s", err, output)
		}
	} else {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build failed: %w", err)
		}
	}

	return nil
}

// optimizeWasm optimizes WASM file using wasm-opt
func (b *Builder) optimizeWasm(wasmPath string) error {
	// Check if wasm-opt is available
	if _, err := exec.LookPath("wasm-opt"); err != nil {
		return fmt.Errorf("wasm-opt not found: %w", err)
	}

	// Verify input file exists
	if !fileExists(wasmPath) {
		return fmt.Errorf("input WASM file does not exist: %s", wasmPath)
	}

	// Single-pass conservative optimization to avoid failures
	outputPath := wasmPath + ".opt"
	args := []string{
		"-Oz",
		"--enable-bulk-memory",
		"--enable-sign-ext",
		"--enable-mutable-globals",
		"--enable-nontrapping-float-to-int",
		wasmPath,
		"-o", outputPath,
	}

	cmd := exec.Command("wasm-opt", args...)
	if err := cmd.Run(); err != nil {
		// If optimization fails, keep original
		return fmt.Errorf("optimization failed: %w", err)
	}

	// Verify optimized file was created
	if !fileExists(outputPath) {
		return fmt.Errorf("optimization did not produce output file: %s", outputPath)
	}

	// Move optimized result to original location
	if err := os.Rename(outputPath, wasmPath); err != nil {
		// Cleanup on failure
		os.Remove(outputPath)
		return fmt.Errorf("failed to move optimized file: %w", err)
	}

	return nil
}

// compressWasm creates compressed versions of WASM file
func (b *Builder) compressWasm(wasmPath string) error {
	// Create gzip version
	cmd := exec.Command("gzip", "-9", "-f", "-k", wasmPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gzip compression failed: %w", err)
	}

	// Create brotli version if available
	if _, err := exec.LookPath("brotli"); err == nil {
		cmd := exec.Command("brotli", "-f", "-Z", wasmPath)
		if err := cmd.Run(); err != nil {
			// Brotli failure is not critical
			if b.config.Verbose {
				fmt.Printf("⚠️ Brotli compression failed: %v\n", err)
			}
		}
	}

	return nil
}

// generateIntegrity generates SHA256 integrity hash
func (b *Builder) generateIntegrity(wasmPath string) (string, error) {
	file, err := os.Open(wasmPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	hash := hasher.Sum(nil)
	integrity := "sha256-" + base64.StdEncoding.EncodeToString(hash)

	// Write to integrity file
	integrityPath := wasmPath + ".integrity"
	if err := os.WriteFile(integrityPath, []byte(integrity), 0644); err != nil {
		return "", fmt.Errorf("failed to write integrity file: %w", err)
	}

	return integrity, nil
}

// cleanModule removes only temporary build artifacts from a module
func (b *Builder) cleanModule(module string) error {
	modulePath := filepath.Join(".", module)
	patterns := []string{
		"*.backup",
		"*.wasm.opt",
		"*.wasm.pass*",
		"*.wasm.br", // Only brotli, keep gzip
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(modulePath, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			os.Remove(match)
		}
	}

	// Move WASM files from subdirectory to root and clean up subdirectory
	moduleName := filepath.Base(modulePath)
	subdirPath := filepath.Join(modulePath, moduleName)
	if dirExists(subdirPath) {
		// Move main.wasm from subdirectory to root if it exists
		srcWasm := filepath.Join(subdirPath, "main.wasm")
		dstWasm := filepath.Join(modulePath, "main.wasm")
		if fileExists(srcWasm) {
			os.Rename(srcWasm, dstWasm)
		}

		// Remove the now-empty subdirectory
		os.RemoveAll(subdirPath)
	}

	return nil
}

// moveWasmFromSubdir moves WASM files from subdirectory to module root
func (b *Builder) moveWasmFromSubdir(modulePath string) {
	moduleName := filepath.Base(modulePath)
	subdirPath := filepath.Join(modulePath, moduleName)

	if dirExists(subdirPath) {
		// Move main.wasm from subdirectory to root if it exists
		srcWasm := filepath.Join(subdirPath, "main.wasm")
		dstWasm := filepath.Join(modulePath, "main.wasm")
		if fileExists(srcWasm) {
			os.Rename(srcWasm, dstWasm)
		}

		// Remove the now-empty subdirectory
		os.RemoveAll(subdirPath)
	}
}

// PrintBuildSummary prints a summary of build results
func PrintBuildSummary(results []*BuildResult) {
	var successful, failed int
	var totalTime time.Duration
	var totalOriginalSize, totalOptimizedSize, totalCompressedSize int64

	fmt.Println("\n📋 Build Summary")
	fmt.Println("================")

	for _, result := range results {
		if result.Success {
			successful++
			fmt.Printf("✅ %-15s %8s → %8s",
				result.Module,
				formatBytes(result.OriginalSize),
				formatBytes(result.OptimizedSize))

			if result.CompressedSize > 0 {
				fmt.Printf(" → %8s", formatBytes(result.CompressedSize))
			}

			fmt.Printf(" (%v)\n", result.BuildTime)

			totalOriginalSize += result.OriginalSize
			totalOptimizedSize += result.OptimizedSize
			totalCompressedSize += result.CompressedSize
		} else {
			failed++
			fmt.Printf("❌ %-15s %s\n", result.Module, result.Error)
		}
		totalTime += result.BuildTime
	}

	fmt.Printf("\n📊 Statistics:\n")
	fmt.Printf("   Successful: %d\n", successful)
	fmt.Printf("   Failed: %d\n", failed)
	fmt.Printf("   Total time: %v\n", totalTime)

	if totalOriginalSize > 0 {
		reduction := totalOriginalSize - totalOptimizedSize
		reductionPercent := (reduction * 100) / totalOriginalSize
		fmt.Printf("   Size reduction: %s (%.1f%%)\n", formatBytes(reduction), float64(reductionPercent))

		if totalCompressedSize > 0 {
			compressionRatio := (totalCompressedSize * 100) / totalOriginalSize
			fmt.Printf("   Compression ratio: %.1f%%\n", float64(compressionRatio))
		}
	}
}

// Helper functions
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
