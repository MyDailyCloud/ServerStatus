# ServerStatus - Lightweight Server Monitoring System

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)
[![Release](https://img.shields.io/github/v/release/MyDailyCloud/ServerStatus)](https://github.com/MyDailyCloud/ServerStatus/releases)

A **simple and easy-to-use** server monitoring system that helps you keep track of your servers effortlessly. Deploy in just 3 minutes!

**[中文文档](README_CN.md)** | **English**

---

## 🎯 Introduction

ServerStatus is a lightweight server monitoring solution built with Go, supporting real-time data collection, multi-platform deployment, and flexible authentication mechanisms. No complex configuration required - download and run!

## ✨ Features

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

## 📊 Monitoring Features

### System Monitoring
- **CPU Usage** - Real-time CPU utilization and load average
- **Memory** - RAM and swap usage monitoring
- **Disk** - Disk space usage for all partitions
- **Network** - Real-time network speed and traffic statistics
- **Temperature** - CPU and GPU temperature monitoring
- **GPU** - NVIDIA GPU utilization, memory, and temperature
- **Per-User Resources** - Resource usage breakdown by user

### Dashboard Features
- **Real-time Updates** - WebSocket-based live data updates
- **Server Grouping** - Organize servers by project or environment
- **Search & Filter** - Quick search and status-based filtering
- **Smart Alerts** - Automatic alerts for resource thresholds
- **Data Export** - Export server data in CSV or JSON format
- **Theme Support** - Light and dark theme options
- **Multi-language** - English and Chinese interface

## 🔐 Authentication Modes

ServerStatus supports three authentication modes:

### 1. Public Mode (Default)

No authentication required, suitable for testing and demos:

```bash
# Server
./data-server -key public -port 8080

# Agent
./monitor-agent -url http://server:8080/api/data -key public

# Access
http://server:8080/?key=public
```

### 2. Project Key Mode

Separate servers by project using project keys:

```bash
# Agent for Project A
./monitor-agent -url http://server:8080/api/data -key project-a

# Agent for Project B
./monitor-agent -url http://server:8080/api/data -key project-b
```

### 3. Dual-Key Authentication Mode

Enterprise-grade security with server key and project key:

```bash
# Server with authentication
./data-server -key public -port 8080 -server-key "your-secret-key"

# Agent with both keys
./monitor-agent -url http://server:8080/api/data \
  -key project-a \
  -server-key "your-secret-key"

# Generate access key for frontend
curl -X POST http://server:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "your-secret-key", "project_key": "project-a"}'

# Access with generated key
http://server:8080/?key=generated-access-key
```

## 🐳 Docker Deployment

### Quick Start with Docker

```bash
# Run server
docker run -d -p 8080:8080 --name serverstatus-server \
  -v ./data:/app/data \
  mydailycloud/serverstatus:latest

# Run agent on monitored servers
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

## 🔧 Configuration

### Server Configuration

Configuration file: `server-config.json`

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

Or use command-line arguments:

```bash
./data-server \
  -port 8080 \
  -key public \
  -server-key "your-secret" \
  -auth \
  -db-path ./data/serverstatus.db
```

### Agent Configuration

Configuration file: `config.json`

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

Or use command-line arguments:

```bash
./monitor-agent \
  -url http://server:8080/api/data \
  -key public \
  -server-key "your-secret" \
  -interval 5 \
  -hostname "custom-name"
```

## 🌐 API Documentation

ServerStatus provides a complete RESTful API:

### Server Endpoints

```bash
# Get all servers
GET /api/servers

# Get server by hostname
GET /api/server/{hostname}

# Get server count
GET /api/uuid-count

# Health check
GET /api/health
```

### Project Endpoints

```bash
# Get servers for a project (with access key)
GET /api/access/{access_key}/servers

# Get server details for a project
GET /api/access/{access_key}/server/{hostname}
```

### Authentication Endpoints

```bash
# Generate access key
POST /api/generate-access-key
Content-Type: application/json
{
  "server_key": "your-server-key",
  "project_key": "project-name"
}

# Register session
POST /api/register-session
Content-Type: application/json
{
  "project_key": "public"
}
```

### Data Collection Endpoint

```bash
# Report server data (used by agent)
POST /api/data
X-Project-Key: public
X-Server-Key: your-secret (optional)
Content-Type: application/json
{
  "session_id": "uuid",
  "hostname": "server1",
  ...
}
```

### WebSocket Endpoint

```bash
# Real-time data updates
WS /ws
```

For complete API documentation, visit: `http://your-server:8080/API.md`

## 🛠️ Troubleshooting

### ❓ Port Already in Use?

Change the startup port:

```bash
./data-server-linux-amd64 -key public -port 9090
```

### ❓ Cannot Access Monitoring Dashboard?

Check if the firewall allows the port:

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

### ❓ Agent Connection Failed?

1. Verify the server IP address is correct
2. Verify the server port (default 8080) is accessible
3. Check that the `-key` parameter matches (both server and agent use `public`)
4. Check server logs for error messages

### ❓ How to Run in Background?

Use `screen` or `nohup`:

```bash
# Using screen
screen -S serverstatus
./data-server-linux-amd64 -key public -port 8080
# Press Ctrl+A then D to detach

# Using nohup
nohup ./data-server-linux-amd64 -key public -port 8080 > server.log 2>&1 &
```

### ❓ How to Set Up as System Service?

Create a systemd service file:

```bash
# Create service file
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
# Enable and start service
sudo systemctl enable serverstatus
sudo systemctl start serverstatus
sudo systemctl status serverstatus
```

### ❓ Redis Connection Failed?

ServerStatus automatically falls back to in-memory cache if Redis is unavailable:

```
2025/10/13 17:35:51 Redis connection failed, using in-memory cache only
```

This is normal and the system will continue to work. To enable Redis:

```bash
# Install Redis
sudo apt-get install redis-server

# Start Redis
sudo systemctl start redis

# Configure ServerStatus to use Redis
./data-server -redis-addr localhost:6379
```

## 📚 Advanced Topics

### Production Deployment with Nginx

```nginx
server {
    listen 80;
    server_name monitor.yourdomain.com;

    # Frontend static files
    location / {
        root /var/www/serverstatus;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket proxy
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Custom Frontend Development

ServerStatus provides a complete API, allowing you to build custom frontends:

#### React Example

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
                    <div>Memory: {server.memory_percent}%</div>
                </div>
            ))}
        </div>
    );
}
```

#### Vue.js Example

```vue
<template>
  <div class="server-grid">
    <div v-for="server in servers" :key="server.hostname" class="server-card">
      <h3>{{ server.hostname }}</h3>
      <div>CPU: {{ server.cpu_percent }}%</div>
      <div>Memory: {{ server.memory_percent }}%</div>
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

### Multi-Project Isolation

Complete data isolation for different projects:

```bash
# Step 1: Generate access key for Project A
curl -X POST http://server:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "your-secret", "project_key": "project-a"}'

# Response: {"access_key": "abc123xyz789"}

# Step 2: Deploy agents with project key
./monitor-agent -url http://server:8080/api/data -key project-a

# Step 3: Access project dashboard
http://server:8080/?key=abc123xyz789
```

## 📖 More Documentation

- **[Complete Chinese Documentation](README_CN.md)** - Full documentation in Chinese
- **[Detailed Guide](docs/README.md)** - Comprehensive guide with examples
- **[Architecture Design](CLAUDE.md)** - System architecture and technical implementation
- **[Development Roadmap](ROADMAP.md)** - Project planning and future features
- **[Demo Guide](docs/DEMO.md)** - Frontend/Backend separation demo
- **[Access Key Demo](docs/ACCESS_KEY_DEMO.md)** - Access key functionality guide
- **[User Resources](docs/USER_RESOURCES_README.md)** - Per-user resource monitoring

## 🤝 Contributing

All forms of contributions are welcome!

- 🐛 **Report Bugs** - [Submit an Issue](https://github.com/MyDailyCloud/ServerStatus/issues)
- ✨ **Feature Requests** - [Request a Feature](https://github.com/MyDailyCloud/ServerStatus/issues/new)
- 📝 **Improve Documentation** - Help us improve the docs
- 🌍 **Multi-language Support** - Add more languages
- 🎨 **UI/UX Design** - Make the interface more beautiful

### Development Workflow

```bash
# 1. Fork the project on GitHub
# 2. Clone to local
git clone https://github.com/yourusername/ServerStatus.git
cd ServerStatus

# 3. Create feature branch
git checkout -b feature/awesome-feature

# 4. Local development and testing
cd data-server && go run main.go &
cd ../frontend-ui && python3 -m http.server 3000

# 5. Commit changes
git add .
git commit -m "Add awesome feature"
git push origin feature/awesome-feature

# 6. Create Pull Request on GitHub
```

## 📄 License

This project is open-sourced under the MIT License and can be freely used and modified.

## 🙏 Acknowledgments

Thanks to the following projects and contributors:

- [Go](https://golang.org/) - Backend programming language
- [Chart.js](https://www.chartjs.org/) - Chart library
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket library
- All contributors and users who gave stars

## 📞 Contact

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/MyDailyCloud/ServerStatus/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/MyDailyCloud/ServerStatus/discussions)
- 📧 **Business**: admin@mydailycloud.com
- 🌐 **Website**: https://serverstatus.mydailycloud.com

---

<div align="center">

### 🌟 If this project helps you, please give it a Star! 🌟

**Making Server Monitoring Simple and Beautiful** ❤️

[⭐ Star](https://github.com/MyDailyCloud/ServerStatus) • [🍴 Fork](https://github.com/MyDailyCloud/ServerStatus/fork) • [📢 Share](https://twitter.com/intent/tweet?text=Check%20out%20this%20awesome%20server%20monitoring%20project!&url=https://github.com/MyDailyCloud/ServerStatus)


</div>
