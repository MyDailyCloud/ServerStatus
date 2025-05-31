class GPUMonitor {
    constructor() {
        this.servers = new Map();
        this.charts = new Map();
        this.selectedServer = null;
        this.pollInterval = null;
        this.pollFrequency = 3000; // 3秒轮询一次
        this.apiEndpoint = '/api/servers'; // 默认API端点
        this.accessKey = null;
        this.projectKey = null;
        this.language = this.detectLanguage();
        
        this.init();
    }

    detectLanguage() {
        // 检测浏览器语言
        const browserLang = navigator.language || navigator.userLanguage;
        return browserLang.startsWith('zh') ? 'zh' : 'en';
    }

    init() {
        this.parseURLParameters();
        this.setupEventListeners();
        this.startPolling();
        this.loadInitialData();
        this.initializeTutorial();
    }

    initializeTutorial() {
        // 如果有URL参数（访问密钥模式），隐藏教程
        const tutorialSection = document.getElementById('tutorial-section');
        if (this.accessKey && tutorialSection) {
            tutorialSection.style.display = 'none';
        }
        
        // 根据语言更新教程内容
        this.updateTutorialLanguage();
    }

    updateTutorialLanguage() {
        const tutorialTexts = {
            zh: {
                quickStart: '📚 快速开始',
                oneClickInstall: '🚀 一键安装监控代理',
                stepByStep: '📋 分步骤安装',
                downloadAgent: '下载代理程序',
                setPermissions: '设置执行权限',
                runInBackground: '后台运行',
                usageTips: '💡 使用提示',
                viewLogs: '查看日志',
                stopMonitor: '停止监控',
                checkStatus: '检查运行状态',
                copyBtn: '复制',
                serverCount: '台服务器',
                uuidDevices: '个UUID设备',
                toggleTutorial: '显示/隐藏教程'
            },
            en: {
                quickStart: '📚 Quick Start',
                oneClickInstall: '🚀 One-Click Install Monitor Agent',
                stepByStep: '📋 Step-by-Step Installation',
                downloadAgent: 'Download Agent',
                setPermissions: 'Set Execute Permissions',
                runInBackground: 'Run in Background',
                usageTips: '💡 Usage Tips',
                viewLogs: 'View Logs',
                stopMonitor: 'Stop Monitor',
                checkStatus: 'Check Running Status',
                copyBtn: 'Copy',
                serverCount: ' servers',
                uuidDevices: ' UUID devices',
                toggleTutorial: 'Show/Hide Tutorial'
            }
        };

        const texts = tutorialTexts[this.language];
        
        // 更新教程标题和内容
        const elements = {
            '.tutorial-section h2': texts.quickStart,
            '.tutorial-content h3:nth-of-type(1)': texts.oneClickInstall,
            '.tutorial-content h3:nth-of-type(2)': texts.stepByStep,
            '.step:nth-child(1) strong': texts.downloadAgent,
            '.step:nth-child(2) strong': texts.setPermissions,
            '.step:nth-child(3) strong': texts.runInBackground,
            '.tips h3': texts.usageTips
        };

        Object.entries(elements).forEach(([selector, text]) => {
            const element = document.querySelector(selector);
            if (element) {
                element.textContent = text;
            }
        });

        // 更新复制按钮文本
        document.querySelectorAll('.copy-btn').forEach(btn => {
            btn.textContent = texts.copyBtn;
        });

        // 更新切换按钮文本
        const toggleBtn = document.querySelector('.tutorial-toggle');
        if (toggleBtn) {
            toggleBtn.textContent = texts.toggleTutorial;
        }

        // 更新tip内容
        const tips = document.querySelectorAll('.tip');
        if (tips.length >= 4) {
            const tipTexts = {
                zh: [
                    '显示在公共面板：',
                    '查看日志：',
                    '停止监控：',
                    '检查运行状态：'
                ],
                en: [
                    'Show on Public Panel: ',
                    'View Logs: ',
                    'Stop Monitor: ',
                    'Check Status: '
                ]
            };
            
            tips.forEach((tip, index) => {
                const strong = tip.querySelector('strong');
                if (strong && tipTexts[this.language][index]) {
                    strong.textContent = tipTexts[this.language][index];
                }
            });
        }
        
        // 初始化UUID设备文本
        const uuidCountElement = document.getElementById('uuid-count');
        if (uuidCountElement && uuidCountElement.textContent.includes('0')) {
            uuidCountElement.textContent = `0${texts.uuidDevices}`;
        }
    }

    parseURLParameters() {
        const urlParams = new URLSearchParams(window.location.search);
        
        // 检查是否有access key参数
        const accessKey = urlParams.get('access') || urlParams.get('accessKey');
        if (accessKey) {
            this.accessKey = accessKey;
            this.apiEndpoint = `/api/access/${accessKey}/servers`;
            console.log('使用访问密钥模式:', accessKey);
            
            // 更新页面标题显示当前访问模式
            this.updatePageTitle('访问密钥模式');
            return;
        }
        
        // 已移除项目密钥和访问令牌支持，只保留AccessKey访问方式
        
        console.log('使用默认模式 (无认证)');
    }

    updatePageTitle(mode) {
        const header = document.querySelector('header h1');
        if (header) {
            header.innerHTML = `🖥️ ServerStatus Monitor <span style="font-size: 0.6em; color: #666; font-weight: normal;">(${mode})</span>`;
        }
    }

    setupEventListeners() {
        document.getElementById('close-charts').addEventListener('click', () => {
            this.hideCharts();
        });

        // 键盘事件
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.hideCharts();
            }
        });
    }

    startPolling() {
        console.log('开始HTTP轮询数据更新');
        this.updateConnectionStatus(true);
        
        // 立即执行一次
        this.pollServers();
        
        // 设置定时轮询
        this.pollInterval = setInterval(() => {
            this.pollServers();
        }, this.pollFrequency);
    }

    stopPolling() {
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
            console.log('停止HTTP轮询');
            this.updateConnectionStatus(false);
        }
    }

    async pollServers() {
        try {
            const response = await fetch(this.apiEndpoint);
            if (response.ok) {
                const servers = await response.json();
                this.updateServers(servers);
                this.updateConnectionStatus(true);
            } else {
                console.error('获取服务器数据失败:', response.status, response.statusText);
                if (response.status === 401) {
                    this.updateConnectionStatus(false, '认证失败');
                } else {
                    this.updateConnectionStatus(false);
                }
            }
        } catch (error) {
            console.error('轮询服务器数据失败:', error);
            this.updateConnectionStatus(false);
        }
    }

    async loadInitialData() {
        try {
            const response = await fetch(this.apiEndpoint);
            if (response.ok) {
                const servers = await response.json();
                this.updateServers(servers);
            } else {
                console.error('加载初始数据失败:', response.status, response.statusText);
                if (response.status === 401) {
                    this.showAuthError();
                }
            }
        } catch (error) {
            console.error('加载初始数据失败:', error);
        }
    }

    // 教程相关功能函数
    copyCommand(button) {
        const commandBox = button.parentElement;
        const code = commandBox.querySelector('code');
        const text = code.textContent;
        
        // 复制到剪贴板
        navigator.clipboard.writeText(text).then(() => {
            // 显示复制成功反馈
            const originalText = button.textContent;
            button.textContent = '已复制!';
            button.style.background = '#10b981';
            
            setTimeout(() => {
                button.textContent = originalText;
                button.style.background = '#667eea';
            }, 2000);
        }).catch(err => {
            console.error('复制失败:', err);
            // 降级方案：选中文本
            const range = document.createRange();
            range.selectNode(code);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);
            
            button.textContent = '已选中';
            setTimeout(() => {
                button.textContent = '复制';
            }, 2000);
        });
    }

    toggleTutorial() {
        const tutorialSection = document.getElementById('tutorial-section');
        const toggleBtn = tutorialSection.querySelector('.toggle-tutorial');
        
        if (tutorialSection.classList.contains('collapsed')) {
            tutorialSection.classList.remove('collapsed');
            toggleBtn.textContent = '收起教程';
        } else {
            tutorialSection.classList.add('collapsed');
            toggleBtn.textContent = '展开教程';
        }
    }



    updateConnectionStatus(connected, customMessage = null) {
        const statusElement = document.getElementById('connection-status');
        if (connected) {
            statusElement.textContent = '数据正常';
            statusElement.className = 'status online';
        } else {
            statusElement.textContent = customMessage || '连接异常';
            statusElement.className = 'status offline';
        }
    }

    showAuthError() {
        const serversGrid = document.getElementById('servers-grid');
        serversGrid.innerHTML = `
            <div class="auth-error" style="
                text-align: center;
                padding: 40px;
                background: #f8f9fa;
                border-radius: 8px;
                border: 2px dashed #dc3545;
                color: #dc3545;
                margin: 20px;
            ">
                <h3>🔒 认证失败</h3>
                <p>访问密钥无效或已过期，请检查URL参数是否正确。</p>
                <p style="font-size: 0.9em; color: #666; margin-top: 20px;">
                    支持的URL格式：<br>
                    • ?access=your-access-key (访问密钥模式)<br>
                    • ?key=your-team-key (API密钥模式)<br>
                    • ?token=your-token (访问令牌模式)
                </p>
            </div>
        `;
    }

    updateServers(serversData) {
        // 更新服务器数据
        serversData.forEach(server => {
            this.servers.set(server.hostname, server);
        });

        // 更新服务器计数
        const countText = this.language === 'zh' ? '台服务器' : ' servers';
        document.getElementById('server-count').textContent = `${serversData.length}${countText}`;
        
        // 加载UUID统计
        this.loadUUIDCount();

        // 渲染服务器卡片
        this.renderServerCards(serversData);

        // 如果有选中的服务器，更新其图表
        if (this.selectedServer) {
            const selectedServerData = this.servers.get(this.selectedServer);
            if (selectedServerData) {
                this.loadServerDetails(this.selectedServer);
            }
        }
    }

    renderServerCards(servers) {
        const grid = document.getElementById('servers-grid');
        
        if (servers.length === 0) {
            grid.innerHTML = '<div class="loading">等待服务器连接...</div>';
            return;
        }

        grid.innerHTML = servers.map(server => this.createServerCard(server)).join('');

        // 只有在有access key时才添加点击事件
        if (this.accessKey) {
            grid.querySelectorAll('.server-card').forEach(card => {
                card.addEventListener('click', () => {
                    const hostname = card.dataset.hostname;
                    const sessionId = card.dataset.session;
                    this.showServerDetails(hostname, sessionId);
                });
            });
        } else {
            // 没有access key时，添加视觉提示表明卡片不可点击
            grid.querySelectorAll('.server-card').forEach(card => {
                card.style.cursor = 'default';
                card.style.opacity = '0.8';
            });
        }
    }

    createServerCard(server) {
        const isOnline = server.status === 'online';
        const lastSeenText = isOnline ? '刚刚' : this.formatTimeAgo(server.last_seen);
        
        return `
            <div class="server-card ${server.status}" data-hostname="${server.hostname}" data-session="${server.session_id || ''}">
                <div class="server-header">
                    <div class="server-name">${server.hostname}${server.session_id ? ` (${server.session_id.substring(0, 8)})` : ''}</div>
                    <div class="server-status ${server.status}">${isOnline ? '在线' : '离线'}</div>
                </div>
                <div class="metrics">
                    <div class="metric">
                        <div class="metric-label">CPU</div>
                        <div class="metric-value cpu ${server.cpu_percent > 80 ? 'high' : ''}">${server.cpu_percent.toFixed(1)}%</div>
                        <div class="progress-bar">
                            <div class="progress-fill cpu ${server.cpu_percent > 80 ? 'high' : ''}" style="width: ${server.cpu_percent}%"></div>
                        </div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">内存</div>
                        <div class="metric-value memory ${server.memory_percent > 80 ? 'high' : ''}">${server.memory_percent.toFixed(1)}%</div>
                        <div class="progress-bar">
                            <div class="progress-fill memory ${server.memory_percent > 80 ? 'high' : ''}" style="width: ${server.memory_percent}%"></div>
                        </div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">磁盘</div>
                        <div class="metric-value disk ${server.disk_percent > 80 ? 'high' : ''}">${server.disk_percent.toFixed(1)}%</div>
                        <div class="progress-bar">
                            <div class="progress-fill disk ${server.disk_percent > 80 ? 'high' : ''}" style="width: ${server.disk_percent}%"></div>
                        </div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">温度</div>
                        <div class="metric-value temp ${server.max_temp > 70 ? 'high' : ''}">${server.max_temp > 0 ? server.max_temp.toFixed(1) + '°C' : 'N/A'}</div>
                        <div class="temp-details">
                            ${server.cpu_temp > 0 ? `CPU: ${server.cpu_temp.toFixed(1)}°C` : ''}
                            ${server.gpus && server.gpus.length > 0 ? 
                                server.gpus.map((gpu, index) => 
                                    gpu.temperature > 0 ? ` GPU${index + 1}: ${gpu.temperature.toFixed(1)}°C` : ''
                                ).filter(temp => temp !== '').join('') : 
                                (server.gpu_temp > 0 ? ` GPU: ${server.gpu_temp.toFixed(1)}°C` : '')
                            }
                        </div>
                    </div>
                </div>
                <div class="server-info">
                    最后更新: ${lastSeenText}
                </div>
            </div>
        `;
    }

    formatTimeAgo(timestamp) {
        const now = new Date();
        const time = new Date(timestamp);
        const diffMs = now - time;
        const diffSecs = Math.floor(diffMs / 1000);
        const diffMins = Math.floor(diffSecs / 60);
        const diffHours = Math.floor(diffMins / 60);

        if (diffSecs < 60) return `${diffSecs}秒前`;
        if (diffMins < 60) return `${diffMins}分钟前`;
        if (diffHours < 24) return `${diffHours}小时前`;
        return time.toLocaleDateString();
    }

    async showServerDetails(hostname, sessionId) {
        this.selectedServer = hostname;
        this.selectedSessionId = sessionId;
        const displayName = sessionId ? `${hostname} (${sessionId.substring(0, 8)})` : hostname;
        document.getElementById('selected-server-name').textContent = `${displayName} - 详细监控`;
        document.getElementById('charts-section').style.display = 'block';
        
        // 滚动到图表区域
        document.getElementById('charts-section').scrollIntoView({ behavior: 'smooth' });
        
        await this.loadServerDetails(hostname, sessionId);
    }

    async loadServerDetails(hostname, sessionId) {
        try {
            let url;
            if (this.accessKey && sessionId) {
                // 如果有accessKey和sessionId，使用session-based API
                url = `/api/access/${this.accessKey}/server-by-session/${sessionId}`;
            } else if (this.accessKey) {
                // 如果只有accessKey，使用hostname-based API
                url = `/api/access/${this.accessKey}/server/${hostname}`;
            } else {
                // 默认API
                url = `/api/server/${hostname}`;
            }
            
            const response = await fetch(url);
            if (response.ok) {
                const serverData = await response.json();
                this.updateCharts(serverData);
                this.updateSystemInfo(serverData.latest);
            }
        } catch (error) {
            console.error('加载服务器详情失败:', error);
        }
    }

    updateCharts(serverData) {
        if (!serverData.history || serverData.history.length === 0) {
            return;
        }

        const history = serverData.history.slice(-30); // 最近30个数据点
        const labels = history.map(item => {
            const time = new Date(item.timestamp);
            return time.toLocaleTimeString();
        });

        // CPU图表
        this.updateChart('cpu-chart', {
            labels: labels,
            datasets: [{
                label: 'CPU使用率 (%)',
                data: history.map(item => item.cpu.usage_percent),
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                tension: 0.4,
                fill: true
            }]
        });

        // 内存图表
        this.updateChart('memory-chart', {
            labels: labels,
            datasets: [{
                label: '内存使用率 (%)',
                data: history.map(item => item.memory.usage_percent),
                borderColor: '#10b981',
                backgroundColor: 'rgba(16, 185, 129, 0.1)',
                tension: 0.4,
                fill: true
            }]
        });

        // 磁盘图表
        this.updateChart('disk-chart', {
            labels: labels,
            datasets: [{
                label: '磁盘使用率 (%)',
                data: history.map(item => item.disk.usage_percent),
                borderColor: '#f59e0b',
                backgroundColor: 'rgba(245, 158, 11, 0.1)',
                tension: 0.4,
                fill: true
            }]
        });

        // GPU使用率图表
        const gpuUsageDatasets = [];
        const hasMultiGPU = history.some(item => item.gpus && item.gpus.length > 0);
        if (hasMultiGPU) {
            const maxGPUs = Math.max(...history.map(item => item.gpus ? item.gpus.length : 0));
            const gpuColors = ['#8b5cf6', '#06b6d4', '#10b981', '#f59e0b', '#ef4444', '#ec4899'];
            
            for (let i = 0; i < maxGPUs; i++) {
                if (history.some(item => item.gpus && item.gpus[i])) {
                    gpuUsageDatasets.push({
                        label: `GPU${i + 1}使用率 (%)`,
                        data: history.map(item => {
                            if (item.gpus && item.gpus[i]) {
                                return item.gpus[i].usage_percent || 0;
                            }
                            return 0;
                        }),
                        borderColor: gpuColors[i % gpuColors.length],
                        backgroundColor: gpuColors[i % gpuColors.length].replace(')', ', 0.1)').replace('rgb', 'rgba'),
                        tension: 0.4,
                        fill: true
                    });
                }
            }
        }
        
        // 控制GPU使用率图表容器的显示/隐藏
        const gpuUsageContainer = document.getElementById('gpu-usage-container');
        if (gpuUsageDatasets.length > 0) {
            gpuUsageContainer.style.display = 'block';
            this.updateChart('gpu-usage-chart', {
                labels: labels,
                datasets: gpuUsageDatasets
            });
        } else {
            gpuUsageContainer.style.display = 'none';
        }

        // GPU显存使用率图表
        const gpuMemoryDatasets = [];
        if (hasMultiGPU) {
            const maxGPUs = Math.max(...history.map(item => item.gpus ? item.gpus.length : 0));
            const gpuColors = ['#8b5cf6', '#06b6d4', '#10b981', '#f59e0b', '#ef4444', '#ec4899'];
            
            for (let i = 0; i < maxGPUs; i++) {
                if (history.some(item => item.gpus && item.gpus[i] && item.gpus[i].memory_total > 0)) {
                    gpuMemoryDatasets.push({
                        label: `GPU${i + 1}显存使用率 (%)`,
                        data: history.map(item => {
                            if (item.gpus && item.gpus[i] && item.gpus[i].memory_total > 0) {
                                return (item.gpus[i].memory_used / item.gpus[i].memory_total) * 100;
                            }
                            return 0;
                        }),
                        borderColor: gpuColors[i % gpuColors.length],
                        backgroundColor: gpuColors[i % gpuColors.length].replace(')', ', 0.1)').replace('rgb', 'rgba'),
                        tension: 0.4,
                        fill: true
                    });
                }
            }
        }
        
        // 控制GPU显存图表容器的显示/隐藏
        const gpuMemoryContainer = document.getElementById('gpu-memory-container');
        if (gpuMemoryDatasets.length > 0) {
            gpuMemoryContainer.style.display = 'block';
            this.updateChart('gpu-memory-chart', {
                labels: labels,
                datasets: gpuMemoryDatasets
            });
        } else {
            gpuMemoryContainer.style.display = 'none';
        }

        // 温度图表
        const tempDatasets = [];
        if (history.some(item => item.temperature && item.temperature.cpu_temp > 0)) {
            tempDatasets.push({
                label: 'CPU温度 (°C)',
                data: history.map(item => item.temperature ? item.temperature.cpu_temp : 0),
                borderColor: '#ef4444',
                backgroundColor: 'rgba(239, 68, 68, 0.1)',
                tension: 0.4,
                fill: false
            });
        }
        // GPU温度数据（使用之前定义的hasMultiGPU变量）
        if (hasMultiGPU) {
            // 获取最大GPU数量
            const maxGPUs = Math.max(...history.map(item => item.gpus ? item.gpus.length : 0));
            const gpuColors = ['#8b5cf6', '#06b6d4', '#10b981', '#f59e0b', '#ef4444', '#ec4899'];
            
            for (let i = 0; i < maxGPUs; i++) {
                if (history.some(item => item.gpus && item.gpus[i] && item.gpus[i].temperature > 0)) {
                    tempDatasets.push({
                        label: `GPU${i + 1}温度 (°C)`,
                        data: history.map(item => {
                            if (item.gpus && item.gpus[i] && item.gpus[i].temperature > 0) {
                                return item.gpus[i].temperature;
                            }
                            return 0;
                        }),
                        borderColor: gpuColors[i % gpuColors.length],
                        backgroundColor: gpuColors[i % gpuColors.length].replace(')', ', 0.1)').replace('rgb', 'rgba'),
                        tension: 0.4,
                        fill: false
                    });
                }
            }
        } else if (history.some(item => item.temperature && item.temperature.gpu_temp > 0)) {
            // 兼容旧版本单GPU显示
            tempDatasets.push({
                label: 'GPU温度 (°C)',
                data: history.map(item => item.temperature ? item.temperature.gpu_temp : 0),
                borderColor: '#8b5cf6',
                backgroundColor: 'rgba(139, 92, 246, 0.1)',
                tension: 0.4,
                fill: false
            });
        }
        if (history.some(item => item.temperature && item.temperature.max_temp > 0)) {
            tempDatasets.push({
                label: '最高温度 (°C)',
                data: history.map(item => item.temperature ? item.temperature.max_temp : 0),
                borderColor: '#f97316',
                backgroundColor: 'rgba(249, 115, 22, 0.1)',
                tension: 0.4,
                fill: false
            });
        }

        if (tempDatasets.length > 0) {
            this.updateChart('temp-chart', {
                labels: labels,
                datasets: tempDatasets
            }, true);
        }
    }

    updateChart(canvasId, data, isTemperature = false) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (this.charts.has(canvasId)) {
            // 如果图表已存在，更新数据而不是重新创建
            const chart = this.charts.get(canvasId);
            
            // 更新标签
            chart.data.labels = data.labels;
            
            // 更新数据集
            chart.data.datasets = data.datasets;
            
            // 应用更新
            chart.update('none'); // 使用'none'模式实现无动画更新，提供流畅效果
            return;
        }

        const options = {
            responsive: true,
            maintainAspectRatio: false,
            animation: {
                duration: 0 // 禁用动画以获得更流畅的实时更新
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return isTemperature ? value + '°C' : value + '%';
                        }
                    }
                },
                x: {
                    ticks: {
                        maxTicksLimit: 10
                    }
                }
            },
            plugins: {
                legend: {
                    display: isTemperature && data.datasets.length > 1
                }
            },
            elements: {
                point: {
                    radius: 3,
                    hoverRadius: 6
                }
            }
        };

        if (!isTemperature) {
            options.scales.y.max = 100;
        }

        const chart = new Chart(ctx, {
            type: 'line',
            data: data,
            options: options
        });

        this.charts.set(canvasId, chart);
    }

    updateSystemInfo(systemInfo) {
        if (!systemInfo) return;

        const formatBytes = (bytes) => {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        };

        const formatUptime = (seconds) => {
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            return `${days}天 ${hours}小时 ${minutes}分钟`;
        };

        const systemInfoHtml = `
            <div class="info-item">
                <span class="info-label">主机名:</span>
                <span class="info-value">${systemInfo.hostname}</span>
            </div>
            <div class="info-item">
                <span class="info-label">操作系统:</span>
                <span class="info-value">${systemInfo.os.platform} ${systemInfo.os.version}</span>
            </div>
            <div class="info-item">
                <span class="info-label">架构:</span>
                <span class="info-value">${systemInfo.os.arch}</span>
            </div>
            <div class="info-item">
                <span class="info-label">CPU型号:</span>
                <span class="info-value">${systemInfo.cpu.model_name || 'Unknown'}</span>
            </div>
            <div class="info-item">
                <span class="info-label">CPU核心:</span>
                <span class="info-value">${systemInfo.cpu.core_count}核</span>
            </div>
            <div class="info-item">
                <span class="info-label">总内存:</span>
                <span class="info-value">${formatBytes(systemInfo.memory.total)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">已用内存:</span>
                <span class="info-value">${formatBytes(systemInfo.memory.used)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">总磁盘:</span>
                <span class="info-value">${formatBytes(systemInfo.disk.total)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">已用磁盘:</span>
                <span class="info-value">${formatBytes(systemInfo.disk.used)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">运行时间:</span>
                <span class="info-value">${formatUptime(systemInfo.os.uptime)}</span>
            </div>
            ${systemInfo.gpus && systemInfo.gpus.length > 0 ? systemInfo.gpus.map((gpu, index) => `
                            <div class="info-item">
                                <span class="info-label">GPU ${index + 1}:</span>
                                <span class="info-value">${gpu.name || 'Unknown GPU'}</span>
                            </div>
                            <div class="info-item">
                                <span class="info-label">GPU ${index + 1} 显存:</span>
                                <span class="info-value">${formatBytes(gpu.memory_used)} / ${formatBytes(gpu.memory_total)} (${gpu.usage_percent ? gpu.usage_percent.toFixed(1) : '0.0'}%)</span>
                            </div>
                            <div class="info-item">
                                <span class="info-label">GPU ${index + 1} 温度:</span>
                                <span class="info-value">${gpu.temperature > 0 ? gpu.temperature.toFixed(1) + '°C' : 'N/A'}</span>
                            </div>
                            ${gpu.driver_version ? `<div class="info-item">
                                <span class="info-label">GPU ${index + 1} 驱动版本:</span>
                                <span class="info-value">${gpu.driver_version}</span>
                            </div>` : ''}
                            ${gpu.cuda_version ? `<div class="info-item">
                                <span class="info-label">GPU ${index + 1} CUDA版本:</span>
                                <span class="info-value">${gpu.cuda_version}</span>
                            </div>` : ''}`).join('') : (systemInfo.gpu && systemInfo.gpu.name !== 'Unknown GPU' ? `
            <div class="info-item">
                                <span class="info-label">GPU:</span>
                                <span class="info-value">${systemInfo.gpu.name}</span>
                            </div>
                            <div class="info-item">
                                <span class="info-label">GPU 显存:</span>
                                <span class="info-value">${formatBytes(systemInfo.gpu.memory_used)} / ${formatBytes(systemInfo.gpu.memory_total)} (${systemInfo.gpu.usage_percent ? systemInfo.gpu.usage_percent.toFixed(1) : '0.0'}%)</span>
                            </div>
                            <div class="info-item">
                                <span class="info-label">GPU 温度:</span>
                                <span class="info-value">${systemInfo.gpu.temperature > 0 ? systemInfo.gpu.temperature.toFixed(1) + '°C' : 'N/A'}</span>
                            </div>
                            ${systemInfo.gpu.driver_version ? `<div class="info-item">
                                <span class="info-label">GPU 驱动版本:</span>
                                <span class="info-value">${systemInfo.gpu.driver_version}</span>
                            </div>` : ''}
                            ${systemInfo.gpu.cuda_version ? `<div class="info-item">
                                <span class="info-label">GPU CUDA版本:</span>
                                <span class="info-value">${systemInfo.gpu.cuda_version}</span>
                            </div>` : ''}` : '')}
            <div class="info-item">
                <span class="info-label">最后更新:</span>
                <span class="info-value">${new Date(systemInfo.timestamp).toLocaleString()}</span>
            </div>
            ${systemInfo.temperature && systemInfo.temperature.cpu_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">CPU温度:</span>
                <span class="info-value">${systemInfo.temperature.cpu_temp.toFixed(1)}°C</span>
            </div>` : ''}
            ${systemInfo.temperature && systemInfo.temperature.gpu_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">GPU温度:</span>
                <span class="info-value">${systemInfo.temperature.gpu_temp.toFixed(1)}°C</span>
            </div>` : ''}
            ${systemInfo.temperature && systemInfo.temperature.max_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">最高温度:</span>
                <span class="info-value">${systemInfo.temperature.max_temp.toFixed(1)}°C</span>
            </div>` : ''}
            ${systemInfo.temperature && systemInfo.temperature.avg_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">平均温度:</span>
                <span class="info-value">${systemInfo.temperature.avg_temp.toFixed(1)}°C</span>
            </div>` : ''}
        `;

        document.getElementById('system-info').innerHTML = systemInfoHtml;
    }

    hideCharts() {
        document.getElementById('charts-section').style.display = 'none';
        this.selectedServer = null;
        
        // 清理图表
        this.charts.forEach(chart => chart.destroy());
        this.charts.clear();
    }

    async loadUUIDCount() {
        try {
            const response = await fetch('/api/uuid-count');
            if (response.ok) {
                const data = await response.json();
                const tutorialTexts = {
                    zh: { uuidDevices: '个UUID设备' },
                    en: { uuidDevices: ' UUID devices' }
                };
                const uuidText = tutorialTexts[this.language].uuidDevices;
                document.getElementById('uuid-count').textContent = `${data.active_uuids}${uuidText}`;
            }
        } catch (error) {
            console.error('Failed to load UUID count:', error);
        }
    }
}

// 启动应用
document.addEventListener('DOMContentLoaded', () => {
    window.gpuMonitor = new GPUMonitor();
});

// 全局函数供HTML调用
function copyCommand(button) {
    if (window.gpuMonitor) {
        window.gpuMonitor.copyCommand(button);
    }
}

function toggleTutorial() {
    if (window.gpuMonitor) {
        window.gpuMonitor.toggleTutorial();
    }
}

function updateDownloadCommands() {
    const osSelect = document.getElementById('os-select');
    const archSelect = document.getElementById('arch-select');
    
    if (!osSelect || !archSelect) return;
    
    const os = osSelect.value;
    const arch = archSelect.value;
    
    // 构建文件名
    let fileName = 'monitor-agent';
    if (os === 'darwin') {
        fileName = arch === 'arm64' ? 'monitor-agent-darwin-arm64' : 'monitor-agent-darwin';
    } else if (os === 'linux') {
        fileName = arch === 'arm64' ? 'monitor-agent-linux-arm64' : 'monitor-agent-linux';
    } else if (os === 'windows') {
        fileName = 'monitor-agent.exe';
    }
    
    const baseUrl = 'https://release.serverstatus.ltd/';
    const downloadUrl = baseUrl + fileName;
    
    // 更新下载命令
    const downloadCommand = document.getElementById('download-command');
    if (downloadCommand) {
        if (os === 'windows') {
            downloadCommand.textContent = `Invoke-WebRequest -Uri "${downloadUrl}" -OutFile "monitor-agent.exe"`;
        } else {
            downloadCommand.textContent = `wget ${downloadUrl} -O monitor-agent`;
        }
    }
    
    // 更新运行命令
    const runCommand = document.getElementById('run-command');
    if (runCommand) {
        if (os === 'windows') {
            runCommand.textContent = '.\\monitor-agent.exe -key public';
        } else {
            runCommand.textContent = './monitor-agent -key public';
        }
    }
    
    // 更新一键安装命令
    const oneClickCommand = document.getElementById('one-click-command');
    if (oneClickCommand) {
        if (os === 'windows') {
            oneClickCommand.textContent = `Invoke-WebRequest -Uri "${downloadUrl}" -OutFile "monitor-agent.exe"; .\\monitor-agent.exe -key public`;
        } else {
            const curlCmd = `curl -L ${downloadUrl} -o monitor-agent`;
            const chmodCmd = 'chmod +x monitor-agent';
            const runCmd = './monitor-agent -key public';
            oneClickCommand.textContent = `${curlCmd} && ${chmodCmd} && ${runCmd}`;
        }
    }
}