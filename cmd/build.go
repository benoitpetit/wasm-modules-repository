package cmd

import (
	"fmt"
	"runtime"

	"wasm-manager/internal/builder"
	"wasm-manager/internal/config"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [module]",
	Short: "Build WASM modules with optimizations",
	Long: `Build WebAssembly modules with advanced optimizations using parallel processing.

Examples:
  wasm-manager build                    # Build all modules
  wasm-manager build math-wasm          # Build specific module
  wasm-manager build --workers 8        # Use 8 workers
  wasm-manager build --no-optimize      # Skip optimizations`,
	RunE: runBuild,
}

var (
	buildOptimize  bool
	buildCompress  bool
	buildIntegrity bool
	buildClean     bool
	buildModules   []string
)

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolVar(&buildOptimize, "optimize", true, "enable WASM optimization")
	buildCmd.Flags().BoolVar(&buildCompress, "compress", true, "create compressed versions")
	buildCmd.Flags().BoolVar(&buildIntegrity, "integrity", true, "generate integrity hashes")
	buildCmd.Flags().BoolVar(&buildClean, "clean", false, "clean before build")
	buildCmd.Flags().StringSliceVar(&buildModules, "modules", []string{}, "specific modules to build")
}

func runBuild(cmd *cobra.Command, args []string) error {
	cfg := &config.BuildConfig{
		Workers:           getWorkerCount(),
		Optimize:          buildOptimize,
		Compress:          buildCompress,
		GenerateIntegrity: buildIntegrity,
		Clean:             buildClean,
		Verbose:           verbose,
	}

	// Determine which modules to build
	var targetModules []string
	if len(args) > 0 {
		targetModules = args
	} else if len(buildModules) > 0 {
		targetModules = buildModules
	} else {
		// Build all modules
		modules, err := builder.DiscoverModules(".")
		if err != nil {
			return fmt.Errorf("failed to discover modules: %w", err)
		}
		targetModules = modules
	}

	if len(targetModules) == 0 {
		return fmt.Errorf("no modules found to build")
	}

	fmt.Printf("🚀 Building %d modules with %d workers\n", len(targetModules), cfg.Workers)

	b := builder.New(cfg)
	results, err := b.BuildModules(targetModules)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Print build summary
	builder.PrintBuildSummary(results)

	return nil
}

func getWorkerCount() int {
	if workers > 0 {
		return workers
	}
	// Auto-detect: number of CPU cores
	return runtime.NumCPU()
}
