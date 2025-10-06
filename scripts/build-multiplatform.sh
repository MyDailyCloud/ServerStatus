#!/bin/bash

echo "========================================"
echo "ServerStatus 多平台自动编译脚本"
echo "========================================"
echo

echo "清理旧的编译文件..."
rm -f release/data-server.exe
rm -f release/monitor-agent.exe
rm -f release/data-server-linux
rm -f release/monitor-agent-linux
rm -f release/data-server-macos
rm -f release/monitor-agent-macos

echo "创建release目录..."
mkdir -p release

echo
echo "更新Go模块依赖..."
go mod tidy

echo
echo "[1/6] 编译 Linux data-server..."
cd data-server
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../release/data-server-linux *.go
if [ $? -ne 0 ]; then
    echo "错误: Linux data-server 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ Linux data-server 编译完成"

echo
echo "[2/6] 编译 Linux monitor-agent..."
cd monitor-agent
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../release/monitor-agent-linux *.go
if [ $? -ne 0 ]; then
    echo "错误: Linux monitor-agent 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ Linux monitor-agent 编译完成"

echo
echo "[3/6] 编译 Windows data-server..."
cd data-server
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ../release/data-server.exe *.go
if [ $? -ne 0 ]; then
    echo "错误: Windows data-server 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ Windows data-server 编译完成"

echo
echo "[4/6] 编译 Windows monitor-agent..."
cd monitor-agent
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ../release/monitor-agent.exe *.go
if [ $? -ne 0 ]; then
    echo "错误: Windows monitor-agent 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ Windows monitor-agent 编译完成"

echo
echo "[5/6] 编译 macOS data-server..."
cd data-server
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ../release/data-server-macos *.go
if [ $? -ne 0 ]; then
    echo "错误: macOS data-server 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ macOS data-server 编译完成"

echo
echo "[6/6] 编译 macOS monitor-agent..."
cd monitor-agent
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ../release/monitor-agent-macos *.go
if [ $? -ne 0 ]; then
    echo "错误: macOS monitor-agent 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ macOS monitor-agent 编译完成"

echo
echo "创建版本信息..."
DATE=$(date +%Y%m%d_%H%M%S)
echo "Build Date: $DATE" > release/build-info.txt
echo "Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" >> release/build-info.txt
echo "Go Version: $(go version)" >> release/build-info.txt

echo
echo "========================================"
echo "编译完成！文件位置:"
echo "========================================"
ls -la release/
echo
echo "所有程序已成功编译到 release 目录"
echo "支持 Linux, Windows, macOS 三个平台"
echo "包含多GPU监控、WebSocket实时推送、Redis缓存等完整功能"
echo "========================================"
echo "版本信息: $DATE"