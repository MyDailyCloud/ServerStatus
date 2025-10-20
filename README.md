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

## 📋 Development Task List

The following tasks are available for team members to claim. Please mark your claim status in the corresponding Issue.

### 🔄 v2.2.0 Current Version Tasks

#### Service Layer Completion (Priority: High)

- [ ] **WebSocket Service Refactoring** `data-server/internal/service/websocket.go`
  - Implement WebSocket connection management
  - Implement real-time data push mechanism
  - Implement client connection pool
  - Add message queue support
  - Unit test coverage ≥ 80%
  - Estimated effort: 3-5 days

- [ ] **Server Management Service** `data-server/internal/service/server_service.go`
  - Implement server CRUD operations
  - Implement server status management
  - Implement server grouping functionality
  - Unit test coverage ≥ 80%
  - Estimated effort: 2-3 days

- [ ] **Data Export Service** `data-server/internal/service/export_service.go`
  - Implement CSV export functionality
  - Implement JSON export functionality
  - Implement batch export
  - Add export task queue
  - Unit test coverage ≥ 80%
  - Estimated effort: 2-3 days

- [ ] **Authentication & Authorization Service** `data-server/internal/service/auth_service.go`
  - Implement dual-key authentication logic
  - Implement access token generation and validation
  - Implement permission checking mechanism
  - Unit test coverage ≥ 80%
  - Estimated effort: 2-3 days

- [ ] **Health Check Service** `data-server/internal/service/health_service.go`
  - Implement system health checks
  - Implement dependency service checks (Redis, SQLite)
  - Implement performance metrics collection
  - Unit test coverage ≥ 80%
  - Estimated effort: 1-2 days

#### Handler Layer Implementation (Priority: High)

- [ ] **HTTP Router Implementation** `data-server/internal/handler/router.go`
  - Design RESTful API routes
  - Implement route registration mechanism
  - Add route grouping functionality
  - Estimated effort: 1-2 days

- [ ] **Middleware System** `data-server/internal/handler/middleware/`
  - Implement authentication middleware
  - Implement logging middleware
  - Implement CORS middleware
  - Implement rate limiting middleware
  - Implement error recovery middleware
  - Estimated effort: 2-3 days

- [ ] **API Handler Implementation** `data-server/internal/handler/`
  - Implement server API handlers
  - Implement authentication API handlers
  - Implement data export API handlers
  - Implement health check API handlers
  - Estimated effort: 3-4 days

- [ ] **Error Handling Mechanism** `data-server/internal/handler/error.go`
  - Unified error response format
  - Implement error code definitions
  - Implement error logging
  - Estimated effort: 1 day

- [ ] **Request Validation & Response Formatting** `data-server/internal/handler/validator.go`
  - Implement request parameter validation
  - Implement response data formatting
  - Add data masking functionality
  - Estimated effort: 1-2 days

#### Testing & Documentation (Priority: Medium)

- [ ] **Unit Test Development**
  - Repository layer test supplementation
  - Service layer tests (target coverage 80%)
  - Handler layer tests (target coverage 70%)
  - Estimated effort: 5-7 days

- [ ] **Integration Test Implementation** `data-server/test/integration/`
  - API integration tests
  - WebSocket integration tests
  - Database integration tests
  - Cache integration tests
  - Estimated effort: 3-5 days

- [ ] **Performance Testing** `data-server/test/benchmark/`
  - API performance benchmarks
  - WebSocket concurrency tests
  - Cache performance tests
  - Database query performance tests
  - Estimated effort: 2-3 days

- [ ] **API Documentation Generation**
  - Generate documentation using Swagger/OpenAPI
  - Add API usage examples
  - Add error code descriptions
  - Estimated effort: 2-3 days

### 🚀 v2.3.0 Enterprise Features Tasks

#### Alert System (Priority: High)

- [ ] **Alert Rule Engine** `data-server/internal/alert/`
  - Implement threshold-based alert rules
  - Implement composite condition alerts
  - Implement alert level definitions
  - Implement alert silencing functionality
  - Estimated effort: 5-7 days

- [ ] **Alert Notification Channels**
  - Email notification implementation
  - Webhook notification implementation
  - DingTalk bot integration
  - WeChat Work bot integration
  - Slack integration
  - Estimated effort: 5-7 days

- [ ] **Alert History & Statistics**
  - Alert history recording
  - Alert statistics reports
  - Alert trend analysis
  - Estimated effort: 3-4 days

#### User Permission Management (Priority: High)

- [ ] **User Management System** `data-server/internal/user/`
  - User CRUD operations
  - User role definitions
  - User group management
  - Estimated effort: 3-5 days

- [ ] **Permission Control System** `data-server/internal/rbac/`
  - RBAC permission model implementation
  - Resource permission definitions
  - Permission checking middleware
  - Estimated effort: 4-6 days

- [ ] **Audit Logging** `data-server/internal/audit/`
  - Operation log recording
  - Audit log querying
  - Audit report generation
  - Estimated effort: 2-3 days

#### Data Backup & Recovery (Priority: Medium)

- [ ] **Automatic Backup System** `data-server/internal/backup/`
  - Scheduled backup tasks
  - Incremental backup support
  - Backup file management
  - Estimated effort: 3-4 days

- [ ] **Data Recovery Functionality**
  - Backup file restoration
  - Data consistency checking
  - Recovery progress display
  - Estimated effort: 2-3 days

#### Multi-Database Support (Priority: Medium)

- [ ] **PostgreSQL Support** `data-server/internal/repository/postgres/`
  - PostgreSQL Repository implementation
  - Data migration scripts
  - Performance optimization
  - Estimated effort: 5-7 days

- [ ] **MySQL Support** `data-server/internal/repository/mysql/`
  - MySQL Repository implementation
  - Data migration scripts
  - Performance optimization
  - Estimated effort: 5-7 days

### ☁️ v2.4.0 Cloud-Native Features Tasks

#### Kubernetes Support (Priority: High)

- [ ] **Kubernetes Operator** `k8s/operator/`
  - CRD definitions
  - Operator controller implementation
  - Auto-scaling support
  - Estimated effort: 10-15 days

- [ ] **Helm Chart** `deploy/helm/`
  - Chart template development
  - Configuration parameter definitions
  - Deployment documentation
  - Estimated effort: 3-5 days

#### Observability Integration (Priority: High)

- [ ] **Prometheus Integration** `data-server/internal/metrics/`
  - Metrics exporter implementation
  - Custom metrics definitions
  - Grafana Dashboard
  - Estimated effort: 5-7 days

- [ ] **Distributed Tracing** `data-server/internal/tracing/`
  - OpenTelemetry integration
  - Jaeger/Zipkin support
  - Trace data visualization
  - Estimated effort: 5-7 days

- [ ] **Grafana Dashboards**
  - System monitoring dashboard
  - Performance analysis dashboard
  - Alert dashboard
  - Estimated effort: 3-5 days

#### Service Mesh Support (Priority: Low)

- [ ] **Istio Integration**
  - Service mesh configuration
  - Traffic management
  - Security policies
  - Estimated effort: 7-10 days

### 🔧 Continuous Improvement Tasks

#### Performance Optimization (Priority: Medium)

- [ ] **API Performance Optimization**
  - Database query optimization
  - Cache strategy optimization
  - Concurrency handling optimization
  - Estimated effort: Ongoing

- [ ] **WebSocket Performance Optimization**
  - Connection pool optimization
  - Message queue optimization
  - Memory usage optimization
  - Estimated effort: Ongoing

#### Code Quality (Priority: Medium)

- [ ] **Code Refactoring**
  - Eliminate code duplication
  - Improve code readability
  - Optimize code structure
  - Estimated effort: Ongoing

- [ ] **Static Code Analysis**
  - golangci-lint rule refinement
  - Code security scanning
  - Dependency vulnerability scanning
  - Estimated effort: Ongoing

#### Documentation Enhancement (Priority: Medium)

- [ ] **Developer Documentation**
  - Architecture design documentation
  - API development guide
  - Contributor guide
  - Estimated effort: Ongoing

- [ ] **User Documentation**
  - Deployment guide
  - Configuration guide
  - Troubleshooting guide
  - Estimated effort: Ongoing

---

### 📝 Task Claiming Process

1. Find the corresponding task in [GitHub Issues](https://github.com/MyDailyCloud/ServerStatus/issues)
2. Comment to indicate your claim and estimated completion time
3. Fork the project and create a feature branch
4. Complete development and submit a Pull Request
5. Merge to main branch after code review approval

### 🎯 Priority Level Explanation

- **High Priority**: Affects core functionality, needs to be completed first
- **Medium Priority**: Important but not urgent, can be progressed as planned
- **Low Priority**: Nice-to-have features, can be deferred

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
