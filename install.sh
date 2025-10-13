#!/bin/bash


set -e

GITHUB_REPO="MyDailyCloud/ServerStatus"
VERSION="${INSTALL_VERSION:-v1.0.4}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/serverstatus}"
SERVER_URL="${SERVER_URL:-http://localhost:8080}"
SERVER_PORT="${SERVER_PORT:-8080}"
SERVER_KEY="${SERVER_KEY:-public}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

show_usage() {
    cat << EOF
${BLUE}ServerStatus Automated Installation Script${NC}

Usage:
    ./install.sh [MODE] [OPTIONS]

Modes:
    server              Install data server
    client              Install monitoring agent
    (no mode)           Interactive installation

Server Options:
    --port PORT         Server port (default: 8080)
    --key KEY           Authentication key (default: public)
    --host HOST         Server host (default: 0.0.0.0)

Client Options:
    --url URL           Server API URL (default: http://localhost:8080/api/data)
    --key KEY           Authentication key (default: public)

Environment Variables:
    INSTALL_VERSION     Version to install (default: v1.0.4)
    INSTALL_DIR         Installation directory (default: ~/serverstatus)
    SERVER_URL          Server URL for client (default: http://localhost:8080)
    SERVER_PORT         Server port (default: 8080)
    SERVER_KEY          Authentication key (default: public)

Examples:
    ./install.sh

    ./install.sh server --port 8080 --key public

    ./install.sh client --url http://example.com:8080/api/data --key public

EOF
    exit 0
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux*)
            OS_NAME="linux"
            ;;
        darwin*)
            OS_NAME="darwin"
            ;;
        *)
            echo -e "${RED}Error: Unsupported OS: $OS${NC}"
            echo "Supported: Linux, macOS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH_NAME="amd64"
            ;;
        aarch64|arm64)
            ARCH_NAME="arm64"
            ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            echo "Supported: x86_64/amd64, aarch64/arm64"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Detected platform: ${OS_NAME}-${ARCH_NAME}${NC}"
}

check_requirements() {
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        echo -e "${RED}Error: curl or wget is required${NC}"
        echo "Install curl: sudo apt install curl (Ubuntu/Debian) or sudo yum install curl (CentOS/RHEL)"
        exit 1
    fi
}

download_binary() {
    local component=$1
    local binary_name="${component}-${OS_NAME}-${ARCH_NAME}"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${binary_name}"
    local target_file="${INSTALL_DIR}/${component}"
    
    echo -e "${YELLOW}Downloading ${component}...${NC}"
    echo -e "${BLUE}URL: ${download_url}${NC}"
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -L -f -o "${target_file}" "${download_url}"; then
            echo -e "${RED}Download failed${NC}"
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -O "${target_file}" "${download_url}"; then
            echo -e "${RED}Download failed${NC}"
            exit 1
        fi
    fi
    
    chmod +x "${target_file}"
    echo -e "${GREEN}Downloaded successfully${NC}"
}

install_server() {
    echo -e "${BLUE}=== Installing ServerStatus Data Server ===${NC}"
    
    mkdir -p "${INSTALL_DIR}"
    cd "${INSTALL_DIR}"
    
    download_binary "data-server"
    
    cat > server-config.json << EOF
{
    "host": "${SERVER_HOST:-0.0.0.0}",
    "port": "${SERVER_PORT}",
    "project_key": "${SERVER_KEY}",
    "server_key": "",
    "require_auth": false,
    "database_path": "./data/serverstatus.db",
    "data_limit": 1500,
    "data_interval": 15,
    "enable_compression": true,
    "compression_level": 6,
    "enable_websocket": true,
    "enable_cache": true,
    "redis_addr": "localhost:6379",
    "redis_password": "",
    "redis_db": 0
}
EOF
    
    cat > start-server.sh << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
./data-server
EOF
    
    chmod +x start-server.sh
    
    echo -e "${GREEN}=== Server Installation Complete ===${NC}"
    echo -e "${BLUE}Installation directory: ${INSTALL_DIR}${NC}"
    echo -e "${BLUE}Configuration file: ${INSTALL_DIR}/server-config.json${NC}"
    echo ""
    echo -e "${YELLOW}Start server:${NC}"
    echo -e "  cd ${INSTALL_DIR} && ./start-server.sh"
    echo ""
    echo -e "${YELLOW}Or start with custom options:${NC}"
    echo -e "  cd ${INSTALL_DIR} && ./data-server -key ${SERVER_KEY} -port ${SERVER_PORT}"
    echo ""
    echo -e "${YELLOW}Access dashboard:${NC}"
    echo -e "  http://localhost:${SERVER_PORT}/?key=${SERVER_KEY}"
    echo ""
    echo -e "${YELLOW}Run in background:${NC}"
    echo -e "  cd ${INSTALL_DIR} && nohup ./start-server.sh > server.log 2>&1 &"
}

install_client() {
    echo -e "${BLUE}=== Installing ServerStatus Monitor Agent ===${NC}"
    
    mkdir -p "${INSTALL_DIR}"
    cd "${INSTALL_DIR}"
    
    download_binary "monitor-agent"
    
    cat > start-agent.sh << EOF
#!/bin/bash
cd "\$(dirname "\$0")"
./monitor-agent -url "${CLIENT_URL}" -key "${SERVER_KEY}"
EOF
    
    chmod +x start-agent.sh
    
    echo -e "${GREEN}=== Agent Installation Complete ===${NC}"
    echo -e "${BLUE}Installation directory: ${INSTALL_DIR}${NC}"
    echo ""
    echo -e "${YELLOW}Start agent:${NC}"
    echo -e "  cd ${INSTALL_DIR} && ./start-agent.sh"
    echo ""
    echo -e "${YELLOW}Or start with custom options:${NC}"
    echo -e "  cd ${INSTALL_DIR} && ./monitor-agent -url ${CLIENT_URL} -key ${SERVER_KEY}"
    echo ""
    echo -e "${YELLOW}Run in background:${NC}"
    echo -e "  cd ${INSTALL_DIR} && nohup ./start-agent.sh > agent.log 2>&1 &"
}

interactive_mode() {
    echo -e "${BLUE}=== ServerStatus Interactive Installation ===${NC}"
    echo ""
    
    echo -e "${YELLOW}What would you like to install?${NC}"
    echo "  1) Data Server (监控服务器)"
    echo "  2) Monitor Agent (监控代理)"
    echo "  3) Both (两者都安装)"
    read -p "Select [1-3]: " choice
    
    case $choice in
        1)
            echo ""
            read -p "Server port (default: 8080): " port_input
            SERVER_PORT="${port_input:-8080}"
            
            read -p "Authentication key (default: public): " key_input
            SERVER_KEY="${key_input:-public}"
            
            read -p "Installation directory (default: ~/serverstatus): " dir_input
            INSTALL_DIR="${dir_input:-$HOME/serverstatus}"
            
            install_server
            ;;
        2)
            echo ""
            read -p "Server API URL (default: http://localhost:8080/api/data): " url_input
            CLIENT_URL="${url_input:-http://localhost:8080/api/data}"
            
            read -p "Authentication key (default: public): " key_input
            SERVER_KEY="${key_input:-public}"
            
            read -p "Installation directory (default: ~/serverstatus): " dir_input
            INSTALL_DIR="${dir_input:-$HOME/serverstatus}"
            
            install_client
            ;;
        3)
            echo ""
            read -p "Server port (default: 8080): " port_input
            SERVER_PORT="${port_input:-8080}"
            
            read -p "Authentication key (default: public): " key_input
            SERVER_KEY="${key_input:-public}"
            
            read -p "Installation directory (default: ~/serverstatus): " dir_input
            INSTALL_DIR="${dir_input:-$HOME/serverstatus}"
            
            install_server
            echo ""
            echo -e "${BLUE}Now installing monitor agent...${NC}"
            CLIENT_URL="http://localhost:${SERVER_PORT}/api/data"
            install_client
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            exit 1
            ;;
    esac
}

main() {
    echo -e "${BLUE}╔═══════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   ServerStatus Automated Installation        ║${NC}"
    echo -e "${BLUE}║   Version: ${VERSION}                            ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════╝${NC}"
    echo ""
    
    detect_platform
    check_requirements
    
    if [ $# -eq 0 ]; then
        interactive_mode
        exit 0
    fi
    
    MODE=$1
    shift
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --port)
                SERVER_PORT="$2"
                shift 2
                ;;
            --key)
                SERVER_KEY="$2"
                shift 2
                ;;
            --host)
                SERVER_HOST="$2"
                shift 2
                ;;
            --url)
                CLIENT_URL="$2"
                shift 2
                ;;
            --help)
                show_usage
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                show_usage
                ;;
        esac
    done
    
    case $MODE in
        server)
            install_server
            ;;
        client)
            CLIENT_URL="${CLIENT_URL:-${SERVER_URL}/api/data}"
            install_client
            ;;
        --help)
            show_usage
            ;;
        *)
            echo -e "${RED}Unknown mode: $MODE${NC}"
            show_usage
            ;;
    esac
}

main "$@"
