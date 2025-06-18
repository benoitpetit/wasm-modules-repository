package cmd

import (
	"fmt"

	"wasm-manager/internal/builder"
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

var validateMetadataCmd = &cobra.Command{
	Use:   "metadata [module]",
	Short: "Validate module metadata and build information",
	Long: `Validate that module.json files contain accurate and up-to-date metadata
including build information, file sizes, line counts, and other important fields.

Examples:
  wasm-manager validate metadata                 # Validate all modules
  wasm-manager validate metadata math-wasm      # Validate specific module
  wasm-manager validate metadata --report       # Generate detailed report`,
	RunE: runValidateMetadata,
}

var (
	validateStrict bool
	validateFix    bool
	metadataReport bool
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.AddCommand(validateMetadataCmd)

	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "enable strict validation rules")
	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "attempt to fix issues automatically")
	validateMetadataCmd.Flags().BoolVar(&metadataReport, "report", false, "generate detailed metadata report")
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

func runValidateMetadata(cmd *cobra.Command, args []string) error {
	// Determine which modules to validate
	var targetModules []string
	if len(args) > 0 {
		targetModules = args
	} else {
		// Validate all modules
		modules, err := builder.DiscoverModules(".")
		if err != nil {
			return fmt.Errorf("failed to discover modules: %w", err)
		}
		targetModules = modules
	}

	if len(targetModules) == 0 {
		return fmt.Errorf("no modules found to validate")
	}

	if metadataReport {
		builder.GenerateMetadataReport(targetModules)
	} else {
		// Quick validation
		allValid := true
		for _, module := range targetModules {
			valid, issues := builder.ValidateModuleMetadata(module)
			if valid {
				fmt.Printf("✅ %s - metadata is valid\n", module)
			} else {
				fmt.Printf("❌ %s - %d issues found\n", module, len(issues))
				if verbose {
					for _, issue := range issues {
						fmt.Printf("   • %s\n", issue)
					}
				}
				allValid = false
			}
		}

		if !allValid {
			fmt.Printf("\n💡 Use --report flag for detailed information\n")
			return fmt.Errorf("metadata validation failed for some modules")
		}
	}

	return nil
}
