<div align="center">

# 🚀 ServerStatus Monitor

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License" />
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=for-the-badge" alt="Platform" />
  <img src="https://img.shields.io/badge/Status-Active-brightgreen?style=for-the-badge" alt="Status" />
</p>

<p align="center">
  <strong>🔥 A lightweight, powerful, and modern server & system monitoring solution 🔥</strong>
</p>

<p align="center">
  Real-time monitoring • Web dashboard • Multi-server support • Enterprise-grade security
</p>

<img src="https://user-images.githubusercontent.com/placeholder/demo.gif" alt="Demo" width="100%" />

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🖥️ **Real-time Monitoring**
- CPU, Memory, Disk, Network usage
- GPU utilization and temperature
- System information and health
- Historical data tracking

### 🌐 **Modern Web Interface**
- Beautiful, responsive dashboard
- Real-time charts and graphs
- Multi-language support
- Dark/Light theme

</td>
<td width="50%">

### 🔐 **Enterprise Security**
- Dual-key authentication
- Project-based access control
- Token-based API access
- Secure data transmission

### 🚀 **Easy Deployment**
- Single binary deployment
- Cross-platform support
- Minimal resource usage
- Docker ready

</td>
</tr>
</table>

## 🎯 Quick Start

### 📦 Installation

**Option 1: Download Pre-built Binaries**

```bash
# Linux/macOS
curl -L https://release.serverstatus.ltd/monitor-agent-linux -o monitor-agent && chmod +x monitor-agent

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://release.serverstatus.ltd/monitor-agent.exe" -OutFile "monitor-agent.exe"
```

**Option 2: Build from Source**

```bash
git clone https://github.com/MyDailyCloud/ServerStatus.git
cd ServerStatus
go build -o release/data-server ./data-server
go build -o release/monitor-agent ./monitor-agent
```

### 🚀 Launch Server

```bash
# Start the data server
./data-server

# Server will be available at http://localhost:8080
```

### 📊 Deploy Agent

```bash
# Start monitoring agent
./monitor-agent -url http://localhost:8080/api/data -key your-project-key
```

### 🌐 Access Dashboard

Open your browser and navigate to: `http://localhost:8080`

## ⚙️ Configuration

### 🔧 Server Configuration

Create a `config.json` file:

```json
{
  "project_key": "your-project-secret-key",
  "server_key": "your-server-secret-key",
  "host": "0.0.0.0",
  "port": "8080",
  "require_auth": true,
  "data_limit": 1000,
  "data_interval": 5
}
```

### 🔑 Authentication Methods

<details>
<summary><strong>🔐 Dual-Key Authentication (Recommended)</strong></summary>

**Generate Access Key:**
```bash
curl -X POST http://server:8080/api/generate-access-key \
     -H "Content-Type: application/json" \
     -d '{"server_key": "server-secret-key", "project_key": "project-alpha"}'
```

**Start Agent:**
```bash
./monitor-agent -url http://server:8080/api/data \
                -key project-alpha \
                -server-key server-secret-key
```

**Access Dashboard:**
`http://server:8080?access={accessKey}`

</details>

<details>
<summary><strong>🎫 Project Key Authentication</strong></summary>

```bash
./monitor-agent -url http://server:8080/api/data -key project-alpha
```

Access: `http://server:8080?key=project-alpha`

</details>

## 📊 API Reference

### 📈 Data Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/data` | Submit monitoring data |
| `GET` | `/api/servers` | Get all servers list |
| `GET` | `/api/server/{hostname}` | Get specific server details |

### 🔐 Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/generate-access-key` | Generate access key |
| `GET` | `/api/access/{accessKey}/servers` | Get servers by access key |
| `GET` | `/api/access/{accessKey}/server/{hostname}` | Get server by access key |

## 🏢 Deployment Examples

### 🏭 Enterprise Setup

<details>
<summary><strong>Click to expand enterprise configuration</strong></summary>

**Server Configuration:**
```json
{
  "project_key": "company-main-key-2024",
  "server_key": "enterprise-server-key",
  "host": "0.0.0.0",
  "port": "8080",
  "require_auth": true
}
```

**Multi-Environment Deployment:**
```bash
# Development
./monitor-agent -url http://monitor.company.com:8080/api/data -key dev-team

# Production
./monitor-agent -url http://monitor.company.com:8080/api/data -key production

# Operations
./monitor-agent -url http://monitor.company.com:8080/api/data -key ops-team
```

</details>

### 🏠 Personal Setup

```bash
# Simple start
./data-server
./monitor-agent -key home-server

# Access dashboard
open https://serverstatus.ltd?key=home-server
```

## 🛠️ Development

### 📋 Prerequisites

- Go 1.19+
- Git

### 🔨 Build Instructions

```bash
# Clone repository
git clone https://github.com/MyDailyCloud/ServerStatus.git
cd ServerStatus

# Initialize modules
go mod init obscura-gpu-monitor
go mod tidy

# Build for current platform
go build -o release/data-server ./data-server
go build -o release/monitor-agent ./monitor-agent

# Cross-platform builds
GOOS=linux go build -o release/data-server-linux ./data-server
GOOS=darwin go build -o release/data-server-darwin ./data-server
GOOS=windows go build -o release/data-server.exe ./data-server
```

### 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Benchmark tests
go test -bench=. ./...
```

## 🐛 Troubleshooting

<details>
<summary><strong>🔍 Common Issues</strong></summary>

**Connection Refused**
- ✅ Check if server is running
- ✅ Verify port availability
- ✅ Check firewall settings

**Authentication Failed**
- ✅ Verify project key
- ✅ Check server key configuration
- ✅ Validate access token expiry

**Data Not Updating**
- ✅ Confirm agent is running
- ✅ Check network connectivity
- ✅ Review server logs

</details>

## 📈 Performance

| Metric | Value |
|--------|-------|
| Memory Usage | < 50MB |
| CPU Usage | < 1% |
| Network Overhead | < 1KB/s per agent |
| Supported Agents | 1000+ |

## 🤝 Contributing

We love contributions! 🎉

1. 🍴 Fork the repository
2. 🌟 Create your feature branch (`git checkout -b feature/amazing-feature`)
3. 💾 Commit your changes (`git commit -m 'Add amazing feature'`)
4. 📤 Push to the branch (`git push origin feature/amazing-feature`)
5. 🔄 Open a Pull Request

### 📝 Development Guidelines

- Follow Go best practices
- Add tests for new features
- Update documentation
- Use conventional commits

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Star History

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=your-username/obscura-gpu-monitor&type=Date)](https://star-history.com/#your-username/obscura-gpu-monitor&Date)

</div>

## 💬 Community

<div align="center">

[![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-333?style=for-the-badge&logo=github)](https://github.com/your-username/obscura-gpu-monitor/discussions)
[![Discord](https://img.shields.io/badge/Discord-Join%20Us-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/your-invite)
[![Twitter](https://img.shields.io/badge/Twitter-Follow-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white)](https://twitter.com/your-handle)

</div>

---

<div align="center">

**Made with ❤️ by the Obscura Team**

*If you find this project helpful, please consider giving it a ⭐!*

[🚀 Get Started](#-quick-start) • [📖 Documentation](docs/) • [🐛 Report Bug](issues/) • [💡 Request Feature](issues/)

</div>