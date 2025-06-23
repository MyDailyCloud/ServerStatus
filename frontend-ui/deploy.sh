#!/bin/bash

# ServerStatus Frontend Deployment Script
# 前端部署脚本

set -e

# 配置
FRONTEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$FRONTEND_DIR/dist"
DEFAULT_API_URL="http://localhost:8080"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== ServerStatus Frontend Deployment ===${NC}"
echo ""

# 获取API服务器地址
read -p "Enter API server URL (default: $DEFAULT_API_URL): " API_URL
API_URL=${API_URL:-$DEFAULT_API_URL}

echo -e "${YELLOW}Building frontend with API URL: $API_URL${NC}"

# 创建构建目录
mkdir -p "$BUILD_DIR"

# 复制所有文件
cp -r "$FRONTEND_DIR"/* "$BUILD_DIR/" 2>/dev/null || true

# 排除不需要的文件
rm -rf "$BUILD_DIR/dist"
rm -f "$BUILD_DIR/deploy.sh"
rm -f "$BUILD_DIR/package.json"
rm -f "$BUILD_DIR/README.md"

# 更新配置文件中的API URL
if [ -f "$BUILD_DIR/js/config.js" ]; then
    sed -i "s|API_BASE_URL:.*|API_BASE_URL: '$API_URL',|g" "$BUILD_DIR/js/config.js"
    echo -e "${GREEN}Updated API URL in config.js${NC}"
fi

echo -e "${GREEN}Build completed!${NC}"
echo -e "Build output: $BUILD_DIR"
echo ""
echo -e "${YELLOW}Deployment options:${NC}"
echo ""

echo -e "${BLUE}1. Local testing:${NC}"
echo "   cd $BUILD_DIR"
echo "   python3 -m http.server 3000"
echo "   # Access: http://localhost:3000"
echo ""

echo -e "${BLUE}2. Nginx deployment:${NC}"
echo "   sudo cp -r $BUILD_DIR/* /var/www/html/"
echo "   # Configure nginx to serve static files"
echo ""

echo -e "${BLUE}3. Docker deployment:${NC}"
echo "   cd $BUILD_DIR"
echo "   docker run -d -p 80:80 -v \$(pwd):/usr/share/nginx/html nginx"
echo ""

echo -e "${BLUE}4. Apache deployment:${NC}"
echo "   sudo cp -r $BUILD_DIR/* /var/www/html/"
echo "   # Configure apache to serve static files"
echo ""

echo -e "${GREEN}Deployment script completed!${NC}"