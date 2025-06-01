#!/bin/bash

# ServerStatus 一键安装脚本
# 支持 Linux 和 macOS

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
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

# 检测操作系统和架构
detect_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $OS in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            print_error "不支持的操作系统: $OS (仅支持 Linux 和 macOS)"
            print_info "Windows 用户请使用 install.bat 或 install.ps1"
            exit 1
            ;;
    esac
    
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        i386|i686)
            ARCH="386"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            print_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    
    print_info "检测到系统: $OS-$ARCH"
}

# 获取最新版本号
get_latest_version() {
    print_info "获取最新版本信息..."
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s https://api.github.com/repos/MyDailyCloud/ServerStatus/releases/latest | grep '"tag_name"' | cut -d'"' -f4 2>/dev/null || echo "")
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- https://api.github.com/repos/MyDailyCloud/ServerStatus/releases/latest | grep '"tag_name"' | cut -d'"' -f4 2>/dev/null || echo "")
    else
        print_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_warning "无法获取最新版本，使用默认版本 v1.0.0"
        VERSION="v1.0.0"
    fi
    
    print_info "使用版本: $VERSION"
}

# 下载文件
download_file() {
    local url=$1
    local output=$2
    
    print_info "下载: $(basename "$output")"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$output" "$url" --progress-bar
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$output" "$url" --progress=bar:force
    else
        print_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    if [ $? -eq 0 ]; then
        print_success "下载完成: $(basename "$output")"
    else
        print_error "下载失败: $(basename "$output")"
        exit 1
    fi
}

# 确定安装目录
get_install_dir() {
    # 尝试安装到系统目录
    if [ -w "/usr/local/bin" ] 2>/dev/null; then
        INSTALL_DIR="/usr/local/bin"
        NEED_SUDO=false
        print_info "安装到系统目录: $INSTALL_DIR"
    elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
        INSTALL_DIR="/usr/local/bin"
        NEED_SUDO=true
        print_info "使用 sudo 安装到系统目录: $INSTALL_DIR"
    else
        # 没有权限，使用当前目录
        INSTALL_DIR="$(pwd)"
        NEED_SUDO=false
        print_warning "没有系统目录写入权限，安装到当前目录: $INSTALL_DIR"
    fi
}

# 下载并安装
install_serverstatus() {
    get_install_dir
    
    # 构建下载URL
    AGENT_URL="https://github.com/MyDailyCloud/ServerStatus/releases/download/$VERSION/monitor-agent-$OS-$ARCH"
    
    # 下载 monitor-agent
    download_file "$AGENT_URL" "$INSTALL_DIR/monitor-agent"
    if [ "$NEED_SUDO" = true ]; then
        sudo chmod +x "$INSTALL_DIR/monitor-agent"
    else
        chmod +x "$INSTALL_DIR/monitor-agent"
    fi
    
    print_success "Monitor Agent 安装完成"
}

# 创建配置文件
create_config() {
    print_info "创建默认配置文件..."
    
    # 创建 monitor-agent 配置
    cat > "$INSTALL_DIR/config.json" << EOF
{
  "server_url": "https://status.example.com",
  "project_key": "your-project-key",
  "interval": 5,
  "timeout": 10
}
EOF
    
    if [ "$NEED_SUDO" = true ]; then
        sudo chown $(whoami):$(whoami) "$INSTALL_DIR/config.json" 2>/dev/null || true
    fi
    
    print_success "配置文件已创建"
}

# 启动服务
start_services() {
    print_info "启动 Monitor Agent..."
    
    # 启动 monitor-agent (后台运行)
    print_info "启动监控代理..."
    nohup "$INSTALL_DIR/monitor-agent" > /dev/null 2>&1 &
    AGENT_PID=$!
    
    # 等待代理启动
    sleep 2
    
    # 检查代理是否启动成功
    if kill -0 $AGENT_PID 2>/dev/null; then
        print_success "监控代理已启动 (PID: $AGENT_PID)"
    else
        print_warning "监控代理启动失败，请检查配置文件中的服务器地址"
    fi
    
    # 保存PID到文件
    echo $AGENT_PID > "$INSTALL_DIR/monitor-agent.pid"
}

# 显示使用说明
show_usage() {
    echo
    print_success "🎉 Monitor Agent 安装并启动完成！"
    echo
    print_info "📁 安装目录: $INSTALL_DIR"
    echo
    print_info "🔧 管理命令:"
    echo "  查看状态: ps aux | grep monitor-agent"
    echo "  停止服务: kill \$(cat $INSTALL_DIR/monitor-agent.pid 2>/dev/null)"
    echo "  重启服务: $0"
    echo
    print_info "📝 配置文件:"
    echo "  - Monitor Agent: $INSTALL_DIR/config.json"
    echo
    print_warning "⚠️  请修改配置文件中的 server_url 和 project_key 后重启服务"
    print_info "💡 配置示例:"
    echo "  - server_url: 你的 ServerStatus 服务器地址"
    echo "  - project_key: 你的项目密钥"
}

# 检查是否已经运行
check_running() {
    if [ -f "$INSTALL_DIR/monitor-agent.pid" ]; then
        local pid=$(cat "$INSTALL_DIR/monitor-agent.pid" 2>/dev/null)
        if [ -n "$pid" ] && kill -0 $pid 2>/dev/null; then
            print_warning "Monitor Agent 已在运行中 (PID: $pid)"
            print_info "如需重新安装，请先停止服务: kill $pid"
            exit 0
        fi
    fi
}

# 主函数
main() {
    echo "="*60
    echo "    ServerStatus 一键安装脚本"
    echo "    无需注册，一键开始监控！"
    echo "="*60
    echo
    
    detect_os_arch
    get_latest_version
    check_running
    install_serverstatus
    create_config
    start_services
    show_usage
    
    echo
    print_success "✅ 安装完成！Monitor Agent 已在后台运行！"
}

# 运行主函数
main "$@"