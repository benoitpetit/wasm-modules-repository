package cmd

import (
	"wasm-manager/internal/installer"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install-tools",
	Short: "Install WASM optimization tools",
	Long: `Install and manage WASM optimization tools and dependencies.

Tools:
• Binaryen (wasm-opt) - WASM optimization
• WABT (WebAssembly Binary Toolkit) - WASM utilities
• Compression tools (gzip, brotli)
• Development utilities

Examples:
  wasm-manager install-tools            # Install all tools
  wasm-manager install-tools --check    # Check installations
  wasm-manager install-tools --binaryen # Install only Binaryen`,
	RunE: runInstall,
}

var (
	installCheck    bool
	installBinaryen bool
	installWABT     bool
	installForce    bool
)

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVar(&installCheck, "check", false, "check existing installations")
	installCmd.Flags().BoolVar(&installBinaryen, "binaryen", false, "install only Binaryen")
	installCmd.Flags().BoolVar(&installWABT, "wabt", false, "install only WABT")
	installCmd.Flags().BoolVar(&installForce, "force", false, "force reinstallation")
}

func runInstall(cmd *cobra.Command, args []string) error {
	cfg := &installer.Config{
		CheckOnly:    installCheck,
		BinaryenOnly: installBinaryen,
		WABTOnly:     installWABT,
		Force:        installForce,
		Verbose:      verbose,
	}

	i := installer.New(cfg)

	if installCheck {
		return i.CheckInstallations()
	}

	return i.InstallTools()
}
