package cmd

import (
	"fmt"

	"wasm-manager/internal/validator"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [module]",
	Short: "Validate module structure and requirements",
	Long: `Validate WASM modules for compliance with project standards.

Checks:
• Required files (main.go, module.json, go.mod)
• Function implementations (getAvailableFunctions, setSilentMode)
• Module.json structure and content
• Go code conventions
• Build artifacts integrity

Examples:
  wasm-manager validate                 # Validate all modules
  wasm-manager validate math-wasm       # Validate specific module
  wasm-manager validate --strict        # Enable strict validation`,
	RunE: runValidate,
}

var (
	validateStrict bool
	validateFix    bool
)

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "enable strict validation rules")
	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "attempt to fix issues automatically")
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg := &validator.Config{
		Strict:  validateStrict,
		Fix:     validateFix,
		Verbose: verbose,
	}

	var targetModules []string
	if len(args) > 0 {
		targetModules = args
	}

	v := validator.New(cfg)
	results, err := v.ValidateModules(targetModules)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Print validation summary
	passed, total := validator.PrintValidationSummary(results)

	if passed == total {
		fmt.Println("🎉 All modules are compliant!")
		return nil
	} else {
		return fmt.Errorf("validation failed: %d/%d modules have issues", total-passed, total)
	}
}
