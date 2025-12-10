# 🖥️ ServerStatus - 服务器监控神器

<div align="center">

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Linux%20|%20Windows%20|%20macOS-lightgrey.svg)](README.md)

**⚡ 3分钟部署 • 🌈 颜值超高 • 📊 功能齐全 • 🔧 超易定制**

[快速开始](#-3分钟快速开始) • [功能特性](#-功能特性) • [安装部署](#-安装部署) • [English](README.md)

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
- **⚙️ API地址保存**：自动记住设置，下次直接使用

## 🛠️ 快速部署指南

### 🚀 方式一：一键安装（最简单）

```bash
# 下载一键安装脚本
curl -L https://raw.githubusercontent.com/MyDailyCloud/ServerStatus/main/install.sh | bash
```

### 📋 方式二：手动安装

#### 1. 安装监控服务器

```bash
# 下载程序
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/data-server-linux
chmod +x data-server-linux

# 启动服务器
./data-server-linux -port 8080
```

#### 2. 添加服务器监控

在每台服务器上运行：

```bash
# 下载监控客户端
wget https://github.com/MyDailyCloud/ServerStatus/releases/latest/download/monitor-agent-linux
chmod +x monitor-agent-linux

# 启动监控（替换成你的面板地址）
./monitor-agent-linux -server "http://你的面板地址:8080" -key public
```

#### 3. 自定义前端（可选）

```bash
# 下载前端源码
git clone https://github.com/MyDailyCloud/ServerStatus.git
cd ServerStatus/frontend-ui

# 修改配置
vi js/config.js  # 设置API地址

# 部署到Web服务器
cp -r * /var/www/html/
```

### 🐳 方式三：Docker部署

```bash
# 启动监控服务器
docker run -d -p 8080:8080 mydailycloud/serverstatus:latest

# 在被监控服务器上启动代理
docker run -d mydailycloud/serverstatus-agent:latest \
  -server "http://面板地址:8080" -key public
```

## 🔧 配置说明

### 服务器配置

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

### 前端设置

现在支持在网页界面直接设置API地址，并自动保存：

1. 点击右上角的 🔑 按钮
2. 输入后端API地址
3. 勾选"保存到浏览器"选项
4. 点击"测试连接"确认可用
5. 点击"应用配置"完成设置

下次访问会自动使用保存的地址！

## 🔐 多项目管理

不同项目的服务器可以完全隔离：

### 1. 生成项目密钥

```bash
# 为项目A生成访问密钥
curl -X POST http://你的面板:8080/api/generate-access-key \
  -H "Content-Type: application/json" \
  -d '{"server_key": "你的服务器密钥", "project_key": "project-a"}'
```

### 2. 使用项目密钥

```bash
# 项目A的服务器
./monitor-agent -server "http://面板地址:8080" -key "project-a"

# 项目B的服务器
./monitor-agent -server "http://面板地址:8080" -key "project-b"
```

### 3. 访问项目面板

```
# 项目A的专属面板
http://你的面板:8080?key=生成的访问密钥

# 公开面板（显示key为public的服务器）
http://你的面板:8080
```

## 🌟 高级功能

### 🎨 自定义UI

基于API可以开发任何前端：

```javascript
// 获取服务器数据
const response = await fetch('http://localhost:8080/api/servers');
const servers = await response.json();

// 显示服务器信息
servers.forEach(server => {
    console.log(`${server.hostname}: CPU ${server.cpu_percent}%`);
});
```

### 🐳 生产环境部署

使用Docker Compose：

```yaml
version: '3.8'
services:
  serverstatus:
    image: mydailycloud/serverstatus:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - AUTH=true
      - SERVER_KEY=your-secret-key
    restart: unless-stopped
```

### 🔒 Nginx反向代理

```nginx
server {
    listen 80;
    server_name monitor.yourdomain.com;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🔧 常见问题

### ❓ 无法访问监控面板？

检查防火墙设置：

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL  
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

### ❓ 服务器不显示数据？

检查监控代理状态：

```bash
# 查看进程
ps aux | grep monitor-agent

# 查看日志
./monitor-agent -server "http://面板地址:8080" -key public -debug
```

### ❓ 如何后台运行？

使用screen或systemd：

```bash
# 使用screen
screen -S serverstatus
./data-server
# 按 Ctrl+A 再按 D 退出

# 使用systemd
sudo cp data-server /usr/local/bin/
sudo systemctl edit --force --full serverstatus.service
# 配置服务文件
sudo systemctl enable serverstatus
sudo systemctl start serverstatus
```

### ❓ 如何自定义界面？

修改CSS变量：

```css
:root {
    --bg-primary: linear-gradient(135deg, #your-color1, #your-color2);
    --accent-color: #your-accent-color;
    --text-primary: #your-text-color;
}
```

### ❓ 如何更换API地址？

1. 使用网页设置（推荐）：点击右上角🔑按钮进行设置
2. 修改配置文件：编辑 `frontend-ui/js/config.js`
3. URL参数：访问 `http://面板地址?api=新的API地址`

## 📊 API接口文档

### 基础接口

```bash
# 获取所有服务器
GET /api/servers

# 获取服务器详情
GET /api/server/{sessionID}

# 获取设备统计
GET /api/uuid-count
```

### 项目接口

```bash
# 获取项目服务器列表
GET /api/access/{access_key}/servers

# 获取项目服务器详情  
GET /api/access/{access_key}/server/{hostname}
```

完整文档：`http://你的服务器:8080/api/docs`

## 🤝 参与贡献

欢迎所有形式的贡献！

### 如何贡献

1. **🐛 报告Bug**：[提交Issue](https://github.com/MyDailyCloud/ServerStatus/issues)
2. **✨ 建议功能**：[功能请求](https://github.com/MyDailyCloud/ServerStatus/issues/new)
3. **📝 完善文档**：帮助改进文档
4. **🎨 优化界面**：让UI更美观
5. **🌍 多语言**：添加更多语言支持

### 开发流程

```bash
# 1. Fork项目
# 2. 克隆代码
git clone https://github.com/你的用户名/ServerStatus.git

# 3. 创建分支
git checkout -b feature/新功能

# 4. 开发测试
cd data-server && go run main.go

# 5. 提交代码
git commit -m "添加新功能"
git push origin feature/新功能

# 6. 创建PR
```

## 📄 开源协议

本项目基于 MIT 协议开源，可自由使用和修改。

## 🙏 致谢

感谢所有贡献者和以下开源项目：

- [Go](https://golang.org/) - 后端开发语言
- [Chart.js](https://chartjs.org/) - 图表展示
- 所有给予Star和反馈的用户

## 📞 联系方式

- 🐛 **Bug反馈**: [GitHub Issues](https://github.com/MyDailyCloud/ServerStatus/issues)
- 💬 **功能讨论**: [GitHub Discussions](https://github.com/MyDailyCloud/ServerStatus/discussions) 
- 📧 **商务合作**: admin@mydailycloud.com
- 🌐 **官方网站**: https://serverstatus.mydailycloud.com

---

<div align="center">

### 🌟 如果这个项目对你有帮助，请给个Star支持一下！🌟

**让服务器监控变得简单而美好** ❤️

[⭐ 给个Star](https://github.com/MyDailyCloud/ServerStatus) • [🍴 Fork项目](https://github.com/MyDailyCloud/ServerStatus/fork) • [📢 分享项目](https://twitter.com/intent/tweet?text=发现了一个超棒的服务器监控项目！&url=https://github.com/MyDailyCloud/ServerStatus)

</div>