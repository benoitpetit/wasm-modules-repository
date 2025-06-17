#!/bin/bash

# Script d'installation des outils d'optimisation WASM
# Installe les dépendances nécessaires pour optimiser les builds WASM

set -e

echo "🔧 Installation des outils d'optimisation WASM..."
echo "================================================"

# Détection de l'OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt-get &> /dev/null; then
            echo "ubuntu"
        elif command -v yum &> /dev/null; then
            echo "rhel"
        elif command -v pacman &> /dev/null; then
            echo "arch"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

install_binaryen() {
    local os=$(detect_os)
    
    echo "📦 Installation de Binaryen (wasm-opt)..."
    
    case $os in
        "ubuntu")
            echo "🐧 Detected Ubuntu/Debian"
            sudo apt update
            sudo apt install -y binaryen
            ;;
        "rhel")
            echo "🎩 Detected RHEL/CentOS/Fedora"
            if command -v dnf &> /dev/null; then
                sudo dnf install -y binaryen
            else
                sudo yum install -y binaryen
            fi
            ;;
        "arch")
            echo "🏹 Detected Arch Linux"
            sudo pacman -S --noconfirm binaryen
            ;;
        "macos")
            echo "🍎 Detected macOS"
            if command -v brew &> /dev/null; then
                brew install binaryen
            else
                echo "❌ Homebrew not found. Please install Homebrew first:"
                echo "   /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                return 1
            fi
            ;;
        *)
            echo "⚠️  OS non détecté automatiquement. Installation manuelle requise:"
            echo "   Visitez: https://github.com/WebAssembly/binaryen/releases"
            echo "   Ou compilez depuis les sources:"
            echo "   git clone https://github.com/WebAssembly/binaryen.git"
            echo "   cd binaryen && cmake . && make"
            return 1
            ;;
    esac
}

install_wabt() {
    local os=$(detect_os)
    
    echo "📦 Installation de WABT (WebAssembly Binary Toolkit)..."
    
    case $os in
        "ubuntu")
            echo "🐧 Installing via apt..."
            sudo apt update
            sudo apt install -y wabt
            ;;
        "macos")
            echo "🍎 Installing via Homebrew..."
            if command -v brew &> /dev/null; then
                brew install wabt
            else
                echo "❌ Homebrew not found."
                return 1
            fi
            ;;
        *)
            echo "📥 Installing from GitHub releases..."
            local latest_url="https://api.github.com/repos/WebAssembly/wabt/releases/latest"
            local download_url
            
            if [[ "$os" == "linux" ]]; then
                download_url=$(curl -s $latest_url | grep "browser_download_url.*linux" | cut -d'"' -f4 | head -n1)
            else
                echo "⚠️  Installation manuelle requise pour votre OS"
                echo "   Visitez: https://github.com/WebAssembly/wabt/releases"
                return 1
            fi
            
            if [ -n "$download_url" ]; then
                local filename=$(basename "$download_url")
                echo "📥 Downloading $filename..."
                curl -L -o "/tmp/$filename" "$download_url"
                
                echo "📂 Extracting to /opt/wabt..."
                sudo mkdir -p /opt/wabt
                sudo tar -xzf "/tmp/$filename" -C /opt/wabt --strip-components=1
                
                echo "🔗 Creating symlinks..."
                sudo ln -sf /opt/wabt/bin/* /usr/local/bin/
                
                rm "/tmp/$filename"
            fi
            ;;
    esac
}

verify_installation() {
    echo "✅ Vérification des installations..."
    
    local all_good=true
    
    # Vérifier wasm-opt
    if command -v wasm-opt &> /dev/null; then
        local version=$(wasm-opt --version | head -n1)
        echo "✅ wasm-opt: $version"
    else
        echo "❌ wasm-opt non trouvé"
        all_good=false
    fi
    
    # Vérifier wasm2wat
    if command -v wasm2wat &> /dev/null; then
        local version=$(wasm2wat --version | head -n1)
        echo "✅ wasm2wat: $version"
    else
        echo "⚠️  wasm2wat non trouvé (optionnel)"
    fi
    
    # Vérifier wat2wasm
    if command -v wat2wasm &> /dev/null; then
        local version=$(wat2wasm --version | head -n1)
        echo "✅ wat2wasm: $version"
    else
        echo "⚠️  wat2wasm non trouvé (optionnel)"
    fi
    
    # Vérifier gzip
    if command -v gzip &> /dev/null; then
        echo "✅ gzip: disponible"
    else
        echo "❌ gzip non trouvé"
        all_good=false
    fi
    
    # Vérifier base64
    if command -v base64 &> /dev/null; then
        echo "✅ base64: disponible"
    else
        echo "❌ base64 non trouvé"
        all_good=false
    fi
    
    if $all_good; then
        echo ""
        echo "🎉 Tous les outils essentiels sont installés!"
        echo ""
        echo "📋 Outils disponibles:"
        echo "   • wasm-opt: Optimisation des fichiers WASM"
        echo "   • gzip: Compression des fichiers"
        echo "   • base64: Génération de hash d'intégrité"
        echo ""
        echo "🚀 Vous pouvez maintenant utiliser les builds optimisés!"
        return 0
    else
        echo ""
        echo "❌ Certains outils essentiels manquent. Veuillez les installer manuellement."
        return 1
    fi
}

show_usage() {
    echo "🔧 Script d'installation des outils d'optimisation WASM"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --binaryen    Installer seulement Binaryen (wasm-opt)"
    echo "  --wabt        Installer seulement WABT"
    echo "  --check       Vérifier les installations existantes"
    echo "  --help        Afficher cette aide"
    echo ""
    echo "Sans options, installe tous les outils disponibles."
}

main() {
    case "${1:-all}" in
        "--binaryen")
            install_binaryen
            verify_installation
            ;;
        "--wabt")
            install_wabt
            verify_installation
            ;;
        "--check")
            verify_installation
            ;;
        "--help")
            show_usage
            ;;
        "all"|"")
            echo "🚀 Installation complète des outils d'optimisation..."
            echo ""
            
            install_binaryen
            echo ""
            install_wabt
            echo ""
            verify_installation
            ;;
        *)
            echo "❌ Option inconnue: $1"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
