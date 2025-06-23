# ServerStatus Monitor Frontend

独立的前端项目，通过API与后端服务器通信。

## 快速开始

### 1. 启动后端API服务器
```bash
cd ../data-server
./data-server -port 8080
```

### 2. 启动前端开发服务器
```bash
# 使用Python内置服务器
npm run dev
# 或
python3 -m http.server 3000

# 访问 http://localhost:3000
```

### 3. 生产部署
```bash
# 部署到Nginx、Apache等Web服务器
cp -r * /var/www/html/

# 或使用Docker
docker run -d -p 80:80 -v $(pwd):/usr/share/nginx/html nginx
```

## 配置

编辑 `js/config.js` 文件配置API服务器地址：

```javascript
const CONFIG = {
    API_BASE_URL: 'http://localhost:8080',  // API服务器地址
    POLL_INTERVAL: 3000,                    // 轮询间隔(毫秒)
    // ... 其他配置
};
```

## 功能特性

- ✅ 实时服务器监控
- ✅ 网络速度和流量监控
- ✅ GPU监控支持
- ✅ 多语言支持 (中文/英文)
- ✅ 主题切换 (亮色/暗色)
- ✅ 服务器分组管理
- ✅ 数据导出功能
- ✅ 键盘快捷键
- ✅ 响应式设计

## 自定义开发

### API集成
```javascript
// 获取服务器列表
fetch(`${API_BASE_URL}/api/servers`)
    .then(response => response.json())
    .then(servers => {
        // 处理服务器数据
    });

// 获取服务器详情
fetch(`${API_BASE_URL}/api/server/${hostname}`)
    .then(response => response.json())
    .then(serverInfo => {
        // 处理服务器详情
    });
```

### 扩展开发
- 修改 `css/style.css` 自定义样式
- 编辑 `js/app.js` 添加新功能
- 更新 `index.html` 调整页面结构

## 技术栈

- HTML5 + CSS3 + JavaScript (原生)
- Chart.js 图表库
- 响应式设计
- RESTful API

## 部署选项

### 方式1: 静态Web服务器
```bash
# Nginx配置示例
server {
    listen 80;
    server_name yourdomain.com;
    root /path/to/frontend-ui;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # API代理 (可选)
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 方式2: Docker部署
```dockerfile
FROM nginx:alpine
COPY . /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 方式3: CDN部署
直接上传到任何支持静态文件的CDN或云存储服务。

## 浏览器支持

- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+

## 开发指南

### 添加新的监控指标
1. 确保后端API返回新的数据字段
2. 在 `js/app.js` 中添加数据处理逻辑
3. 更新显示模板和样式

### 添加新语言
1. 在 `js/app.js` 的 `getTranslations()` 方法中添加新语言
2. 更新语言切换按钮

### 自定义主题
1. 编辑 `css/style.css` 中的CSS变量
2. 添加新的主题类

## 问题排查

### 常见问题

1. **无法连接API服务器**
   - 检查 `js/config.js` 中的API地址配置
   - 确认后端服务器已启动
   - 检查CORS设置

2. **数据不显示**
   - 打开浏览器开发者工具查看网络请求
   - 检查API返回的数据格式
   - 确认客户端已向服务器上报数据

3. **页面显示异常**
   - 清除浏览器缓存
   - 检查控制台错误信息
   - 确认所有资源文件加载正常

## 贡献

欢迎提交Issue和Pull Request来改进这个项目！

## 许可证

MIT License