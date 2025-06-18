package installer

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Installer handles tool installation
type Installer struct {
	config *Config
}

// Config holds installer configuration
type Config struct {
	CheckOnly    bool
	BinaryenOnly bool
	WABTOnly     bool
	Force        bool
	Verbose      bool
}

// New creates a new Installer instance
func New(cfg *Config) *Installer {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Installer{config: cfg}
}

// CheckInstallations checks which tools are installed
func (i *Installer) CheckInstallations() error {
	fmt.Println("🔍 Checking tool installations...")
	fmt.Println("================================")

	tools := []struct {
		name    string
		command string
		args    []string
	}{
		{"wasm-opt", "wasm-opt", []string{"--version"}},
		{"wasm2wat", "wasm2wat", []string{"--version"}},
		{"wat2wasm", "wat2wasm", []string{"--version"}},
		{"gzip", "gzip", []string{"--version"}},
		{"brotli", "brotli", []string{"--version"}},
		{"base64", "base64", []string{"--version"}},
	}

	allInstalled := true

	for _, tool := range tools {
		if i.isToolInstalled(tool.command, tool.args) {
			version := i.getToolVersion(tool.command, tool.args)
			fmt.Printf("✅ %-10s %s\n", tool.name+":", version)
		} else {
			fmt.Printf("❌ %-10s not installed\n", tool.name+":")
			allInstalled = false
		}
	}

	fmt.Println()
	if allInstalled {
		fmt.Println("🎉 All tools are installed!")
	} else {
		fmt.Println("⚠️  Some tools are missing. Run without --check to install them.")
	}

	return nil
}

// InstallTools installs the required tools
func (i *Installer) InstallTools() error {
	fmt.Println("🔧 Installing WASM optimization tools...")
	fmt.Println("========================================")

	os := i.detectOS()
	fmt.Printf("Detected OS: %s\n\n", os)

	var installFunctions []func() error

	if !i.config.WABTOnly {
		installFunctions = append(installFunctions, func() error {
			return i.installBinaryen(os)
		})
	}

	if !i.config.BinaryenOnly {
		installFunctions = append(installFunctions, func() error {
			return i.installWABT(os)
		})
	}

	// Always check for compression tools
	installFunctions = append(installFunctions, func() error {
		return i.installCompressionTools(os)
	})

	for _, installFunc := range installFunctions {
		if err := installFunc(); err != nil {
			return err
		}
	}

	fmt.Println("\n✅ Installation completed!")
	return i.CheckInstallations()
}

// detectOS detects the operating system
func (i *Installer) detectOS() string {
	switch runtime.GOOS {
	case "linux":
		// Try to detect distribution
		if i.commandExists("apt-get") {
			return "ubuntu"
		} else if i.commandExists("yum") || i.commandExists("dnf") {
			return "rhel"
		} else if i.commandExists("pacman") {
			return "arch"
		}
		return "linux"
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return "unknown"
	}
}

// installBinaryen installs Binaryen (wasm-opt)
func (i *Installer) installBinaryen(os string) error {
	fmt.Println("📦 Installing Binaryen (wasm-opt)...")

	switch os {
	case "ubuntu":
		return i.runCommand("sudo", "apt", "update")
		if err := i.runCommand("sudo", "apt", "install", "-y", "binaryen"); err != nil {
			return err
		}
	case "rhel":
		if i.commandExists("dnf") {
			return i.runCommand("sudo", "dnf", "install", "-y", "binaryen")
		} else {
			return i.runCommand("sudo", "yum", "install", "-y", "binaryen")
		}
	case "arch":
		return i.runCommand("sudo", "pacman", "-S", "--noconfirm", "binaryen")
	case "macos":
		if !i.commandExists("brew") {
			return fmt.Errorf("Homebrew is required for macOS installation. Please install it first")
		}
		return i.runCommand("brew", "install", "binaryen")
	default:
		fmt.Println("⚠️  Automatic installation not supported for your OS.")
		fmt.Println("Please install Binaryen manually from: https://github.com/WebAssembly/binaryen/releases")
		return nil
	}

	return nil
}

// installWABT installs WebAssembly Binary Toolkit
func (i *Installer) installWABT(os string) error {
	fmt.Println("📦 Installing WABT (WebAssembly Binary Toolkit)...")

	switch os {
	case "ubuntu":
		return i.runCommand("sudo", "apt", "install", "-y", "wabt")
	case "macos":
		if !i.commandExists("brew") {
			return fmt.Errorf("Homebrew is required for macOS installation")
		}
		return i.runCommand("brew", "install", "wabt")
	default:
		fmt.Println("⚠️  WABT installation not supported for your OS.")
		fmt.Println("Please install manually from: https://github.com/WebAssembly/wabt/releases")
		return nil
	}

	return nil
}

// installCompressionTools installs compression tools
func (i *Installer) installCompressionTools(os string) error {
	fmt.Println("📦 Checking compression tools...")

	// Check gzip (usually pre-installed)
	if !i.commandExists("gzip") {
		fmt.Println("Installing gzip...")
		switch os {
		case "ubuntu":
			i.runCommand("sudo", "apt", "install", "-y", "gzip")
		case "rhel":
			if i.commandExists("dnf") {
				i.runCommand("sudo", "dnf", "install", "-y", "gzip")
			} else {
				i.runCommand("sudo", "yum", "install", "-y", "gzip")
			}
		case "macos":
			// gzip should be pre-installed on macOS
		}
	}

	// Check brotli
	if !i.commandExists("brotli") {
		fmt.Println("Installing brotli...")
		switch os {
		case "ubuntu":
			i.runCommand("sudo", "apt", "install", "-y", "brotli")
		case "rhel":
			if i.commandExists("dnf") {
				i.runCommand("sudo", "dnf", "install", "-y", "brotli")
			}
		case "macos":
			if i.commandExists("brew") {
				i.runCommand("brew", "install", "brotli")
			}
		}
	}

	return nil
}

// Helper functions
func (i *Installer) isToolInstalled(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	return cmd.Run() == nil
}

func (i *Installer) getToolVersion(command string, args []string) string {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return "unknown"
}

func (i *Installer) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (i *Installer) runCommand(name string, args ...string) error {
	if i.config.Verbose {
		fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
	}

	cmd := exec.Command(name, args...)
	if i.config.Verbose {
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run()
	} else {
		return cmd.Run()
	}
}
