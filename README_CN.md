# ServerStatus - 轻量级服务器监控系统

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)
[![Release](https://img.shields.io/github/v/release/MyDailyCloud/ServerStatus)](https://github.com/MyDailyCloud/ServerStatus/releases)

一个**简单易用**的服务器监控系统，帮助您轻松跟踪服务器状态。3分钟即可部署！

**中文** | **[English](README.md)**

---

## 🎯 简介

ServerStatus 是一个基于 Go 语言开发的轻量级服务器监控解决方案，支持实时数据采集、多平台部署和灵活的认证机制。无需复杂配置，下载即用！

## ✨ 特点

- 🚀 **超级简单** - 一行命令启动，3分钟完成部署
- 📊 **实时监控** - CPU、内存、磁盘、网络、GPU、温度全面监控
- 🌍 **跨平台** - 支持 Linux、macOS、Windows 多平台
- 🔐 **灵活认证** - 支持公开模式、项目密钥、双密钥认证
- 💾 **数据持久化** - SQLite 存储历史数据，可选 Redis 缓存
- 🌈 **精美界面** - 现代化设计，支持亮色/暗色主题
- 🌐 **多语言** - 支持中文/英文界面
- 📱 **响应式设计** - 桌面和移动设备完美适配

## 🚀 快速开始

### 方式 A：自动化安装（推荐）

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

### 方式 B：手动安装

#### 步骤 1：下载并启动服务器

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

#### 步骤 2：下载并启动监控代理

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

#### 步骤 3：访问监控面板

打开浏览器访问：

```
http://您的服务器IP:8080/?key=public
```

🎉 **完成！** 您现在可以在网页上看到服务器的实时监控数据了！

## 📊 监控功能

### 系统监控
- **CPU 使用率** - 实时 CPU 利用率和负载平均值
- **内存** - RAM 和 swap 使用情况监控
- **磁盘** - 所有分区的磁盘空间使用
- **网络** - 实时网速和流量统计
- **温度** - CPU 和 GPU 温度监测
- **GPU** - NVIDIA GPU 利用率、显存和温度
- **用户资源** - 按用户划分的资源使用情况

### 面板功能
- **实时更新** - 基于 WebSocket 的实时数据更新
- **服务器分组** - 按项目或环境组织服务器
- **搜索与筛选** - 快速搜索和基于状态的筛选
- **智能告警** - 资源阈值自动告警
- **数据导出** - 以 CSV 或 JSON 格式导出服务器数据
- **主题支持** - 亮色和暗色主题选项
- **多语言** - 中英文界面

## 🔐 认证模式

ServerStatus 支持三种认证模式：

### 1. 公开模式（默认）

无需认证，适合测试和演示：

```bash
# 服务器
./data-server -key public -port 8080

# 代理
./monitor-agent -url http://server:8080/api/data -key public

# 访问
http://server:8080/?key=public
```

### 2. 项目密钥模式

使用项目密钥分离不同项目的服务器：

```bash
# 项目 A 的代理
./monitor-agent -url http://server:8080/api/data -key project-a

# 项目 B 的代理
./monitor-agent -url http://server:8080/api/data -key project-b
```

### 3. 双密钥认证模式

企业级安全，使用服务器密钥和项目密钥：

```bash
# 启用认证的服务器
./data-server -key public -port 8080 -server-key "your-secret-key"

# 使用双密钥的代理
./monitor-agent -url http://server:8080/api/data \
  -key project-a \
  -server-key "your-secret-key"

# 为前端生成访问密钥
curl -X POST http://server:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "your-secret-key", "project_key": "project-a"}'

# 使用生成的密钥访问
http://server:8080/?key=generated-access-key
```

## 🐳 Docker 部署

### Docker 快速开始

```bash
# 运行服务器
docker run -d -p 8080:8080 --name serverstatus-server \
  -v ./data:/app/data \
  mydailycloud/serverstatus:latest

# 在被监控服务器上运行代理
docker run -d --name serverstatus-agent \
  mydailycloud/serverstatus-agent:latest \
  -url http://server-ip:8080/api/data \
  -key public
```

### Docker Compose

```yaml
version: '3.8'
services:
  serverstatus-server:
    image: mydailycloud/serverstatus:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - PORT=8080
      - PROJECT_KEY=public
      - SERVER_KEY=your-secret-key
    restart: unless-stopped

  serverstatus-agent:
    image: mydailycloud/serverstatus-agent:latest
    environment:
      - SERVER_URL=http://serverstatus-server:8080/api/data
      - PROJECT_KEY=public
      - SERVER_KEY=your-secret-key
    depends_on:
      - serverstatus-server
    restart: unless-stopped
```

## 🔧 配置说明

### 服务器配置

配置文件：`server-config.json`

```json
{
  "host": "0.0.0.0",
  "port": 8080,
  "project_key": "public",
  "server_key": "your-secret-key",
  "require_auth": false,
  "database_path": "./data/serverstatus.db",
  "enable_cache": true,
  "redis_addr": "localhost:6379"
}
```

或使用命令行参数：

```bash
./data-server \
  -port 8080 \
  -key public \
  -server-key "your-secret" \
  -auth \
  -db-path ./data/serverstatus.db
```

### 代理配置

配置文件：`config.json`

```json
{
  "server_url": "http://server:8080/api/data",
  "project_key": "public",
  "server_key": "",
  "report_interval": 5000000000,
  "timeout": 10000000000,
  "enable_user_resources": true
}
```

或使用命令行参数：

```bash
./monitor-agent \
  -url http://server:8080/api/data \
  -key public \
  -server-key "your-secret" \
  -interval 5 \
  -hostname "custom-name"
```

## 🌐 API 文档

ServerStatus 提供完整的 RESTful API：

### 服务器端点

```bash
# 获取所有服务器
GET /api/servers

# 根据主机名获取服务器
GET /api/server/{hostname}

# 获取服务器数量
GET /api/uuid-count

# 健康检查
GET /api/health
```

### 项目端点

```bash
# 获取项目的服务器（使用访问密钥）
GET /api/access/{access_key}/servers

# 获取项目的服务器详情
GET /api/access/{access_key}/server/{hostname}
```

### 认证端点

```bash
# 生成访问密钥
POST /api/generate-access-key
Content-Type: application/json
{
  "server_key": "your-server-key",
  "project_key": "project-name"
}

# 注册会话
POST /api/register-session
Content-Type: application/json
{
  "project_key": "public"
}
```

### 数据采集端点

```bash
# 上报服务器数据（代理使用）
POST /api/data
X-Project-Key: public
X-Server-Key: your-secret (可选)
Content-Type: application/json
{
  "session_id": "uuid",
  "hostname": "server1",
  ...
}
```

### WebSocket 端点

```bash
# 实时数据更新
WS /ws
```

完整 API 文档，请访问：`http://your-server:8080/API.md`

## 🛠️ 常见问题

### ❓ 端口被占用怎么办？

更改启动端口：

```bash
./data-server-linux-amd64 -key public -port 9090
```

### ❓ 无法访问监控面板？

检查防火墙是否开放端口：

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

### ❓ 代理连接失败？

1. 确认服务器 IP 地址正确
2. 确认服务器端口（默认 8080）可访问
3. 检查 `-key` 参数是否一致（服务器和代理都使用 `public`）
4. 查看服务器日志获取错误信息

### ❓ 如何后台运行？

使用 `screen` 或 `nohup`：

```bash
# 使用 screen
screen -S serverstatus
./data-server-linux-amd64 -key public -port 8080
# 按 Ctrl+A 再按 D 退出

# 使用 nohup
nohup ./data-server-linux-amd64 -key public -port 8080 > server.log 2>&1 &
```

### ❓ 如何设置为系统服务？

创建 systemd 服务文件：

```bash
# 创建服务文件
sudo nano /etc/systemd/system/serverstatus.service
```

```ini
[Unit]
Description=ServerStatus Monitor
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/serverstatus
ExecStart=/opt/serverstatus/data-server -port 8080 -key public
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# 启用并启动服务
sudo systemctl enable serverstatus
sudo systemctl start serverstatus
sudo systemctl status serverstatus
```

### ❓ Redis 连接失败？

ServerStatus 在 Redis 不可用时会自动降级到内存缓存：

```
2025/10/13 17:35:51 Redis connection failed, using in-memory cache only
```

这是正常的，系统会继续工作。要启用 Redis：

```bash
# 安装 Redis
sudo apt-get install redis-server

# 启动 Redis
sudo systemctl start redis

# 配置 ServerStatus 使用 Redis
./data-server -redis-addr localhost:6379
```

## 📚 高级主题

### 使用 Nginx 的生产环境部署

```nginx
server {
    listen 80;
    server_name monitor.yourdomain.com;

    # 前端静态文件
    location / {
        root /var/www/serverstatus;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket 代理
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 自定义前端开发

ServerStatus 提供完整的 API，允许您构建自定义前端：

#### React 示例

```jsx
import React, { useState, useEffect } from 'react';

function ServerMonitor() {
    const [servers, setServers] = useState([]);
    
    useEffect(() => {
        const fetchServers = async () => {
            const response = await fetch('http://localhost:8080/api/servers');
            setServers(await response.json());
        };
        
        fetchServers();
        const interval = setInterval(fetchServers, 3000);
        return () => clearInterval(interval);
    }, []);
    
    return (
        <div>
            {servers.map(server => (
                <div key={server.hostname} className="server-card">
                    <h3>{server.hostname}</h3>
                    <div>CPU: {server.cpu_percent}%</div>
                    <div>内存: {server.memory_percent}%</div>
                </div>
            ))}
        </div>
    );
}
```

#### Vue.js 示例

```vue
<template>
  <div class="server-grid">
    <div v-for="server in servers" :key="server.hostname" class="server-card">
      <h3>{{ server.hostname }}</h3>
      <div>CPU: {{ server.cpu_percent }}%</div>
      <div>内存: {{ server.memory_percent }}%</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const servers = ref([])
let interval = null

const fetchServers = async () => {
  const response = await fetch('http://localhost:8080/api/servers')
  servers.value = await response.json()
}

onMounted(() => {
  fetchServers()
  interval = setInterval(fetchServers, 3000)
})

onUnmounted(() => {
  clearInterval(interval)
})
</script>
```

### 多项目隔离

不同项目的完整数据隔离：

```bash
# 步骤 1：为项目 A 生成访问密钥
curl -X POST http://server:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "your-secret", "project_key": "project-a"}'

# 响应: {"access_key": "abc123xyz789"}

# 步骤 2：使用项目密钥部署代理
./monitor-agent -url http://server:8080/api/data -key project-a

# 步骤 3：访问项目面板
http://server:8080/?key=abc123xyz789
```

## 📖 更多文档

- **[英文完整文档](README.md)** - 完整英文文档
- **[详细指南](docs/README_zh.md)** - 包含示例的综合指南
- **[架构设计](CLAUDE.md)** - 系统架构和技术实现
- **[开发路线图](ROADMAP.md)** - 项目规划和未来特性
- **[演示指南](docs/DEMO.md)** - 前后端分离演示
- **[访问密钥演示](docs/ACCESS_KEY_DEMO.md)** - 访问密钥功能指南
- **[用户资源](docs/USER_RESOURCES_README.md)** - 按用户资源监控

## 🤝 参与贡献

欢迎各种形式的贡献！

- 🐛 **报告 Bug** - [提交 Issue](https://github.com/MyDailyCloud/ServerStatus/issues)
- ✨ **功能建议** - [功能请求](https://github.com/MyDailyCloud/ServerStatus/issues/new)
- 📝 **改进文档** - 帮助我们改进文档
- 🌍 **多语言支持** - 添加更多语言
- 🎨 **UI/UX 设计** - 让界面更加美观

### 开发流程

```bash
# 1. Fork 项目到你的 GitHub
# 2. 克隆到本地
git clone https://github.com/yourusername/ServerStatus.git
cd ServerStatus

# 3. 创建功能分支
git checkout -b feature/awesome-feature

# 4. 本地开发测试
cd data-server && go run main.go &
cd ../frontend-ui && python3 -m http.server 3000

# 5. 提交更改
git add .
git commit -m "Add awesome feature"
git push origin feature/awesome-feature

# 6. 在 GitHub 创建 Pull Request
```

## 📄 开源协议

本项目基于 MIT 协议开源，可自由使用和修改。

## 🙏 致谢

感谢以下项目和贡献者：

- [Go](https://golang.org/) - 后端编程语言
- [Chart.js](https://www.chartjs.org/) - 图表库
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket 库
- 所有给予 Star 的贡献者和用户

## 📞 联系方式

- 🐛 **Bug 反馈**: [GitHub Issues](https://github.com/MyDailyCloud/ServerStatus/issues)
- 💬 **讨论**: [GitHub Discussions](https://github.com/MyDailyCloud/ServerStatus/discussions)
- 📧 **商务合作**: admin@mydailycloud.com
- 🌐 **官方网站**: https://serverstatus.mydailycloud.com

---

<div align="center">

### 🌟 如果这个项目对您有帮助，请给个 Star 支持一下！🌟

**让服务器监控变得简单而美好** ❤️

[⭐ Star](https://github.com/MyDailyCloud/ServerStatus) • [🍴 Fork](https://github.com/MyDailyCloud/ServerStatus/fork) • [📢 分享](https://twitter.com/intent/tweet?text=发现了一个超棒的服务器监控项目！&url=https://github.com/MyDailyCloud/ServerStatus)

</div>
