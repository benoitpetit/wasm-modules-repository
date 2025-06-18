package cmd

import (
	"fmt"

	"wasm-manager/internal/cleaner"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean [module]",
	Short: "Clean build artifacts",
	Long: `Clean build artifacts and temporary files.

Removes:
• WASM binaries (*.wasm)
• Compressed files (*.wasm.gz, *.wasm.br)
• Integrity files (*.wasm.integrity)
• Backup files (*.backup)
• Temporary build files

Examples:
  wasm-manager clean                    # Clean all modules
  wasm-manager clean math-wasm          # Clean specific module
  wasm-manager clean --all              # Clean everything including caches`,
	RunE: runClean,
}

var (
	cleanAll   bool
	cleanCache bool
)

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "clean all artifacts including caches")
	cleanCmd.Flags().BoolVar(&cleanCache, "cache", false, "clean build caches only")
}

func runClean(cmd *cobra.Command, args []string) error {
	cfg := &cleaner.Config{
		All:     cleanAll,
		Cache:   cleanCache,
		Verbose: verbose,
	}

	var targetModules []string
	if len(args) > 0 {
		targetModules = args
	}

	c := cleaner.New(cfg)
	cleaned, err := c.CleanModules(targetModules)
	if err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	fmt.Printf("🧹 Cleaned %d modules\n", cleaned)
	return nil
}
