# ServerStatus v2.x - Next-Generation Server Monitoring System

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)
[![Release](https://img.shields.io/badge/release-v2.2.0-brightgreen.svg)](https://github.com/MyDailyCloud/ServerStatus/releases)
[![Architecture](https://img.shields.io/badge/architecture-Clean%20Architecture-green.svg)](CLAUDE.md)

Enterprise-grade distributed server monitoring solution with Clean Architecture design, real-time monitoring, multi-tenancy isolation, and high-performance caching.

[中文](README_CN.md) | English

---

## 📖 Overview

ServerStatus v2.x is an enterprise-grade lightweight server monitoring system developed in Go, based on Clean Architecture design. Compared to v1.x, v2.x has significant improvements in architecture, performance, and features.

### ✨ v2.x Core Features

**Architecture Upgrade**
- Clean Architecture layered design (Handler → Service → Repository → Models)
- Interface abstraction and dependency injection
- Modular components for easy extension and maintenance

**Performance Improvements**
- Redis + in-memory dual-layer caching system
- API response speed improved by 50-70%
- WebSocket supports 1000+ concurrent connections
- Database index optimization and connection pool management

**Feature Enhancements**
- Comprehensive monitoring: CPU, memory, disk, network, GPU, temperature, user resources
- Multi-GPU support (NVIDIA multi-card monitoring)
- WebSocket real-time data push
- Multi-tenant data isolation (based on ProjectKey)
- Dual-key authentication system (ServerKey + ProjectKey)
- SQLite historical data storage
- CSV/JSON data export

**Cross-platform and Deployment**
- Supports Linux, Windows, macOS (x86_64/ARM64)
- Docker/Docker Compose one-click deployment
- Kubernetes deployment support (planned)

---

## 📊 Version Comparison (v1.x vs v2.x)

| Feature | v1.x | v2.x |
|---|---|---|
| **Architecture** | Monolithic | Clean Architecture (Layered) |
| **Caching** | None | Redis + In-memory Dual Cache |
| **Real-time Push** | HTTP Polling | WebSocket Real-time Push |
| **Multi-GPU Support** | Basic Single Card | Complete Multi-card Monitoring |
| **Data Persistence** | None | SQLite Historical Data Storage |
| **Authentication** | Single Key | Dual Key + Access Token |
| **Test Coverage** | None | 60%+ (Target) |
| **API Response** | Baseline | 50-70% Improvement |
| **Concurrent Connections** | 100+ | 1000+ |
| **Configuration** | Command Line | Config File + Env Vars + Hot Reload |

---

## 🚀 Quick Start

### Method 1: Automated Installation Script (Recommended)

```bash
# Download installation script
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/install.sh -o install.sh
chmod +x install.sh

# Install server (interactive)
./install.sh server

# Install monitoring client (interactive)
./install.sh client
```

Script features:
- Auto-detect platform and architecture
- Download corresponding binaries
- Generate configuration files
- Optional system service registration

### Method 2: Manual Installation

**Server Deployment (Example: Linux x86_64)**

```bash
# Download server
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-linux-amd64
chmod +x data-server-linux-amd64

# Start server
./data-server-linux-amd64 -key public -port 8080
```

**Monitoring Client Deployment (on monitored hosts)**

```bash
# Download client
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-linux-amd64
chmod +x monitor-agent-linux-amd64

# Start client
./monitor-agent-linux-amd64 -url http://<server-ip>:8080/api/data -key public
```

**Access Monitoring Dashboard**

```
http://<server-ip>:8080/?key=public
```

### Method 3: Docker/Docker Compose

```bash
cd deploy
docker-compose up -d

# Access dashboard
http://localhost:8080/?key=public
```

---

## 🏗️ Architecture Design and Progress

### Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│              Handler Layer               │  HTTP request handling, routing, middleware
├─────────────────────────────────────────┤
│               Service Layer              │  Business logic, use case implementation
├─────────────────────────────────────────┤
│             Repository Layer             │  Data access, storage abstraction
├─────────────────────────────────────────┤
│                Models Layer              │  Domain models, data structures
└─────────────────────────────────────────┘
```

### Refactoring Progress

- ✅ **Repository Layer**: 100% complete
- 🔄 **Service Layer**: 90% complete (WebSocket service in development)
- ⏳ **Handler Layer**: 0% (planned)

**Detailed Architecture Documentation**: [CLAUDE.md](CLAUDE.md)  
**Development Roadmap**: [ROADMAP.md](ROADMAP.md)

---

## 🗺️ Development Roadmap

### v2.2.0 (Current Version)

- ✅ Clean Architecture refactoring
- ✅ Repository layer complete
- 🔄 Service layer in development
- ⏳ Handler layer planned

### v2.3.0 (Enterprise Features)

- Alert system (Email, Webhook, DingTalk, WeChat Work)
- User permission management
- Audit logs
- Data backup and recovery
- Multi-database support (PostgreSQL, MySQL)

### v2.4.0 (Cloud Native)

- Kubernetes native support
- Prometheus integration
- Grafana dashboard
- Distributed tracing
- Service mesh support

**Complete Roadmap**: [ROADMAP.md](ROADMAP.md)

---

## 👥 Contributing

### Development Environment

- Go 1.21+
- Redis 7+
- SQLite 3+

### Commit Convention

Follow Conventional Commits specification:

```
feat: New feature
fix: Bug fix
docs: Documentation update
style: Code formatting
refactor: Code refactoring
test: Testing related
chore: Build/toolchain related
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

---

If this project helps you, please ⭐ Star to support!
