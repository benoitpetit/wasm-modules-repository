package cmd

import (
	"fmt"

	"wasm-manager/internal/tester"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [module]",
	Short: "Test function implementations",
	Long: `Test WASM module function implementations and compatibility.

Tests:
• getAvailableFunctions implementation
• setSilentMode functionality
• Function registration in main()
• Module.json documentation
• WASM binary functionality (if built)

Examples:
  wasm-manager test                     # Test all modules
  wasm-manager test math-wasm           # Test specific module
  wasm-manager test --integration       # Run integration tests`,
	RunE: runTest,
}

var (
	testIntegration bool
	testCoverage    bool
)

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().BoolVar(&testIntegration, "integration", false, "run integration tests")
	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, "generate test coverage report")
}

func runTest(cmd *cobra.Command, args []string) error {
	cfg := &tester.Config{
		Integration: testIntegration,
		Coverage:    testCoverage,
		Verbose:     verbose,
		Workers:     getWorkerCount(),
	}

	var targetModules []string
	if len(args) > 0 {
		targetModules = args
	}

	t := tester.New(cfg)
	results, err := t.TestModules(targetModules)
	if err != nil {
		return fmt.Errorf("testing failed: %w", err)
	}

	// Print test summary
	passed, total := tester.PrintTestSummary(results)

	if passed == total {
		fmt.Println("🎉 All tests passed!")
		return nil
	} else {
		return fmt.Errorf("tests failed: %d/%d modules failed", total-passed, total)
	}
}
