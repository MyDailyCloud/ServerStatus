#!/bin/bash

echo "========================================"
echo "GPU Monitor 自动编译脚本"
echo "========================================"
echo

echo "清理旧的编译文件..."
rm -f release/data-server.exe
rm -f release/monitor-agent.exe
rm -f release/data-server-linux
rm -f release/monitor-agent-linux

echo "创建release目录..."
mkdir -p release

echo
echo "[1/4] 编译 Linux data-server..."
cd data-server
go build -o ../release/data-server-linux ./main.go
if [ $? -ne 0 ]; then
    echo "错误: Linux data-server 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ Linux data-server 编译完成"

echo
echo "[2/4] 编译 Linux monitor-agent..."
cd monitor-agent
go build -o ../release/monitor-agent-linux ./main.go
if [ $? -ne 0 ]; then
    echo "错误: Linux monitor-agent 编译失败"
    cd ..
    exit 1
fi
cd ..
echo "✓ Linux monitor-agent 编译完成"

echo
echo "[3/4] 设置Windows环境并编译 data-server..."
export GOOS=windows
cd data-server
go build -o ../release/data-server.exe ./main.go
if [ $? -ne 0 ]; then
    echo "错误: Windows data-server 编译失败"
    cd ..
    unset GOOS
    exit 1
fi
cd ..
echo "✓ Windows data-server 编译完成"

echo
echo "[4/4] 编译 Windows monitor-agent..."
cd monitor-agent
go build -o ../release/monitor-agent.exe ./main.go
if [ $? -ne 0 ]; then
    echo "错误: Windows monitor-agent 编译失败"
    cd ..
    unset GOOS
    exit 1
fi
cd ..
echo "✓ Windows monitor-agent 编译完成"

echo
echo "恢复Linux环境..."
unset GOOS

echo
echo "========================================"
echo "编译完成！文件位置:"
echo "========================================"
ls -la release/
echo
echo "所有程序已成功编译到 release 目录"
echo "包含多GPU监控功能的完整版本"
echo "========================================"