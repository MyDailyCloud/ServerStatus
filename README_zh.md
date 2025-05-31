# 🚀 ServerStatus Monitor

![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Active-brightgreen?style=for-the-badge)

**🔥 轻量级、强大且现代化的GPU和系统监控解决方案 🔥**

实时监控 • Web仪表板 • 多服务器支持 • 企业级安全

[English](README.md) | **中文**

---

## ✨ 功能特性

<table>
<tr>
<td width="50%">

### 🖥️ **实时监控**
- CPU、内存、磁盘、网络使用情况
- GPU利用率和温度
- 系统信息和健康状态
- 历史数据追踪

### 🌐 **现代化Web界面**
- 美观、响应式仪表板
- 实时图表和图形
- 多语言支持
- 深色/浅色主题

</td>
<td width="50%">

### 🔐 **企业级安全**
- 双密钥认证
- 基于项目的访问控制
- 基于令牌的API访问
- 安全数据传输

### 🚀 **易于部署**
- 单二进制文件部署
- 跨平台支持
- 最小资源占用
- Docker就绪

</td>
</tr>
</table>

## 🎯 快速开始

### 🌐 选项1：使用我们的托管服务（推荐）

**🚀 无需设置！直接使用我们的 ServerStatus 托管服务：**

- **仪表板**: [https://serverstatus.ltd](https://serverstatus.ltd)
- **API 端点**: `https://serverstatus.ltd/api/data`

只需下载监控代理并连接到我们的服务：

```bash
# 下载监控代理
# Linux
curl -L https://release.serverstatus.ltd/monitor-agent-linux -o monitor-agent && chmod +x monitor-agent

# macOS
curl -L https://release.serverstatus.ltd/monitor-agent-darwin -o monitor-agent && chmod +x monitor-agent

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://release.serverstatus.ltd/monitor-agent.exe" -OutFile "monitor-agent.exe"

# 开始监控（连接到我们的托管服务）
./monitor-agent -url https://serverstatus.ltd/api/data -key your-project-key
```

**🌐 访问您的仪表板**: 访问 [https://serverstatus.ltd](https://serverstatus.ltd) 查看您的服务器状态！

---

### 🏠 选项2：自托管部署

**下载预构建二进制文件**

```bash
# 从GitHub Releases下载
# 访问：https://github.com/MyDailyCloud/ServerStatus/releases

# Linux
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-linux -o monitor-agent && chmod +x monitor-agent
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-linux -o data-server && chmod +x data-server

# macOS
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-darwin -o monitor-agent && chmod +x monitor-agent
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-darwin -o data-server && chmod +x data-server

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent.exe" -OutFile "monitor-agent.exe"
Invoke-WebRequest -Uri "https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server.exe" -OutFile "data-server.exe"
```

**从源码构建**

```bash
git clone https://github.com/MyDailyCloud/ServerStatus.git
cd ServerStatus
go build -o release/data-server ./data-server
go build -o release/monitor-agent ./monitor-agent
```

**启动您自己的服务器**

```bash
# 启动数据服务器
./data-server

# 服务器将在 http://localhost:8080 可用
```

**将代理部署到您的服务器**

```bash
# 启动监控代理（连接到您自己的服务器）
./monitor-agent -url http://localhost:8080/api/data -key your-project-key
```

**访问您的自托管仪表板**

打开浏览器并导航到：`http://localhost:8080`

## ⚙️ 配置

### 🔧 服务器配置

创建 `config.json` 文件：

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

**配置项说明：**
- `project_key`: 项目主密钥，用于生成访问令牌
- `server_key`: 服务器密钥，用于双密钥认证
- `host`: 服务器监听地址
- `port`: 服务器监听端口
- `require_auth`: 是否启用认证
- `data_limit`: 每台客户端数据保留条数限制
- `data_interval`: 推荐的数据上报间隔秒数

### 🔑 认证方式

<details>
<summary><strong>🔐 双密钥认证（推荐）</strong></summary>

**生成访问密钥：**
```bash
curl -X POST http://server:8080/api/generate-access-key \
     -H "Content-Type: application/json" \
     -d '{"server_key": "server-secret-key", "project_key": "project-alpha"}'
```

**启动代理：**
```bash
./monitor-agent -url http://server:8080/api/data \
                -key project-alpha \
                -server-key server-secret-key
```

**访问仪表板：**
`http://server:8080?access={accessKey}`

</details>

<details>
<summary><strong>🎫 项目密钥认证</strong></summary>

```bash
./monitor-agent -url http://server:8080/api/data -key project-alpha
```

访问：`http://server:8080?key=project-alpha`

</details>

### 🔧 客户端配置

监控代理支持以下命令行参数：

```bash
./monitor-agent [选项]

选项：
  -url string
        服务器URL (默认: "https://serverstatus.ltd/api/data")
  -key string
        项目密钥
  -server-key string
        服务器密钥（双密钥认证时使用）
  -interval duration
        上报间隔 (默认: 1s)
  -timeout duration
        请求超时 (默认: 10s)
  -config string
        配置文件路径
```

## 📊 API参考

### 📈 数据端点

| 方法 | 端点 | 描述 |
|------|------|------|
| `POST` | `/api/data` | 提交监控数据 |
| `GET` | `/api/servers` | 获取所有服务器列表 |
| `GET` | `/api/server/{hostname}` | 获取特定服务器详情 |

### 🔐 认证端点

| 方法 | 端点 | 描述 |
|------|------|------|
| `POST` | `/api/generate-access-key` | 生成访问密钥 |
| `GET` | `/api/access/{accessKey}/servers` | 通过访问密钥获取服务器 |
| `GET` | `/api/access/{accessKey}/server/{hostname}` | 通过访问密钥获取服务器 |

## 🏢 部署示例

### 🏭 企业环境设置

<details>
<summary><strong>点击展开企业配置</strong></summary>

**服务器配置：**
```json
{
  "project_key": "company-main-key-2024",
  "server_key": "enterprise-server-key",
  "host": "0.0.0.0",
  "port": "8080",
  "require_auth": true
}
```

**多环境部署：**
```bash
# 开发环境
./monitor-agent -url http://monitor.company.com:8080/api/data -key dev-team

# 生产环境
./monitor-agent -url http://monitor.company.com:8080/api/data -key production

# 运维团队
./monitor-agent -url http://monitor.company.com:8080/api/data -key ops-team
```

</details>

### 🏠 个人环境设置

```bash
# 简单启动
./data-server
./monitor-agent -key home-server

# 访问仪表板
open https://serverstatus.ltd?key=home-server
```

## 🛠️ 开发

### 📋 先决条件

- Go 1.19+
- Git

### 🔨 构建说明

```bash
# 克隆仓库
git clone https://github.com/MyDailyCloud/ServerStatus.git
cd ServerStatus

# 初始化模块
go mod init ServerStatus
go mod tidy

# 为当前平台构建
go build -o release/data-server ./data-server
go build -o release/monitor-agent ./monitor-agent

# 跨平台构建
GOOS=linux go build -o release/data-server-linux ./data-server
GOOS=darwin go build -o release/data-server-darwin ./data-server
GOOS=windows go build -o release/data-server.exe ./data-server
```


## 🐛 故障排除

<details>
<summary><strong>🔍 常见问题</strong></summary>

**连接被拒绝**
- ✅ 检查服务器是否正在运行
- ✅ 验证端口可用性
- ✅ 检查防火墙设置

**认证失败**
- ✅ 验证项目密钥
- ✅ 检查服务器密钥配置
- ✅ 验证访问令牌过期

**数据未更新**
- ✅ 确认代理正在运行
- ✅ 检查网络连接
- ✅ 查看服务器日志

</details>

### 📋 日志查看

服务器和代理都会输出详细的运行日志，包括：
- 连接状态
- 认证结果
- 错误信息
- 性能统计

## 📈 性能

| 指标 | 值 |
|------|----|
| 内存使用 | < 50MB |
| CPU使用 | < 1% |
| 网络开销 | < 1KB/s 每个代理 |
| 支持的代理数 | 1000+ |

## 🤝 贡献

我们欢迎贡献！🎉

1. 🍴 Fork 仓库
2. 🌟 创建你的功能分支 (`git checkout -b feature/amazing-feature`)
3. 💾 提交你的更改 (`git commit -m 'Add amazing feature'`)
4. 📤 推送到分支 (`git push origin feature/amazing-feature`)
5. 🔄 打开一个 Pull Request

### 📝 开发指南

- 遵循 Go 最佳实践
- 为新功能添加测试
- 更新文档
- 使用常规提交

## 🎨 自动化脚本

### Windows 批处理脚本

**启动服务器 (start-server.bat)：**
```batch
@echo off
cd /d "%~dp0"
release\data-server.exe
pause
```

**启动代理 (start-agent.bat)：**
```batch
@echo off
set PROJECT_KEY=your-project-key
cd /d "%~dp0"
release\monitor-agent.exe -url https://serverstatus.ltd/api/data -key %PROJECT_KEY%
pause
```

### Linux Shell 脚本

**启动服务器 (start-server.sh)：**
```bash
#!/bin/bash
cd "$(dirname "$0")"
./release/data-server-linux
```

**启动代理 (start-agent.sh)：**
```bash
#!/bin/bash
PROJECT_KEY="your-project-key"
cd "$(dirname "$0")"
./release/monitor-agent-linux -url https://serverstatus.ltd/api/data -key $PROJECT_KEY
```

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🌟 Star 历史

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=MyDailyCloud/ServerStatus&type=Date)](https://star-history.com/#MyDailyCloud/ServerStatus&Date)

</div>

## 💬 社区

<div align="center">

[![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-333?style=for-the-badge&logo=github)](https://github.com/MyDailyCloud/ServerStatus/discussions)

</div>

---

<div align="center">

**由 Obscura 团队用 ❤️ 制作**

*如果你觉得这个项目有帮助，请考虑给它一个 ⭐！*

[🚀 开始使用](#-快速开始) • [📖 文档](docs/) • [🐛 报告Bug](issues/) • [💡 请求功能](issues/)