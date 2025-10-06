# 🖥️ ServerStatus - 服务器监控神器

<div align="center">

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)

**⚡ 3分钟部署 • 🌈 颜值超高 • 📊 功能齐全 • 🔧 超易定制**

[快速开始](#-3分钟快速开始) • [在线演示](https://demo.example.com) • [功能特性](#-功能特性) • [安装部署](#-安装部署)

</div>

---

## 🎯 这是什么？

ServerStatus 是一个**颜值超高、功能齐全**的服务器监控面板，让你轻松掌控所有服务器状态。

### 🌟 为什么选择 ServerStatus？

- **⚡ 超级简单**：一行命令启动，3分钟完成部署
- **🌈 颜值在线**：精美UI设计，支持亮色/暗色主题
- **📊 功能丰富**：CPU、内存、网络、GPU、温度全监控
- **🔧 易于定制**：前后端分离，API优先，随意定制
- **🌍 多语言**：支持中文/英文，国际化友好
- **📱 全平台**：响应式设计，手机电脑都完美

## 🚀 3分钟快速开始

### 第一步：启动监控面板

```bash
# 下载并启动（Linux/macOS）
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-linux -o data-server
chmod +x data-server
./data-server

# Windows用户下载 data-server-windows.exe 双击运行
```

### 第二步：添加服务器监控

在每台要监控的服务器上执行：

```bash
# 一键安装监控代理
curl -L https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-linux -o monitor-agent
chmod +x monitor-agent
./monitor-agent -server "http://你的面板地址:8080" -key public
```

### 第三步：查看监控面板

打开浏览器访问：`http://你的服务器IP:8080`

🎉 **搞定！** 现在你就有了一个专业的服务器监控面板！

## 📸 效果预览

<div align="center">

### 🌞 亮色主题
![亮色主题](docs/images/light-theme.png)

### 🌙 暗色主题  
![暗色主题](docs/images/dark-theme.png)

### 📊 详细监控
![详细监控](docs/images/detailed-monitoring.png)

</div>

## ✨ 功能特性

### 🎨 用户体验
- **🌈 精美界面**：现代化设计，赏心悦目
- **🌓 主题切换**：亮色/暗色随心选择
- **🌍 多语言**：中文/英文界面
- **📱 响应式**：手机电脑都完美适配
- **⌨️ 快捷键**：键盘操作更高效

### 📊 监控功能
- **💻 系统监控**：CPU、内存、磁盘使用率
- **🌐 网络监控**：实时网速、流量统计
- **🌡️ 温度监控**：CPU、GPU温度检测
- **🎮 GPU监控**：显卡使用率、显存占用
- **📈 历史图表**：性能趋势一目了然

### 🔧 管理功能
- **📁 服务器分组**：按项目、环境分类管理
- **🔔 智能告警**：性能异常及时提醒
- **📤 数据导出**：CSV、JSON、PDF格式
- **🔐 访问控制**：多项目隔离，安全可靠

## 🛠️ 安装部署

### 方式一：快速部署（推荐）

使用我们的一键脚本：

```bash
# 下载一键部署脚本
curl -L https://raw.githubusercontent.com/MyDailyCloud/ServerStatus/main/install.sh -o install.sh
chmod +x install.sh

# 运行安装脚本
./install.sh
```

### 方式二：手动部署

#### 1. 部署监控服务器

```bash
# 下载服务器程序
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-linux

# 启动服务器
chmod +x data-server-linux
./data-server-linux -port 8080
```

#### 2. 部署前端界面（可选）

```bash
# 下载前端文件
git clone https://github.com/MyDailyCloud/ServerStatus.git
cd ServerStatus/frontend-ui

# 配置API地址
vi js/config.js  # 修改 API_BASE_URL

# 部署到Web服务器
cp -r * /var/www/html/
```

#### 3. 添加服务器监控

在每台服务器上：

```bash
# 下载监控代理
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-linux

# 启动监控
chmod +x monitor-agent-linux
./monitor-agent-linux -server "http://监控面板地址:8080" -key public
```

### 方式三：Docker部署

```bash
# 启动监控服务器
docker run -d -p 8080:8080 --name serverstatus-server \
  mydailycloud/serverstatus:latest

# 在被监控服务器上启动代理
docker run -d --name serverstatus-agent \
  mydailycloud/serverstatus-agent:latest \
  -server "http://监控面板地址:8080" -key public
```

## 🔧 配置说明

### 服务器配置

编辑配置文件或使用命令行参数：

```bash
./data-server \
  -port 8080 \                    # 监听端口
  -auth \                         # 启用认证
  -server-key "your-secret" \     # 服务器密钥
  -storage-path "./data"          # 数据存储路径
```

### 客户端配置

```bash
./monitor-agent \
  -server "http://面板地址:8080" \  # 监控面板地址
  -key "public" \                  # 项目密钥
  -hostname "自定义名称" \          # 自定义服务器名称
  -interval 3                      # 上报间隔（秒）
```

### 前端配置

编辑 `frontend-ui/js/config.js`：

```javascript
const CONFIG = {
    API_BASE_URL: 'http://localhost:8080',  // API服务器地址
    POLL_INTERVAL: 3000,                    // 刷新间隔(毫秒)
    DEFAULT_LANGUAGE: 'auto',               // 默认语言
    DEFAULT_THEME: 'light',                 // 默认主题
    // ... 更多配置选项
};
```

## 🔐 多项目管理

ServerStatus 支持多项目隔离，不同项目的服务器数据完全分离：

### 1. 生成项目访问密钥

```bash
# 生成项目A的访问密钥
curl -X POST http://你的面板:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "你的服务器密钥", "project_key": "project-a"}'

# 返回: {"access_key": "abc123xyz789"}
```

### 2. 使用项目密钥部署

```bash
# 项目A的服务器使用项目密钥
./monitor-agent -server "http://面板地址:8080" -key "project-a"

# 项目B的服务器使用不同密钥
./monitor-agent -server "http://面板地址:8080" -key "project-b"
```

### 3. 访问项目监控

```bash
# 访问项目A的监控面板
http://你的面板:8080?key=abc123xyz789

# 访问公开面板
http://你的面板:8080
```

## 🌍 API文档

ServerStatus 提供完整的 RESTful API，方便集成和二次开发：

### 基础接口

```bash
# 获取服务器列表
GET /api/servers

# 获取服务器详情
GET /api/server/{hostname}

# 获取统计信息
GET /api/uuid-count
```

### 项目接口

```bash
# 获取项目服务器列表
GET /api/access/{access_key}/servers

# 获取项目服务器详情
GET /api/access/{access_key}/server/{hostname}
```

完整API文档：启动服务后访问 `http://你的服务器:8080/api/docs`

## 🎨 自定义开发

### React 示例

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

### Vue.js 示例

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

## 🐳 生产环境部署

### Nginx + 后端分离

```nginx
# /etc/nginx/sites-available/serverstatus
server {
    listen 80;
    server_name monitor.yourdomain.com;
    
    # 前端静态文件
    location / {
        root /var/www/serverstatus;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # API代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
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
      - REQUIRE_AUTH=true
      - SERVER_KEY=your-secret-key
    restart: unless-stopped
    
  serverstatus-frontend:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./frontend-ui:/usr/share/nginx/html
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - serverstatus-server
    restart: unless-stopped
```

### 系统服务

创建 systemd 服务文件：

```ini
# /etc/systemd/system/serverstatus.service
[Unit]
Description=ServerStatus Monitor
After=network.target

[Service]
Type=simple
User=www-data
ExecStart=/usr/local/bin/data-server -port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl enable serverstatus
sudo systemctl start serverstatus
sudo systemctl status serverstatus
```

## 🔧 常见问题

### Q: 无法连接到监控面板？
**A:** 检查防火墙设置，确保8080端口开放：
```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

### Q: 服务器数据不显示？
**A:** 检查监控代理是否正常运行：
```bash
# 查看代理状态
ps aux | grep monitor-agent

# 查看代理日志
./monitor-agent -server "http://面板地址:8080" -key public -debug
```

### Q: 如何修改监控面板的外观？
**A:** 编辑 `frontend-ui/css/style.css` 文件，修改CSS变量：
```css
:root {
    --bg-primary: your-color;
    --accent-color: your-accent;
    /* 更多自定义选项... */
}
```

### Q: 如何设置HTTPS？
**A:** 使用Nginx反向代理：
```nginx
server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
    }
}
```

## 🤝 参与贡献

我们欢迎所有形式的贡献！

### 贡献方式
- 🐛 **报告Bug**：发现问题请提交 [Issue](https://github.com/MyDailyCloud/ServerStatus/issues)
- ✨ **新功能**：有好想法请提交 [Feature Request](https://github.com/MyDailyCloud/ServerStatus/issues/new?template=feature_request.md)
- 📝 **改进文档**：帮助我们完善文档
- 🎨 **UI/UX设计**：让界面更加美观易用
- 🌍 **多语言翻译**：支持更多语言

### 开发指南

```bash
# 1. Fork 项目到你的GitHub
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

# 6. 创建 Pull Request
```

## 📄 开源协议

本项目基于 [MIT 协议](LICENSE) 开源，你可以自由使用、修改和分发。

## 🙏 致谢

感谢以下项目和贡献者：

- [Chart.js](https://www.chartjs.org/) - 图表库
- [Go](https://golang.org/) - 后端语言
- 所有贡献者们的努力付出

## 📞 联系我们

- 🐛 **Bug报告**: [GitHub Issues](https://github.com/MyDailyCloud/ServerStatus/issues)
- 💬 **功能讨论**: [GitHub Discussions](https://github.com/MyDailyCloud/ServerStatus/discussions)
- 📧 **商务合作**: admin@mydailycloud.com
- 🌐 **官方网站**: https://serverstatus.mydailycloud.com

---

<div align="center">

### 🌟 如果这个项目对你有帮助，请给个Star支持一下！🌟

**让服务器监控变得简单而美好** ❤️

[⭐ Star 项目](https://github.com/MyDailyCloud/ServerStatus) • [🍴 Fork 项目](https://github.com/MyDailyCloud/ServerStatus/fork) • [📢 分享给朋友](https://twitter.com/intent/tweet?text=发现了一个超棒的服务器监控项目！&url=https://github.com/MyDailyCloud/ServerStatus)

</div>