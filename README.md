# ServerStatus - Lightweight Server Monitoring System

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)
[![Release](https://img.shields.io/github/v/release/MyDailyCloud/ServerStatus)](https://github.com/MyDailyCloud/ServerStatus/releases)

A **simple and easy-to-use** server monitoring system that helps you keep track of your servers effortlessly. Deploy in just 3 minutes!

[中文文档](#中文文档) | [English](#english-documentation)

---

## English Documentation

### 🎯 Introduction

ServerStatus is a lightweight server monitoring solution built with Go, supporting real-time data collection, multi-platform deployment, and flexible authentication mechanisms. No complex configuration required - download and run!

### ✨ Features

- 🚀 **Super Simple** - One command to start, 3 minutes to deploy
- 📊 **Real-time Monitoring** - Comprehensive monitoring of CPU, Memory, Disk, Network, GPU, and Temperature
- 🌍 **Cross-platform** - Supports Linux, macOS, and Windows
- 🔐 **Flexible Authentication** - Public mode, project key, and dual-key authentication
- 💾 **Data Persistence** - SQLite storage for historical data with optional Redis cache

### 🚀 Quick Start

#### Option A: Automated Installation (Recommended)

Download and run the automated installation script:

```bash
# Download the installation script
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/install.sh -o install.sh
chmod +x install.sh

# Install server (interactive mode)
./install.sh server

# Install monitoring agent (interactive mode)
./install.sh client
```

The script will automatically:
- Detect your platform and architecture
- Download the appropriate binaries
- Generate configuration files
- Set up startup scripts

#### Option B: Manual Installation

**Step 1: Download and Start Server**

```bash
# Linux (x86_64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-linux-amd64
chmod +x data-server-linux-amd64
./data-server-linux-amd64 -key public -port 8080
```

<details>
<summary>Other Platform Download Links</summary>

```bash
# Linux (ARM64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-linux-arm64

# macOS (Intel)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-darwin-amd64

# macOS (Apple Silicon)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-darwin-arm64

# Windows - Download and double-click to run
https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-windows-amd64.exe
```

</details>

After the server starts, you'll see output similar to:

```
2025/10/13 17:35:51 Starting ServerStatus Monitor Data Server...
2025/10/13 17:35:51 Port: 8080
2025/10/13 17:35:51 API Server started on 0.0.0.0:8080
```

**Step 2: Download and Start Monitoring Agent**

Run on each server you want to monitor:

```bash
# Linux (x86_64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-linux-amd64
chmod +x monitor-agent-linux-amd64
./monitor-agent-linux-amd64 -url http://YOUR_SERVER_IP:8080/api/data -key public
```

<details>
<summary>Other Platform Download Links</summary>

```bash
# Linux (ARM64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-linux-arm64

# macOS (Intel)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-darwin-amd64

# macOS (Apple Silicon)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-darwin-arm64

# Windows - Download and double-click to run
https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-windows-amd64.exe
```

</details>

> **Tip**: For local testing, use `http://localhost:8080/api/data`

After the agent starts, you'll see output similar to:

```
2025/10/13 17:36:30 Starting ServerStatus Monitor Agent...
2025/10/13 17:36:30 Session registered successfully: bf94d98f-b27e-4f9a-8256-9fc06abf9865
2025/10/13 17:36:31 Data reported successfully - CPU: 0.4%, Memory: 2.9%, Disk: 8.4%
```

**Step 3: Access Monitoring Dashboard**

Open your browser and visit:

```
http://YOUR_SERVER_IP:8080/?key=public
```

🎉 **Done!** You can now see real-time monitoring data from your servers!

### 🔧 Troubleshooting

#### ❓ Port Already in Use?

Change the startup port:

```bash
./data-server-linux-amd64 -key public -port 9090
```

#### ❓ Cannot Access Monitoring Dashboard?

Check if the firewall allows the port:

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

#### ❓ Agent Connection Failed?

1. Verify the server IP address is correct
2. Verify the server port (default 8080) is accessible
3. Check that the `-key` parameter matches (both server and agent use `public`)

#### ❓ How to Run in Background?

Use `screen` or `nohup`:

```bash
# Using screen
screen -S serverstatus
./data-server-linux-amd64 -key public -port 8080
# Press Ctrl+A then D to detach

# Using nohup
nohup ./data-server-linux-amd64 -key public -port 8080 > server.log 2>&1 &
```

### 📚 More Documentation

Want to learn more? Check out the complete documentation:

- **[Full Chinese Documentation](docs/README_zh.md)** - Detailed installation, configuration, and advanced features
- **[English Documentation](docs/README.md)** - Full documentation in English
- **[Architecture Design](CLAUDE.md)** - System architecture and technical implementation
- **[Development Roadmap](ROADMAP.md)** - Project planning and future features

#### Advanced Features

- 🔐 **Dual-Key Authentication** - Enterprise-grade security
- 📁 **Multi-Project Management** - Complete isolation between project servers
- 🐳 **Docker Deployment** - One-click containerized deployment
- 🌐 **Nginx Reverse Proxy** - Production environment best practices
- 📊 **API Interface** - Develop your custom monitoring dashboard

See [Full Documentation](docs/README_zh.md) for details

### 🤝 Contributing

All forms of contributions are welcome!

- 🐛 **Report Bugs** - [Submit an Issue](https://github.com/MyDailyCloud/ServerStatus/issues)
- ✨ **Feature Requests** - [Request a Feature](https://github.com/MyDailyCloud/ServerStatus/issues/new)
- 📝 **Improve Documentation** - Help us improve the docs
- 🌍 **Multi-language Support** - Add more languages

### 📄 License

This project is open-sourced under the MIT License and can be freely used and modified.

---

## 中文文档

### 🎯 简介

ServerStatus 是一个基于 Go 语言开发的轻量级服务器监控解决方案，支持实时数据采集、多平台部署和灵活的认证机制。无需复杂配置，下载即用。

### ✨ 特点

- 🚀 **超级简单** - 一行命令启动，3分钟完成部署
- 📊 **实时监控** - CPU、内存、磁盘、网络、GPU、温度全面监控
- 🌍 **跨平台** - 支持 Linux、macOS、Windows 多平台
- 🔐 **灵活认证** - 支持公开模式、项目密钥、双密钥认证
- 💾 **数据持久化** - SQLite 存储历史数据，可选 Redis 缓存

### 🚀 快速开始

#### 方式 A：自动化安装（推荐）

下载并运行自动安装脚本：

```bash
# 下载安装脚本
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/install.sh -o install.sh
chmod +x install.sh

# 安装服务器（交互式模式）
./install.sh server

# 安装监控代理（交互式模式）
./install.sh client
```

脚本将自动：
- 检测您的平台和架构
- 下载适当的二进制文件
- 生成配置文件
- 设置启动脚本

#### 方式 B：手动安装

**步骤 1：下载并启动服务器**

```bash
# Linux (x86_64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-linux-amd64
chmod +x data-server-linux-amd64
./data-server-linux-amd64 -key public -port 8080
```

<details>
<summary>其他平台下载链接</summary>

```bash
# Linux (ARM64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-linux-arm64

# macOS (Intel)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-darwin-amd64

# macOS (Apple Silicon)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-darwin-arm64

# Windows - 下载后双击运行
https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/data-server-windows-amd64.exe
```

</details>

服务器启动后，您将看到类似输出：

```
2025/10/13 17:35:51 启动 ServerStatus Monitor Data Server...
2025/10/13 17:35:51 端口: 8080
2025/10/13 17:35:51 API服务器启动在 0.0.0.0:8080
```

**步骤 2：下载并启动监控代理**

在每台需要监控的服务器上执行：

```bash
# Linux (x86_64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-linux-amd64
chmod +x monitor-agent-linux-amd64
./monitor-agent-linux-amd64 -url http://您的服务器IP:8080/api/data -key public
```

<details>
<summary>其他平台下载链接</summary>

```bash
# Linux (ARM64)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-linux-arm64

# macOS (Intel)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-darwin-amd64

# macOS (Apple Silicon)
wget https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-darwin-arm64

# Windows - 下载后双击运行
https://github.com/MyDailyCloud/ServerStatus/releases/download/v1.0.4/monitor-agent-windows-amd64.exe
```

</details>

> **提示**：如果在本机测试，使用 `http://localhost:8080/api/data`

代理启动后，您将看到类似输出：

```
2025/10/13 17:36:30 启动 ServerStatus Monitor Agent...
2025/10/13 17:36:30 Session注册成功: bf94d98f-b27e-4f9a-8256-9fc06abf9865
2025/10/13 17:36:31 成功上报数据 - CPU: 0.4%, 内存: 2.9%, 磁盘: 8.4%
```

**步骤 3：访问监控面板**

打开浏览器访问：

```
http://您的服务器IP:8080/?key=public
```

🎉 **完成！** 您现在可以在网页上看到服务器的实时监控数据了！

### 🔧 常见问题

#### ❓ 端口被占用怎么办？

更改启动端口：

```bash
./data-server-linux-amd64 -key public -port 9090
```

#### ❓ 无法访问监控面板？

检查防火墙是否开放端口：

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

#### ❓ 代理连接失败？

1. 确认服务器 IP 地址正确
2. 确认服务器端口（默认 8080）可访问
3. 检查 `-key` 参数是否一致（服务器和代理都使用 `public`）

#### ❓ 如何后台运行？

使用 `screen` 或 `nohup`：

```bash
# 使用 screen
screen -S serverstatus
./data-server-linux-amd64 -key public -port 8080
# 按 Ctrl+A 再按 D 退出

# 使用 nohup
nohup ./data-server-linux-amd64 -key public -port 8080 > server.log 2>&1 &
```

### 📚 更多文档

想了解更多功能？查看完整文档：

- **[完整中文文档](docs/README_zh.md)** - 详细的安装部署、配置说明、高级功能
- **[English Documentation](docs/README.md)** - Full documentation in English
- **[架构设计](CLAUDE.md)** - 系统架构和技术实现
- **[开发路线图](ROADMAP.md)** - 项目规划和未来特性

#### 高级功能

- 🔐 **双密钥认证模式** - 企业级安全认证
- 📁 **多项目管理** - 不同项目服务器完全隔离
- 🐳 **Docker 部署** - 容器化一键部署
- 🌐 **Nginx 反向代理** - 生产环境最佳实践
- 📊 **API 接口** - 自定义开发您的监控面板

详见 [完整文档](docs/README_zh.md)

### 🤝 贡献

欢迎各种形式的贡献！

- 🐛 **报告 Bug** - [提交 Issue](https://github.com/MyDailyCloud/ServerStatus/issues)
- ✨ **功能建议** - [功能请求](https://github.com/MyDailyCloud/ServerStatus/issues/new)
- 📝 **改进文档** - 帮助完善文档
- 🌍 **多语言支持** - 添加更多语言

### 📄 开源协议

本项目基于 MIT 协议开源，可自由使用和修改。

---

<div align="center">

### 🌟 If this project helps you, please give it a Star! / 如果这个项目对您有帮助，请给个 Star 支持一下！🌟

[⭐ Star](https://github.com/MyDailyCloud/ServerStatus) • [🍴 Fork](https://github.com/MyDailyCloud/ServerStatus/fork) • [📢 Share](https://twitter.com/intent/tweet?text=Check%20out%20this%20awesome%20server%20monitoring%20project!&url=https://github.com/MyDailyCloud/ServerStatus)

**Making Server Monitoring Simple and Beautiful / 让服务器监控变得简单而美好** ❤️

</div>
