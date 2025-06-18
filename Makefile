# WASM Manager Makefile
# High-performance WASM build system

.PHONY: help build install clean test validate build-all dev setup

# Default target
help:
	@echo "🚀 WASM Manager v2.0"
	@echo "===================="
	@echo ""
	@echo "Available targets:"
	@echo "  setup        - Install Go dependencies and build the manager"
	@echo "  build        - Build the wasm-manager binary"
	@echo "  install      - Install the manager globally"
	@echo "  build-all    - Build all WASM modules using Go manager"
	@echo "  test         - Test all modules"
	@echo "  validate     - Validate all modules"
	@echo "  clean        - Clean build artifacts"
	@echo "  clean-all    - Clean everything including Go binary"
	@echo "  dev          - Build and run in development mode"
	@echo "  install-tools - Install WASM optimization tools"
	@echo ""
	@echo "Examples:"
	@echo "  make setup                    # Initial setup"
	@echo "  make build-all                # Build all modules"
	@echo "  make build crypto-wasm        # Build specific module"
	@echo "  make validate                 # Validate all modules"

# Setup: Install dependencies and build
setup:
	@echo "🔧 Setting up WASM Manager..."
	go mod tidy
	go build -o wasm-manager .
	@echo "✅ Setup complete! Run './wasm-manager --help' to get started."

# Build the Go manager binary
build:
	@echo "🔨 Building WASM Manager..."
	go build -ldflags="-s -w" -o wasm-manager .
	@echo "✅ Manager built successfully!"

# Install globally
install: build
	@echo "📦 Installing WASM Manager globally..."
	sudo cp wasm-manager /usr/local/bin/
	@echo "✅ Manager installed! You can now use 'wasm-manager' from anywhere."

# Build all WASM modules using the Go manager
build-all: build
	@echo "🏗️ Building all WASM modules..."
	./wasm-manager build --verbose

# Build specific module
build-%: build
	@echo "🏗️ Building module $*..."
	./wasm-manager build $* --verbose

# Test all modules
test: build
	@echo "🧪 Testing all modules..."
	./wasm-manager test --verbose

# Validate all modules
validate: build
	@echo "🔍 Validating all modules..."
	./wasm-manager validate --verbose

# Clean build artifacts from modules
clean: build
	@echo "🧹 Cleaning module artifacts..."
	./wasm-manager clean --verbose

# Clean everything including Go binary
clean-all:
	@echo "🧹 Cleaning everything..."
	rm -f wasm-manager
	@if [ -f "./wasm-manager" ]; then ./wasm-manager clean --all; fi
	go clean -cache

# Install WASM optimization tools
install-tools: build
	@echo "🔧 Installing WASM optimization tools..."
	./wasm-manager install-tools --verbose

# Development mode: build and show help
dev: build
	@echo "🔧 Development mode..."
	./wasm-manager --help

# Benchmark builds
benchmark: build
	@echo "⏱️ Benchmarking builds..."
	time ./wasm-manager build --workers 1
	time ./wasm-manager build --workers 4  
	time ./wasm-manager build --workers 8

# Legacy compatibility (will show migration message)
legacy-build:
	@echo "⚠️  Legacy build.sh scripts are deprecated!"
	@echo "🔄 Please use the new Go manager:"
	@echo "   make build-all    # Build all modules"
	@echo "   make build-math   # Build math-wasm module"
	@echo ""
	@echo "🚀 The new manager is much faster with parallel builds!"

# Remove old shell scripts (optional cleanup)
remove-legacy:
	@echo "🗑️  Removing legacy shell scripts..."
	@echo "This will remove all build.sh files. Are you sure? [y/N]"
	@read -r REPLY; \
	if [ "$$REPLY" = "y" ] || [ "$$REPLY" = "Y" ]; then \
		find . -name "build.sh" -type f -delete; \
		rm -f scripts/*.sh; \
		rm -f wasm-manager.sh; \
		echo "✅ Legacy scripts removed!"; \
	else \
		echo "❌ Operation cancelled."; \
	fi 