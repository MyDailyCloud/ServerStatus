#!/bin/bash

# ServerStatus Dashboard Installation Script
# This script downloads and installs the data-server component for Linux and macOS

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on supported OS
check_os() {
    case "$(uname -s)" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            print_error "This script only supports Linux and macOS"
            exit 1
            ;;
    esac
}

# Detect system architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    print_status "Fetching latest version information..."
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s https://api.github.com/repos/MyDailyCloud/ServerStatus/releases/latest | grep '"tag_name"' | cut -d'"' -f4 2>/dev/null || echo "")
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- https://api.github.com/repos/MyDailyCloud/ServerStatus/releases/latest | grep '"tag_name"' | cut -d'"' -f4 2>/dev/null || echo "")
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version information"
        exit 1
    fi
    
    print_success "Latest version: $VERSION"
}

# Determine installation directory
get_install_dir() {
    if [ -w "/usr/local/bin" ] 2>/dev/null; then
        INSTALL_DIR="/usr/local/bin"
        print_status "Installing to system directory: $INSTALL_DIR"
    else
        INSTALL_DIR="$(pwd)"
        print_warning "No write permission to /usr/local/bin, installing to current directory: $INSTALL_DIR"
    fi
}

# Download and install data-server
install_server() {
    print_status "Downloading data-server for $OS-$ARCH..."
    
    SERVER_URL="https://github.com/MyDailyCloud/ServerStatus/releases/download/$VERSION/data-server-$OS-$ARCH"
    SERVER_PATH="$INSTALL_DIR/data-server"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L "$SERVER_URL" -o "$SERVER_PATH"
    elif command -v wget >/dev/null 2>&1; then
        wget "$SERVER_URL" -O "$SERVER_PATH"
    fi
    
    if [ ! -f "$SERVER_PATH" ]; then
        print_error "Failed to download data-server"
        exit 1
    fi
    
    chmod +x "$SERVER_PATH"
    print_success "data-server installed to: $SERVER_PATH"
}

# Create default configuration
create_config() {
    CONFIG_PATH="$INSTALL_DIR/server-config.json"
    
    if [ ! -f "$CONFIG_PATH" ]; then
        print_status "Creating default configuration file..."
        
        cat > "$CONFIG_PATH" << 'EOF'
{
  "port": 8080,
  "database": {
    "type": "sqlite",
    "path": "./data.db"
  },
  "auth": {
    "secret_key": "your-secret-key-here",
    "admin_user": "admin",
    "admin_password": "admin123"
  },
  "web_ui": {
    "enabled": true,
    "path": "./web-ui"
  }
}
EOF
        
        print_success "Configuration file created: $CONFIG_PATH"
        print_warning "Please edit $CONFIG_PATH to configure your settings"
    else
        print_status "Configuration file already exists: $CONFIG_PATH"
    fi
}

# Start data-server in background
start_server() {
    print_status "Starting data-server in background..."
    
    SERVER_PATH="$INSTALL_DIR/data-server"
    CONFIG_PATH="$INSTALL_DIR/server-config.json"
    PID_FILE="$INSTALL_DIR/data-server.pid"
    LOG_FILE="$INSTALL_DIR/data-server.log"
    
    # Stop existing instance if running
    if [ -f "$PID_FILE" ]; then
        OLD_PID=$(cat "$PID_FILE")
        if kill -0 "$OLD_PID" 2>/dev/null; then
            print_status "Stopping existing data-server (PID: $OLD_PID)..."
            kill "$OLD_PID"
            sleep 2
        fi
        rm -f "$PID_FILE"
    fi
    
    # Start new instance
    cd "$INSTALL_DIR"
    nohup "$SERVER_PATH" -config="$CONFIG_PATH" > "$LOG_FILE" 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > "$PID_FILE"
    
    # Wait a moment and check if process is still running
    sleep 2
    if kill -0 "$SERVER_PID" 2>/dev/null; then
        print_success "data-server started successfully (PID: $SERVER_PID)"
        print_success "Log file: $LOG_FILE"
        print_success "PID file: $PID_FILE"
    else
        print_error "Failed to start data-server"
        if [ -f "$LOG_FILE" ]; then
            print_error "Check log file for details: $LOG_FILE"
        fi
        exit 1
    fi
}

# Print usage information
print_usage() {
    echo
    print_success "=== ServerStatus Dashboard Installation Complete ==="
    echo
    echo "Files installed:"
    echo "  • data-server: $INSTALL_DIR/data-server"
    echo "  • Configuration: $INSTALL_DIR/server-config.json"
    echo "  • Log file: $INSTALL_DIR/data-server.log"
    echo "  • PID file: $INSTALL_DIR/data-server.pid"
    echo
    echo "Management commands:"
    echo "  • View logs: tail -f $INSTALL_DIR/data-server.log"
    echo "  • Stop server: kill \$(cat $INSTALL_DIR/data-server.pid)"
    echo "  • Restart: $0"
    echo
    echo "Next steps:"
    echo "  1. Edit configuration file: $INSTALL_DIR/server-config.json"
    echo "  2. Access web interface: http://localhost:8080"
    echo "  3. Configure monitor agents to connect to this server"
    echo
}

# Main installation process
main() {
    print_status "Starting ServerStatus Dashboard installation..."
    
    check_os
    detect_arch
    get_latest_version
    get_install_dir
    install_server
    create_config
    start_server
    print_usage
    
    print_success "Installation completed successfully!"
}

# Run main function
main "$@"