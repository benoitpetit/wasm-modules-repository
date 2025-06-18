package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var workers int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wasm-manager",
	Short: "High-performance WASM modules build manager",
	Long: `🚀 WASM Projects Manager

A high-performance build system for WebAssembly modules written in Go.
Replaces all shell scripts with optimized Go implementation using goroutines
and workers for parallel processing.

Features:
• Parallel builds with worker pools
• Advanced WASM optimization
• Module validation and testing
• Dependency management
• Integrity verification
• Compression optimization`,
	Version: "2.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .wasm-manager.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 0, "number of worker goroutines (default: auto-detect)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("workers", rootCmd.PersistentFlags().Lookup("workers"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".wasm-manager" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wasm-manager")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
