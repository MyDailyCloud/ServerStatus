class GPUMonitor {
    constructor() {
        this.servers = new Map();
        this.charts = new Map();
        this.selectedServer = null;
        this.pollInterval = null;
        this.pollFrequency = CONFIG.POLL_INTERVAL; // 使用配置文件
        this.apiBaseUrl = CONFIG.API_BASE_URL; // API基础地址
        this.apiEndpoint = this.apiBaseUrl + '/api/servers'; // API端点
        this.accessKey = CONFIG.ACCESS_KEY;
        this.projectKey = null;
        this.language = this.detectLanguage();
        this.currentTheme = localStorage.getItem('preferred-theme') || CONFIG.DEFAULT_THEME;
        this.filteredServers = [];
        this.searchQuery = '';
        this.statusFilter = 'all';
        this.sortBy = 'hostname';
        this.alerts = [];
        this.alertThresholds = CONFIG.ALERT_THRESHOLDS; // 使用配置文件
        this.notifications = [];
        this.statsVisible = true;
        this.groups = new Map();
        this.serverGroups = new Map(); // hostname -> groupId
        this.groupFilter = 'all';
        this.viewMode = 'cards'; // 'cards' or 'grouped'
        this.selectedServers = new Set();
        this.contextMenuTarget = null;
        
        this.init();
    }

    detectLanguage() {
        // 如果配置指定了固定语言
        if (CONFIG.DEFAULT_LANGUAGE !== 'auto') {
            return CONFIG.DEFAULT_LANGUAGE;
        }
        
        // 首先检查localStorage中的用户选择
        const savedLang = localStorage.getItem('preferred-language');
        if (savedLang && ['zh', 'en'].includes(savedLang)) {
            return savedLang;
        }
        
        // 检测浏览器语言
        const browserLang = navigator.language || navigator.userLanguage;
        return browserLang.startsWith('zh') ? 'zh' : 'en';
    }

    init() {
        this.parseURLParameters();
        this.setupEventListeners();
        this.setupLanguageSwitching();
        this.setupThemeSwitching();
        this.setupSearchAndFilter();
        this.setupExportFunctionality();
        this.setupAlerts();
        this.setupStatsPanel();
        this.setupGroupManagement();
        this.setupHelpButton();
        this.requestNotificationPermission();
        this.loadGroups();
        this.applyTheme();
        this.applyLanguage();
        // 确保初始化完成后再次应用翻译
        setTimeout(() => {
            this.applyLanguage();
        }, 100);
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

    // 创建全面的翻译系统
    getTranslations() {
        return {
            zh: {
                // 页面标题和导航
                pageTitle: 'ServerStatus Monitor',
                connectionStatus: {
                    online: '数据正常',
                    offline: '连接异常',
                    authFailed: '认证失败'
                },
                serverCount: '台服务器',
                uuidDevices: '个UUID设备',
                
                // 搜索和筛选
                search: {
                    placeholder: '搜索服务器...',
                    allStatus: '所有状态',
                    onlineOnly: '仅在线',
                    offlineOnly: '仅离线',
                    sortByName: '按名称排序',
                    sortByCpu: '按CPU排序',
                    sortByMemory: '按内存排序',
                    sortByDisk: '按磁盘排序',
                    sortByTemp: '按温度排序',
                    noResults: '未找到匹配的服务器',
                    foundWithSearch: '找到 {count}/{total} 个包含 "{query}" 的服务器',
                    foundWithStatus: '显示 {count}/{total} 个{status}服务器',
                    foundWithFilter: '找到 {count}/{total} 个包含 "{query}" 且状态为{filter}的服务器'
                },
                
                // 教程部分
                tutorial: {
                    quickStart: '📚 快速开始',
                    oneClickInstall: '🚀 一键安装监控代理',
                    stepByStep: '📋 分步骤安装',
                    selectSystem: '选择您的系统和架构：',
                    operatingSystem: '操作系统：',
                    architecture: '架构：',
                    downloadAgent: '下载代理程序',
                    setPermissions: '设置执行权限',
                    runProgram: '运行程序',
                    usageTips: '💡 使用提示',
                    publicPanel: '显示在公共面板：',
                    publicPanelDesc: ' 使用 <code>-key public</code> 参数将服务器显示在此公共面板上',
                    backgroundRun: '后台运行：',
                    backgroundRunDesc: ' 使用 <code>screen</code> 或 <code>tmux</code> 等工具在后台运行程序',
                    networkMonitor: '网络监控：',
                    networkMonitorDesc: ' 实时网络速度和流量监控，带可视化指示器',
                    stopMonitor: '停止监控：',
                    stopMonitorDesc: ' <code>pkill monitor-agent</code>',
                    checkStatus: '检查运行状态：',
                    checkStatusDesc: ' <code>ps aux | grep monitor-agent</code>',
                    copyBtn: '复制',
                    copied: '已复制!',
                    selected: '已选中',
                    hideTutorial: '隐藏教程',
                    showTutorial: '显示教程'
                },
                
                // 布局控制
                layout: {
                    displayLayout: '显示布局：',
                    auto: '自动'
                },
                
                // 服务器状态
                server: {
                    online: '在线',
                    offline: '离线',
                    lastUpdate: '最后更新：',
                    justNow: '刚刚',
                    secondsAgo: '秒前',
                    minutesAgo: '分钟前',
                    hoursAgo: '小时前',
                    waiting: '等待服务器连接...',
                    detailed: '详细监控'
                },
                
                // 指标名称
                metrics: {
                    cpu: 'CPU',
                    memory: '内存',
                    disk: '磁盘',
                    temperature: '温度',
                    network: '网络',
                    cpuUsage: 'CPU使用率',
                    memoryUsage: '内存使用率',
                    diskUsage: '磁盘使用率',
                    temperatureMonitor: '温度监控',
                    networkUsage: '网络使用率',
                    networkSpeed: '网络速度',
                    networkSpeedMonitor: '网络速度监控',
                    gpuUsage: 'GPU使用率',
                    gpuMemoryUsage: 'GPU显存使用率',
                    systemInfo: '系统信息',
                    total: '总计'
                },
                
                // 系统信息
                systemInfo: {
                    hostname: '主机名：',
                    os: '操作系统：',
                    arch: '架构：',
                    cpuModel: 'CPU型号：',
                    cpuCores: 'CPU核心：',
                    totalMemory: '总内存：',
                    usedMemory: '已用内存：',
                    totalDisk: '总磁盘：',
                    usedDisk: '已用磁盘：',
                    uptime: '运行时间：',
                    lastUpdate: '最后更新：',
                    cores: '核',
                    days: '天',
                    hours: '小时',
                    minutes: '分钟'
                },
                
                // 统计和导出
                stats: {
                    performanceOverview: '性能概览',
                    avgCpu: '平均CPU',
                    avgMemory: '平均内存',
                    avgDisk: '平均磁盘',
                    avgNetworkSent: '平均发送',
                    avgNetworkRecv: '平均接收',
                    highUsage: '高负载',
                    totalLoad: '总负载',
                    critical: '严重',
                    hide: '隐藏',
                    show: '显示'
                },
                
                // 导出功能
                export: {
                    export: '导出',
                    exportCsv: '导出CSV',
                    exportJson: '导出JSON',
                    exportPdf: '导出PDF',
                    exportSuccess: '数据已成功导出',
                    exportError: '导出失败，请重试'
                },
                
                // 告警系统
                alerts: {
                    systemAlerts: '系统告警',
                    clearAll: '清除全部',
                    noAlerts: '当前没有告警',
                    highCpuUsage: 'CPU使用率过高',
                    highMemoryUsage: '内存使用率过高',
                    highDiskUsage: '磁盘使用率过高',
                    highTemperature: '温度过高',
                    serverOffline: '服务器离线',
                    newAlert: '新告警'
                },
                
                // 快捷键
                shortcuts: {
                    keyboardShortcuts: '键盘快捷键',
                    exportData: 'Ctrl/Cmd + E: 导出数据',
                    searchServers: 'Ctrl/Cmd + F: 搜索服务器',
                    toggleTheme: 'Ctrl/Cmd + D: 切换主题',
                    focusSearch: '/: 聚焦搜索框',
                    switchTheme: 'T: 切换主题',
                    switchLanguage: 'L: 切换语言',
                    toggleAlerts: 'A: 切换告警面板',
                    toggleStats: 'S: 切换统计面板',
                    switchLayout: '1-4: 切换网格布局',
                    closeDialogs: 'ESC: 关闭弹窗'
                },
                
                // 分组管理
                groups: {
                    allGroups: '所有分组',
                    ungrouped: '未分组',
                    groups: '分组',
                    manageGroups: '管理分组',
                    serverGroups: '服务器分组管理',
                    createGroup: '创建新分组',
                    groupName: '分组名称',
                    description: '描述（可选）',
                    groupColor: '分组颜色',
                    existingGroups: '已有分组',
                    assignServers: '分配服务器到分组',
                    selectGroup: '选择分组...',
                    bulkAssign: '批量分配',
                    assignToGroup: '分配到分组',
                    removeFromGroup: '从分组中移除',
                    noGroups: '尚未创建分组',
                    environment: '环境',
                    location: '位置',
                    project: '项目',
                    custom: '自定义',
                    servers: '台服务器',
                    edit: '编辑',
                    delete: '删除',
                    confirmDelete: '确定删除该分组吗？',
                    groupCreated: '分组创建成功',
                    groupDeleted: '分组已删除',
                    serversAssigned: '服务器已分配到分组',
                    sortByGroup: '按分组排序',
                    importExport: '导入/导出分组',
                    exportGroups: '导出分组',
                    importGroups: '导入分组',
                    exportHelp: '导出所有分组和分配信息为JSON文件',
                    importHelp: '从JSON文件导入分组'
                },
                
                // 访问密钥配置
                accessKey: {
                    modalTitle: '访问密钥配置',
                    currentConfig: '当前配置',
                    apiServer: 'API服务器',
                    accessMode: '访问模式',
                    currentKey: '当前密钥',
                    updateConfig: '更新配置',
                    apiServerUrl: 'API服务器地址',
                    accessKeyLabel: '访问密钥（可选）',
                    accessKeyPlaceholder: '输入私有项目的访问密钥',
                    accessKeyHelp: '留空为公开模式',
                    testConnection: '测试连接',
                    applyReload: '应用并重新加载',
                    connectionStatus: '连接状态',
                    notTested: '未测试',
                    testing: '正在测试连接...',
                    connectionSuccess: '✅ 连接成功！找到 {count} 台服务器',
                    connectionFailed: '❌ 连接失败',
                    helpExamples: '帮助和示例',
                    publicMode: '公开模式：',
                    publicModeDesc: '留空访问密钥查看所有公开服务器',
                    projectKeyMode: '项目密钥模式：',
                    projectKeyModeDesc: '输入项目密钥查看特定项目数据',
                    accessKeyMode: '访问密钥模式：',
                    accessKeyModeDesc: '输入生成的访问密钥进行安全项目访问',
                    urlParams: 'URL参数：',
                    urlParamApi: '设置API服务器',
                    urlParamKey: '设置访问密钥',
                    urlParamDebug: '启用调试模式'
                },

                // 错误信息
                errors: {
                    authFailed: '🔒 认证失败',
                    authFailedDesc: '访问密钥无效或已过期，请检查URL参数是否正确。',
                    supportedFormats: '支持的URL格式：'
                }
            },
            en: {
                // 页面标题和导航
                pageTitle: 'ServerStatus Monitor',
                connectionStatus: {
                    online: 'Online',
                    offline: 'Offline',
                    authFailed: 'Authentication Failed'
                },
                serverCount: ' servers',
                uuidDevices: ' UUID devices',
                
                // 搜索和筛选
                search: {
                    placeholder: 'Search servers...',
                    allStatus: 'All Status',
                    onlineOnly: 'Online Only',
                    offlineOnly: 'Offline Only',
                    sortByName: 'Sort by Name',
                    sortByCpu: 'Sort by CPU',
                    sortByMemory: 'Sort by Memory',
                    sortByDisk: 'Sort by Disk',
                    sortByTemp: 'Sort by Temperature',
                    noResults: 'No servers found matching your criteria',
                    foundWithSearch: 'Found {count}/{total} servers containing "{query}"',
                    foundWithStatus: 'Showing {count}/{total} {status} servers',
                    foundWithFilter: 'Found {count}/{total} servers containing "{query}" with {filter} status'
                },
                
                // 统计和导出
                stats: {
                    performanceOverview: 'Performance Overview',
                    avgCpu: 'Avg CPU',
                    avgMemory: 'Avg Memory',
                    avgDisk: 'Avg Disk',
                    avgNetworkSent: 'Avg Send',
                    avgNetworkRecv: 'Avg Recv',
                    highUsage: 'High Usage',
                    totalLoad: 'Total Load',
                    critical: 'Critical',
                    hide: 'Hide',
                    show: 'Show'
                },
                
                // 导出功能
                export: {
                    export: 'Export',
                    exportCsv: 'Export as CSV',
                    exportJson: 'Export as JSON',
                    exportPdf: 'Export as PDF',
                    exportSuccess: 'Data exported successfully',
                    exportError: 'Export failed, please try again'
                },
                
                // 告警系统
                alerts: {
                    systemAlerts: 'System Alerts',
                    clearAll: 'Clear All',
                    noAlerts: 'No alerts at this time',
                    highCpuUsage: 'High CPU Usage',
                    highMemoryUsage: 'High Memory Usage',
                    highDiskUsage: 'High Disk Usage',
                    highTemperature: 'High Temperature',
                    serverOffline: 'Server Offline',
                    newAlert: 'New Alert'
                },
                
                // 快捷键
                shortcuts: {
                    keyboardShortcuts: 'Keyboard Shortcuts',
                    exportData: 'Ctrl/Cmd + E: Export data',
                    searchServers: 'Ctrl/Cmd + F: Search servers',
                    toggleTheme: 'Ctrl/Cmd + D: Toggle theme',
                    focusSearch: '/: Focus search box',
                    switchTheme: 'T: Toggle theme',
                    switchLanguage: 'L: Switch language',
                    toggleAlerts: 'A: Toggle alerts panel',
                    toggleStats: 'S: Toggle stats panel',
                    switchLayout: '1-4: Switch grid layout',
                    closeDialogs: 'ESC: Close dialogs'
                },
                
                // 分组管理
                groups: {
                    allGroups: 'All Groups',
                    ungrouped: 'Ungrouped',
                    groups: 'Groups',
                    manageGroups: 'Manage Groups',
                    serverGroups: 'Server Groups Management',
                    createGroup: 'Create New Group',
                    groupName: 'Group name',
                    description: 'Description (optional)',
                    groupColor: 'Group color',
                    existingGroups: 'Existing Groups',
                    assignServers: 'Assign Servers to Groups',
                    selectGroup: 'Select a group...',
                    bulkAssign: 'Bulk Assign',
                    assignToGroup: 'Assign to Group',
                    removeFromGroup: 'Remove from Group',
                    noGroups: 'No groups created yet',
                    environment: 'Environment',
                    location: 'Location',
                    project: 'Project',
                    custom: 'Custom',
                    servers: 'servers',
                    edit: 'Edit',
                    delete: 'Delete',
                    confirmDelete: 'Are you sure you want to delete this group?',
                    groupCreated: 'Group created successfully',
                    groupDeleted: 'Group deleted',
                    serversAssigned: 'Servers assigned to group',
                    sortByGroup: 'Sort by Group',
                    importExport: 'Import/Export Groups',
                    exportGroups: 'Export Groups',
                    importGroups: 'Import Groups',
                    exportHelp: 'Export all groups and assignments as JSON',
                    importHelp: 'Import groups from JSON file'
                },
                
                // 教程部分
                tutorial: {
                    quickStart: '📚 Quick Start',
                    oneClickInstall: '🚀 One-Click Install Monitor Agent',
                    stepByStep: '📋 Step-by-Step Installation',
                    selectSystem: 'Select your system and architecture:',
                    operatingSystem: 'Operating System:',
                    architecture: 'Architecture:',
                    downloadAgent: 'Download Agent',
                    setPermissions: 'Set Execute Permissions',
                    runProgram: 'Run Program',
                    usageTips: '💡 Usage Tips',
                    publicPanel: 'Show on Public Panel:',
                    publicPanelDesc: ' Use <code>-key public</code> parameter to display server on this public panel',
                    backgroundRun: 'Run in Background:',
                    backgroundRunDesc: ' Use tools like <code>screen</code> or <code>tmux</code> to run the program in the background',
                    networkMonitor: 'Network Monitoring:',
                    networkMonitorDesc: ' Real-time network speed and traffic monitoring with visual indicators',
                    stopMonitor: 'Stop Monitor:',
                    stopMonitorDesc: ' <code>pkill monitor-agent</code>',
                    checkStatus: 'Check Running Status:',
                    checkStatusDesc: ' <code>ps aux | grep monitor-agent</code>',
                    copyBtn: 'Copy',
                    copied: 'Copied!',
                    selected: 'Selected',
                    hideTutorial: 'Hide Tutorial',
                    showTutorial: 'Show Tutorial'
                },
                
                // 布局控制
                layout: {
                    displayLayout: 'Display Layout:',
                    auto: 'Auto'
                },
                
                // 服务器状态
                server: {
                    online: 'Online',
                    offline: 'Offline',
                    lastUpdate: 'Last Update: ',
                    justNow: 'Just now',
                    secondsAgo: 'seconds ago',
                    minutesAgo: 'minutes ago',
                    hoursAgo: 'hours ago',
                    waiting: 'Waiting for server connection...',
                    detailed: 'Server Details'
                },
                
                // 指标名称
                metrics: {
                    cpu: 'CPU',
                    memory: 'Memory',
                    disk: 'Disk',
                    temperature: 'Temperature',
                    network: 'Network',
                    cpuUsage: 'CPU Usage',
                    memoryUsage: 'Memory Usage',
                    diskUsage: 'Disk Usage',
                    temperatureMonitor: 'Temperature Monitor',
                    networkUsage: 'Network Usage',
                    networkSpeed: 'Network Speed',
                    networkSpeedMonitor: 'Network Speed Monitor',
                    gpuUsage: 'GPU Usage',
                    gpuMemoryUsage: 'GPU Memory Usage',
                    systemInfo: 'System Information',
                    total: 'Total'
                },
                
                // 系统信息
                systemInfo: {
                    hostname: 'Hostname:',
                    os: 'Operating System:',
                    arch: 'Architecture:',
                    cpuModel: 'CPU Model:',
                    cpuCores: 'CPU Cores:',
                    totalMemory: 'Total Memory:',
                    usedMemory: 'Used Memory:',
                    totalDisk: 'Total Disk:',
                    usedDisk: 'Used Disk:',
                    uptime: 'Uptime:',
                    lastUpdate: 'Last Update:',
                    cores: 'cores',
                    days: 'days',
                    hours: 'hours',
                    minutes: 'minutes'
                },
                
                // 访问密钥配置
                accessKey: {
                    modalTitle: 'Access Key Configuration',
                    currentConfig: 'Current Configuration',
                    apiServer: 'API Server',
                    accessMode: 'Access Mode',
                    currentKey: 'Current Key',
                    updateConfig: 'Update Configuration',
                    apiServerUrl: 'API Server URL',
                    accessKeyLabel: 'Access Key (Optional)',
                    accessKeyPlaceholder: 'Enter access key for private projects',
                    accessKeyHelp: 'Leave empty for public mode',
                    testConnection: 'Test Connection',
                    applyReload: 'Apply & Reload',
                    connectionStatus: 'Connection Status',
                    notTested: 'Not tested',
                    testing: 'Testing connection...',
                    connectionSuccess: '✅ Connection successful! Found {count} servers',
                    connectionFailed: '❌ Connection failed',
                    helpExamples: 'Help & Examples',
                    publicMode: 'Public Mode:',
                    publicModeDesc: 'Leave access key empty to view all public servers',
                    projectKeyMode: 'Project Key Mode:',
                    projectKeyModeDesc: 'Enter project key to view specific project data',
                    accessKeyMode: 'Access Key Mode:',
                    accessKeyModeDesc: 'Enter generated access key for secure project access',
                    urlParams: 'URL Parameters:',
                    urlParamApi: 'Set API server',
                    urlParamKey: 'Set access key',
                    urlParamDebug: 'Enable debug mode'
                },
                
                // 错误信息
                errors: {
                    authFailed: '🔒 Authentication Failed',
                    authFailedDesc: 'Access key is invalid or expired. Please check if URL parameters are correct.',
                    supportedFormats: 'Supported URL formats:'
                }
            }
        };
    }
    
    t(key) {
        const translations = this.getTranslations();
        const keys = key.split('.');
        let value = translations[this.language];
        
        // 调试信息
        if (!value) {
            console.warn(`Translation object not found for language: ${this.language}`);
            return key;
        }
        
        for (const k of keys) {
            if (value && typeof value === 'object' && k in value) {
                value = value[k];
            } else {
                console.warn(`Translation key not found: ${key} (missing: ${k})`);
                return key; // 返回原键名如果找不到翻译
            }
        }
        
        return value || key;
    }
    
    updateAllTexts() {
        // 更新语言按钮状态
        document.querySelectorAll('.lang-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.getElementById(`lang-${this.language}`).classList.add('active');
        
        // 更新各种文本元素
        this.updateElementText('tutorial-title', this.t('tutorial.quickStart'));
        this.updateElementText('one-click-title', this.t('tutorial.oneClickInstall'));
        this.updateElementText('step-by-step-title', this.t('tutorial.stepByStep'));
        this.updateElementText('select-system-title', this.t('tutorial.selectSystem'));
        this.updateElementText('os-label', this.t('tutorial.operatingSystem'));
        this.updateElementText('arch-label', this.t('tutorial.architecture'));
        this.updateElementText('download-agent-title', this.t('tutorial.downloadAgent'));
        this.updateElementText('set-permissions-title', this.t('tutorial.setPermissions'));
        this.updateElementText('run-program-title', this.t('tutorial.runProgram'));
        this.updateElementText('usage-tips-title', this.t('tutorial.usageTips'));
        this.updateElementText('display-layout-label', this.t('layout.displayLayout'));
        this.updateElementText('auto-grid-btn', this.t('layout.auto'));
        
        // 更新图表标题
        this.updateElementText('cpu-usage-title', this.t('metrics.cpuUsage'));
        this.updateElementText('memory-usage-title', this.t('metrics.memoryUsage'));
        this.updateElementText('disk-usage-title', this.t('metrics.diskUsage'));
        this.updateElementText('temperature-title', this.t('metrics.temperatureMonitor'));
        this.updateElementText('network-speed-title', this.t('metrics.networkSpeedMonitor'));
        this.updateElementText('gpu-usage-title', this.t('metrics.gpuUsage'));
        this.updateElementText('gpu-memory-title', this.t('metrics.gpuMemoryUsage'));
        this.updateElementText('system-info-title', this.t('metrics.systemInfo'));
        
        // 更新提示内容
        this.updateTipContents();
        
        // 更新复制按钮
        document.querySelectorAll('.copy-btn').forEach(btn => {
            if (!btn.textContent.includes('!')) {
                btn.textContent = this.t('tutorial.copyBtn');
            }
        });
        
        // 更新搜索和筛选文本
        this.updateFilterTexts();
        
        // 更新其他元素文本
        this.updateStatsTexts();
        this.updateExportTexts();
        this.updateAlertsTexts();
        this.updateGroupTexts();
        this.updateAccessKeyTexts();
    }
    
    updateGroupTexts() {
        this.updateElementText('manage-groups-text', this.t('groups.groups'));
        this.updateElementText('group-modal-title', this.t('groups.serverGroups'));
        this.updateElementText('create-group-title', this.t('groups.createGroup'));
        this.updateElementText('existing-groups-title', this.t('groups.existingGroups'));
        this.updateElementText('assign-servers-title', this.t('groups.assignServers'));
        
        // 更新表单占位符
        const groupNameInput = document.getElementById('group-name');
        const groupDescInput = document.getElementById('group-description');
        if (groupNameInput) groupNameInput.placeholder = this.t('groups.groupName');
        if (groupDescInput) groupDescInput.placeholder = this.t('groups.description');
        
        // 更新按钮文本
        this.updateElementText('create-group-btn', this.t('groups.createGroup'));
        this.updateElementText('bulk-assign-btn', this.t('groups.bulkAssign'));
        this.updateElementText('export-groups-btn', this.t('groups.exportGroups'));
        this.updateElementText('import-groups-btn', this.t('groups.importGroups'));
        this.updateElementText('import-export-title', this.t('groups.importExport'));
        this.updateElementText('export-help-text', this.t('groups.exportHelp'));
        this.updateElementText('import-help-text', this.t('groups.importHelp'));
        
        // 更新筛选器
        this.updateGroupFilter();
        
        // 更新排序选项
        const sortSelect = document.getElementById('sort-by');
        if (sortSelect) {
            const options = sortSelect.querySelectorAll('option');
            if (options.length >= 6) {
                options[1].textContent = this.t('groups.sortByGroup');
            }
        }
    }
    
    updateStatsTexts() {
        this.updateElementText('stats-title', this.t('stats.performanceOverview'));
        this.updateElementText('avg-cpu-label', this.t('stats.avgCpu'));
        this.updateElementText('avg-memory-label', this.t('stats.avgMemory'));
        this.updateElementText('avg-disk-label', this.t('stats.avgDisk'));
        this.updateElementText('avg-network-sent-label', this.t('stats.avgNetworkSent'));
        this.updateElementText('avg-network-recv-label', this.t('stats.avgNetworkRecv'));
        this.updateElementText('high-usage-label', this.t('stats.highUsage'));
        this.updateElementText('total-load-label', this.t('stats.totalLoad'));
        this.updateElementText('critical-alerts-label', this.t('stats.critical'));
        
        const toggleBtn = document.getElementById('toggle-stats');
        if (toggleBtn) {
            toggleBtn.textContent = this.statsVisible ? this.t('stats.hide') : this.t('stats.show');
        }
    }
    
    updateExportTexts() {
        this.updateElementText('export-text', this.t('export.export'));
        this.updateElementText('export-csv', this.t('export.exportCsv'));
        this.updateElementText('export-json', this.t('export.exportJson'));
        this.updateElementText('export-pdf', this.t('export.exportPdf'));
    }
    
    updateAlertsTexts() {
        this.updateElementText('alerts-title', this.t('alerts.systemAlerts'));
        this.updateElementText('clear-alerts', this.t('alerts.clearAll'));
        this.updateElementText('no-alerts', this.t('alerts.noAlerts'));
    }
    
    setupSearchAndFilter() {
        const searchInput = document.getElementById('server-search');
        const clearBtn = document.getElementById('search-clear');
        const statusFilter = document.getElementById('status-filter');
        const sortSelect = document.getElementById('sort-by');
        
        // 搜索输入
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value.toLowerCase();
                this.updateClearButton();
                this.applyFilters();
            });
        }
        
        // 清除按钮
        if (clearBtn) {
            clearBtn.addEventListener('click', () => {
                searchInput.value = '';
                this.searchQuery = '';
                this.updateClearButton();
                this.applyFilters();
            });
        }
        
        // 状态筛选
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                this.statusFilter = e.target.value;
                this.applyFilters();
            });
        }
        
        // 排序选择
        if (sortSelect) {
            sortSelect.addEventListener('change', (e) => {
                this.sortBy = e.target.value;
                this.applyFilters();
            });
        }
        
        // 更新按钮文本
        this.updateFilterTexts();
    }
    
    updateFilterTexts() {
        const searchInput = document.getElementById('server-search');
        const statusFilter = document.getElementById('status-filter');
        const sortSelect = document.getElementById('sort-by');
        
        if (searchInput) {
            searchInput.placeholder = this.t('search.placeholder');
        }
        
        if (statusFilter) {
            const options = statusFilter.querySelectorAll('option');
            if (options.length >= 3) {
                options[0].textContent = this.t('search.allStatus');
                options[1].textContent = this.t('search.onlineOnly');
                options[2].textContent = this.t('search.offlineOnly');
            }
        }
        
        if (sortSelect) {
            const options = sortSelect.querySelectorAll('option');
            if (options.length >= 5) {
                options[0].textContent = this.t('search.sortByName');
                options[1].textContent = this.t('search.sortByCpu');
                options[2].textContent = this.t('search.sortByMemory');
                options[3].textContent = this.t('search.sortByDisk');
                options[4].textContent = this.t('search.sortByTemp');
            }
        }
    }
    
    updateClearButton() {
        const clearBtn = document.getElementById('search-clear');
        if (clearBtn) {
            if (this.searchQuery.length > 0) {
                clearBtn.classList.add('visible');
            } else {
                clearBtn.classList.remove('visible');
            }
        }
    }
    
    applyFilters() {
        if (!this.servers || this.servers.size === 0) {
            return;
        }
        
        let serversArray = Array.from(this.servers.values());
        
        // 应用搜索筛选
        if (this.searchQuery) {
            serversArray = serversArray.filter(server => {
                const searchInHostname = server.hostname.toLowerCase().includes(this.searchQuery);
                const searchInSession = server.session_id && server.session_id.toLowerCase().includes(this.searchQuery);
                
                // 在分组名中搜索
                const groupId = this.serverGroups.get(server.hostname);
                const group = groupId ? this.groups.get(groupId) : null;
                const searchInGroup = group && group.name.toLowerCase().includes(this.searchQuery);
                
                return searchInHostname || searchInSession || searchInGroup;
            });
        }
        
        // 应用状态筛选
        if (this.statusFilter !== 'all') {
            serversArray = serversArray.filter(server => server.status === this.statusFilter);
        }
        
        // 应用分组筛选
        if (this.groupFilter !== 'all') {
            if (this.groupFilter === 'ungrouped') {
                serversArray = serversArray.filter(server => !this.serverGroups.has(server.hostname));
            } else {
                serversArray = serversArray.filter(server => this.serverGroups.get(server.hostname) === this.groupFilter);
            }
        }
        
        // 应用排序
        serversArray.sort((a, b) => {
            switch (this.sortBy) {
                case 'hostname':
                    return a.hostname.localeCompare(b.hostname);
                case 'group':
                    const groupA = this.getServerGroupName(a.hostname);
                    const groupB = this.getServerGroupName(b.hostname);
                    return groupA.localeCompare(groupB);
                case 'cpu':
                    return b.cpu_percent - a.cpu_percent;
                case 'memory':
                    return b.memory_percent - a.memory_percent;
                case 'disk':
                    return b.disk_percent - a.disk_percent;
                case 'temperature':
                    return b.max_temp - a.max_temp;
                default:
                    return 0;
            }
        });
        
        this.filteredServers = serversArray;
        this.renderFilteredServers();
        this.updateSearchResults();
    }
    
    getServerGroupName(hostname) {
        const groupId = this.serverGroups.get(hostname);
        const group = groupId ? this.groups.get(groupId) : null;
        return group ? group.name : this.t('groups.ungrouped');
    }
    
    renderFilteredServers() {
        const grid = document.getElementById('servers-grid');
        
        if (this.filteredServers.length === 0) {
            if (this.searchQuery || this.statusFilter !== 'all' || this.groupFilter !== 'all') {
                grid.innerHTML = `<div class="no-results">${this.t('search.noResults')}</div>`;
            } else {
                grid.innerHTML = `<div class="loading">${this.t('server.waiting')}</div>`;
            }
            return;
        }
        
        if (this.viewMode === 'grouped') {
            this.renderGroupedServers(grid);
        } else {
            this.renderCardServers(grid);
        }
    }
    
    renderCardServers(grid) {
        grid.innerHTML = this.filteredServers.map(server => this.createServerCard(server)).join('');
        this.addServerCardListeners();
    }
    
    renderGroupedServers(grid) {
        // 按分组组织服务器
        const groupedServers = new Map();
        const ungroupedServers = [];
        
        this.filteredServers.forEach(server => {
            const groupId = this.serverGroups.get(server.hostname);
            if (groupId && this.groups.has(groupId)) {
                if (!groupedServers.has(groupId)) {
                    groupedServers.set(groupId, []);
                }
                groupedServers.get(groupId).push(server);
            } else {
                ungroupedServers.push(server);
            }
        });
        
        let html = '';
        
        // 渲染分组
        for (const [groupId, servers] of groupedServers) {
            const group = this.groups.get(groupId);
            if (group && servers.length > 0) {
                html += this.createGroupSection(group, servers);
            }
        }
        
        // 渲染未分组服务器
        if (ungroupedServers.length > 0) {
            html += this.createUngroupedSection(ungroupedServers);
        }
        
        grid.innerHTML = html;
        this.addServerCardListeners();
        this.addGroupListeners();
    }
    
    createGroupSection(group, servers) {
        const isCollapsed = localStorage.getItem(`group-collapsed-${group.id}`) === 'true';
        const stats = this.calculateGroupStatsFromServers(servers);
        
        return `
            <div class="group-header" data-group="${group.id}">
                <div class="group-header-info">
                    <div class="group-header-color" style="background-color: ${group.color}"></div>
                    <div class="group-header-details">
                        <h3>${group.name}</h3>
                        <p class="group-header-meta">${this.t(`groups.${group.type}`)} • ${servers.length} ${this.t('groups.servers')}</p>
                    </div>
                </div>
                <div class="group-header-stats">
                    <span class="group-header-stat online">${stats.onlineCount}/${stats.totalCount}</span>
                    <span class="group-header-stat">CPU: ${stats.avgCpu.toFixed(1)}%</span>
                    <span class="group-header-stat">MEM: ${stats.avgMemory.toFixed(1)}%</span>
                    ${stats.criticalCount > 0 ? `<span class="group-header-stat critical">⚠ ${stats.criticalCount}</span>` : ''}
                </div>
                <button class="group-toggle" data-group="${group.id}">
                    ${isCollapsed ? '▶' : '▼'}
                </button>
            </div>
            <div class="group-servers ${isCollapsed ? 'collapsed' : ''}" id="group-servers-${group.id}">
                ${servers.map(server => this.createServerCard(server)).join('')}
            </div>
        `;
    }
    
    createUngroupedSection(servers) {
        const isCollapsed = localStorage.getItem('group-collapsed-ungrouped') === 'true';
        const stats = this.calculateGroupStatsFromServers(servers);
        
        return `
            <div class="group-header" data-group="ungrouped">
                <div class="group-header-info">
                    <div class="group-header-color" style="background-color: #94a3b8"></div>
                    <div class="group-header-details">
                        <h3>${this.t('groups.ungrouped')}</h3>
                        <p class="group-header-meta">${servers.length} ${this.t('groups.servers')}</p>
                    </div>
                </div>
                <div class="group-header-stats">
                    <span class="group-header-stat online">${stats.onlineCount}/${stats.totalCount}</span>
                    <span class="group-header-stat">CPU: ${stats.avgCpu.toFixed(1)}%</span>
                    <span class="group-header-stat">MEM: ${stats.avgMemory.toFixed(1)}%</span>
                    ${stats.criticalCount > 0 ? `<span class="group-header-stat critical">⚠ ${stats.criticalCount}</span>` : ''}
                </div>
                <button class="group-toggle" data-group="ungrouped">
                    ${isCollapsed ? '▶' : '▼'}
                </button>
            </div>
            <div class="group-servers ${isCollapsed ? 'collapsed' : ''}" id="group-servers-ungrouped">
                ${servers.map(server => this.createServerCard(server)).join('')}
            </div>
        `;
    }
    
    addGroupListeners() {
        const groupToggles = document.querySelectorAll('.group-toggle');
        groupToggles.forEach(toggle => {
            toggle.addEventListener('click', (e) => {
                e.stopPropagation();
                const groupId = toggle.dataset.group;
                const groupServers = document.getElementById(`group-servers-${groupId}`);
                const isCollapsed = groupServers.classList.contains('collapsed');
                
                if (isCollapsed) {
                    groupServers.classList.remove('collapsed');
                    toggle.textContent = '▼';
                } else {
                    groupServers.classList.add('collapsed');
                    toggle.textContent = '▶';
                }
                
                localStorage.setItem(`group-collapsed-${groupId}`, !isCollapsed);
            });
        });
    }
    
    addServerCardListeners() {
        const grid = document.getElementById('servers-grid');
        
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
    
    updateSearchResults() {
        // 更新服务器计数
        const totalServers = this.servers ? this.servers.size : 0;
        const filteredCount = this.filteredServers.length;
        
        document.getElementById('server-count').textContent = 
            `${filteredCount}${filteredCount !== totalServers ? '/' + totalServers : ''}${this.t('serverCount')}`;
        
        // 显示搜索结果信息
        this.showSearchResultsInfo();
    }
    
    showSearchResultsInfo() {
        const existingInfo = document.querySelector('.search-results');
        if (existingInfo) {
            existingInfo.remove();
        }
        
        const totalServers = this.servers ? this.servers.size : 0;
        const filteredCount = this.filteredServers.length;
        
        if (filteredCount !== totalServers && (this.searchQuery || this.statusFilter !== 'all')) {
            const info = document.createElement('div');
            info.className = 'search-results';
            
            let message = '';
            if (this.searchQuery && this.statusFilter !== 'all') {
                message = this.t('search.foundWithFilter').replace('{count}', filteredCount)
                    .replace('{total}', totalServers)
                    .replace('{query}', this.searchQuery)
                    .replace('{filter}', this.statusFilter);
            } else if (this.searchQuery) {
                message = this.t('search.foundWithSearch').replace('{count}', filteredCount)
                    .replace('{total}', totalServers)
                    .replace('{query}', this.searchQuery);
            } else {
                message = this.t('search.foundWithStatus').replace('{count}', filteredCount)
                    .replace('{total}', totalServers)
                    .replace('{status}', this.statusFilter);
            }
            
            info.textContent = message;
            
            const grid = document.getElementById('servers-grid');
            grid.parentNode.insertBefore(info, grid);
        }
    }
    
    // 导出功能
    setupExportFunctionality() {
        const exportBtn = document.getElementById('export-btn');
        const exportMenu = document.getElementById('export-menu');
        const exportCsv = document.getElementById('export-csv');
        const exportJson = document.getElementById('export-json');
        const exportPdf = document.getElementById('export-pdf');
        
        // 切换导出菜单
        if (exportBtn) {
            exportBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                exportMenu.classList.toggle('visible');
            });
        }
        
        // 点击其他地方关闭菜单
        document.addEventListener('click', () => {
            if (exportMenu) {
                exportMenu.classList.remove('visible');
            }
        });
        
        // 导出事件
        if (exportCsv) {
            exportCsv.addEventListener('click', () => this.exportData('csv'));
        }
        if (exportJson) {
            exportJson.addEventListener('click', () => this.exportData('json'));
        }
        if (exportPdf) {
            exportPdf.addEventListener('click', () => this.exportData('pdf'));
        }
    }
    
    exportData(format) {
        try {
            const data = this.filteredServers || Array.from(this.servers.values());
            const timestamp = new Date().toISOString().split('T')[0];
            const filename = `server-status-${timestamp}`;
            
            switch (format) {
                case 'csv':
                    this.exportToCsv(data, filename);
                    break;
                case 'json':
                    this.exportToJson(data, filename);
                    break;
                case 'pdf':
                    this.exportToPdf(data, filename);
                    break;
            }
            
            this.showNotification('success', this.t('export.exportSuccess'));
            document.getElementById('export-menu').classList.remove('visible');
        } catch (error) {
            console.error('Export error:', error);
            this.showNotification('error', this.t('export.exportError'));
        }
    }
    
    exportToCsv(data, filename) {
        const headers = ['hostname', 'status', 'cpu_percent', 'memory_percent', 'disk_percent', 'max_temp', 'last_seen'];
        const csvContent = [
            headers.join(','),
            ...data.map(server => headers.map(header => {
                let value = server[header];
                if (typeof value === 'string' && value.includes(',')) {
                    value = `"${value}"`;
                }
                return value || '';
            }).join(','))
        ].join('\n');
        
        this.downloadFile(csvContent, `${filename}.csv`, 'text/csv');
    }
    
    exportToJson(data, filename) {
        const jsonContent = JSON.stringify(data, null, 2);
        this.downloadFile(jsonContent, `${filename}.json`, 'application/json');
    }
    
    exportToPdf(data, filename) {
        // 简化版PDF导出（实际中可以使用jsPDF等库）
        const htmlContent = this.generatePdfHtml(data);
        const printWindow = window.open('', '_blank');
        printWindow.document.write(htmlContent);
        printWindow.document.close();
        
        setTimeout(() => {
            printWindow.print();
            printWindow.close();
        }, 500);
    }
    
    generatePdfHtml(data) {
        const timestamp = new Date().toLocaleString();
        return `
            <!DOCTYPE html>
            <html>
            <head>
                <title>Server Status Report</title>
                <style>
                    body { font-family: Arial, sans-serif; margin: 20px; }
                    h1 { color: #333; border-bottom: 2px solid #667eea; }
                    table { width: 100%; border-collapse: collapse; margin-top: 20px; }
                    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
                    th { background-color: #f2f2f2; }
                    .online { color: #10b981; font-weight: bold; }
                    .offline { color: #ef4444; font-weight: bold; }
                </style>
            </head>
            <body>
                <h1>Server Status Report</h1>
                <p>Generated on: ${timestamp}</p>
                <p>Total Servers: ${data.length}</p>
                <table>
                    <thead>
                        <tr>
                            <th>Hostname</th>
                            <th>Status</th>
                            <th>CPU %</th>
                            <th>Memory %</th>
                            <th>Disk %</th>
                            <th>Temperature</th>
                            <th>Last Seen</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${data.map(server => `
                            <tr>
                                <td>${server.hostname}</td>
                                <td class="${server.status}">${server.status}</td>
                                <td>${server.cpu_percent.toFixed(1)}%</td>
                                <td>${server.memory_percent.toFixed(1)}%</td>
                                <td>${server.disk_percent.toFixed(1)}%</td>
                                <td>${server.max_temp > 0 ? server.max_temp.toFixed(1) + '°C' : 'N/A'}</td>
                                <td>${new Date(server.last_seen).toLocaleString()}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </body>
            </html>
        `;
    }
    
    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
    }
    
    // 告警系统
    setupAlerts() {
        const alertsBtn = document.getElementById('alerts-toggle');
        const alertsPanel = document.getElementById('alerts-panel');
        const closeAlerts = document.getElementById('close-alerts');
        const clearAlerts = document.getElementById('clear-alerts');
        
        if (alertsBtn) {
            alertsBtn.addEventListener('click', () => {
                const isVisible = alertsPanel.style.display !== 'none';
                alertsPanel.style.display = isVisible ? 'none' : 'block';
                
                if (!isVisible) {
                    this.updateAlertsDisplay();
                }
            });
        }
        
        if (closeAlerts) {
            closeAlerts.addEventListener('click', () => {
                alertsPanel.style.display = 'none';
            });
        }
        
        if (clearAlerts) {
            clearAlerts.addEventListener('click', () => {
                this.clearAllAlerts();
            });
        }
    }
    
    requestNotificationPermission() {
        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    }
    
    checkServerAlerts(server) {
        const alerts = [];
        const now = Date.now();
        
        // CPU告警
        if (server.cpu_percent > this.alertThresholds.cpu) {
            alerts.push({
                id: `${server.hostname}-cpu-${now}`,
                type: 'critical',
                title: this.t('alerts.highCpuUsage'),
                description: `${server.hostname}: CPU ${server.cpu_percent.toFixed(1)}%`,
                server: server.hostname,
                timestamp: now
            });
        }
        
        // 内存告警
        if (server.memory_percent > this.alertThresholds.memory) {
            alerts.push({
                id: `${server.hostname}-memory-${now}`,
                type: 'critical',
                title: this.t('alerts.highMemoryUsage'),
                description: `${server.hostname}: Memory ${server.memory_percent.toFixed(1)}%`,
                server: server.hostname,
                timestamp: now
            });
        }
        
        // 磁盘告警
        if (server.disk_percent > this.alertThresholds.disk) {
            alerts.push({
                id: `${server.hostname}-disk-${now}`,
                type: 'critical',
                title: this.t('alerts.highDiskUsage'),
                description: `${server.hostname}: Disk ${server.disk_percent.toFixed(1)}%`,
                server: server.hostname,
                timestamp: now
            });
        }
        
        // 温度告警
        if (server.max_temp > this.alertThresholds.temperature) {
            alerts.push({
                id: `${server.hostname}-temp-${now}`,
                type: 'warning',
                title: this.t('alerts.highTemperature'),
                description: `${server.hostname}: Temperature ${server.max_temp.toFixed(1)}°C`,
                server: server.hostname,
                timestamp: now
            });
        }
        
        // 离线告警
        if (server.status === 'offline') {
            alerts.push({
                id: `${server.hostname}-offline-${now}`,
                type: 'critical',
                title: this.t('alerts.serverOffline'),
                description: `${server.hostname} is currently offline`,
                server: server.hostname,
                timestamp: now
            });
        }
        
        return alerts;
    }
    
    addAlert(alert) {
        // 检查是否已存在相同类型的告警
        const existingAlert = this.alerts.find(a => 
            a.server === alert.server && 
            a.type === alert.type && 
            a.title === alert.title
        );
        
        if (!existingAlert) {
            this.alerts.unshift(alert);
            
            // 限制告警数量
            if (this.alerts.length > 50) {
                this.alerts = this.alerts.slice(0, 50);
            }
            
            // 显示通知
            this.showNotification(alert.type, alert.title, alert.description);
            
            // 浏览器通知
            if (Notification.permission === 'granted') {
                new Notification(alert.title, {
                    body: alert.description,
                    icon: '/favicon.ico'
                });
            }
            
            this.updateAlertCount();
        }
    }
    
    updateAlertCount() {
        const alertCount = document.getElementById('alert-count');
        const alertsBtn = document.getElementById('alerts-toggle');
        
        if (alertCount && alertsBtn) {
            const criticalAlerts = this.alerts.filter(a => a.type === 'critical').length;
            
            if (criticalAlerts > 0) {
                alertCount.textContent = criticalAlerts;
                alertCount.classList.add('visible');
                alertsBtn.classList.add('has-alerts');
            } else {
                alertCount.classList.remove('visible');
                alertsBtn.classList.remove('has-alerts');
            }
        }
    }
    
    updateAlertsDisplay() {
        const alertsContent = document.getElementById('alerts-content');
        const noAlerts = document.getElementById('no-alerts');
        
        if (this.alerts.length === 0) {
            if (noAlerts) {
                noAlerts.style.display = 'block';
            }
            alertsContent.innerHTML = `<div class="no-alerts">${this.t('alerts.noAlerts')}</div>`;
        } else {
            if (noAlerts) {
                noAlerts.style.display = 'none';
            }
            
            alertsContent.innerHTML = this.alerts.map(alert => `
                <div class="alert-item ${alert.type}">
                    <div class="alert-title">${alert.title}</div>
                    <div class="alert-description">${alert.description}</div>
                    <div class="alert-time">${new Date(alert.timestamp).toLocaleString()}</div>
                </div>
            `).join('');
        }
    }
    
    clearAllAlerts() {
        this.alerts = [];
        this.updateAlertCount();
        this.updateAlertsDisplay();
    }
    
    // 统计面板
    setupStatsPanel() {
        const toggleBtn = document.getElementById('toggle-stats');
        const statsPanel = document.getElementById('stats-panel');
        
        if (toggleBtn) {
            toggleBtn.addEventListener('click', () => {
                this.statsVisible = !this.statsVisible;
                statsPanel.classList.toggle('collapsed', !this.statsVisible);
                toggleBtn.textContent = this.statsVisible ? this.t('stats.hide') : this.t('stats.show');
                
                // 保存状态
                localStorage.setItem('stats-visible', this.statsVisible);
            });
        }
        
        // 恢复状态
        const savedStatsVisible = localStorage.getItem('stats-visible');
        if (savedStatsVisible !== null) {
            this.statsVisible = savedStatsVisible === 'true';
            if (statsPanel) {
                statsPanel.classList.toggle('collapsed', !this.statsVisible);
            }
        }
    }
    
    updatePerformanceStats(servers) {
        if (!servers || servers.length === 0) return;
        
        const onlineServers = servers.filter(s => s.status === 'online');
        const totalServers = servers.length;
        
        // 计算平均值
        const avgCpu = onlineServers.reduce((sum, s) => sum + s.cpu_percent, 0) / onlineServers.length || 0;
        const avgMemory = onlineServers.reduce((sum, s) => sum + s.memory_percent, 0) / onlineServers.length || 0;
        const avgDisk = onlineServers.reduce((sum, s) => sum + s.disk_percent, 0) / onlineServers.length || 0;
        
        // 计算网络平均速率
        const avgNetworkSent = onlineServers.reduce((sum, s) => sum + (s.network_speed_sent || 0), 0) / onlineServers.length || 0;
        const avgNetworkRecv = onlineServers.reduce((sum, s) => sum + (s.network_speed_recv || 0), 0) / onlineServers.length || 0;
        
        // 高负载服务器数量
        const highUsageServers = onlineServers.filter(s => 
            s.cpu_percent > 80 || s.memory_percent > 80 || s.disk_percent > 80
        ).length;
        
        // 总负载（加权平均）
        const totalLoad = (avgCpu * 0.4 + avgMemory * 0.4 + avgDisk * 0.2);
        
        // 严重告警数量
        const criticalAlerts = this.alerts.filter(a => a.type === 'critical').length;
        
        // 更新界面
        this.updateElementText('avg-cpu', `${avgCpu.toFixed(1)}%`);
        this.updateElementText('avg-memory', `${avgMemory.toFixed(1)}%`);
        this.updateElementText('avg-disk', `${avgDisk.toFixed(1)}%`);
        this.updateElementText('high-usage', highUsageServers.toString());
        this.updateElementText('total-load', `${totalLoad.toFixed(1)}%`);
        this.updateElementText('critical-count', criticalAlerts.toString());
        this.updateElementText('avg-network-sent', this.formatNetworkSpeed(avgNetworkSent));
        this.updateElementText('avg-network-recv', this.formatNetworkSpeed(avgNetworkRecv));
        
        // 更新样式
        const avgCpuElement = document.getElementById('avg-cpu');
        const avgMemoryElement = document.getElementById('avg-memory');
        const avgDiskElement = document.getElementById('avg-disk');
        const totalLoadElement = document.getElementById('total-load');
        
        if (avgCpuElement) {
            avgCpuElement.className = `stat-value ${avgCpu > 80 ? 'critical' : ''}`;
        }
        if (avgMemoryElement) {
            avgMemoryElement.className = `stat-value ${avgMemory > 80 ? 'critical' : ''}`;
        }
        if (avgDiskElement) {
            avgDiskElement.className = `stat-value ${avgDisk > 80 ? 'critical' : ''}`;
        }
        if (totalLoadElement) {
            totalLoadElement.className = `stat-value ${totalLoad > 75 ? 'critical' : ''}`;
        }
    }
    
    // 通知系统
    showNotification(type, title, message = '') {
        const container = document.getElementById('notifications-container');
        if (!container) return;
        
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="notification-header">
                <div class="notification-title">${title}</div>
                <button class="notification-close">×</button>
            </div>
            ${message ? `<div class="notification-message">${message}</div>` : ''}
        `;
        
        container.appendChild(notification);
        
        // 关闭按钮
        const closeBtn = notification.querySelector('.notification-close');
        closeBtn.addEventListener('click', () => {
            notification.remove();
        });
        
        // 自动关闭
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 5000);
        
        // 限制通知数量
        const notifications = container.querySelectorAll('.notification');
        if (notifications.length > 5) {
            notifications[0].remove();
        }
    }
    
    // 键盘快捷键
    handleKeyboardShortcuts(e) {
        // 如果用户正在输入，则不处理快捷键
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT' || e.target.tagName === 'TEXTAREA') {
            if (e.key === 'Escape') {
                e.target.blur();
            }
            return;
        }
        
        // Ctrl/Cmd + 组合键
        if (e.ctrlKey || e.metaKey) {
            switch (e.key) {
                case 'e': // 导出
                    e.preventDefault();
                    document.getElementById('export-btn').click();
                    break;
                case 'f': // 搜索
                    e.preventDefault();
                    document.getElementById('server-search').focus();
                    break;
                case 'd': // 切换主题
                    e.preventDefault();
                    this.toggleTheme();
                    break;
            }
            return;
        }
        
        // 单个按键
        switch (e.key) {
            case 'Escape':
                // 关闭所有弹窗
                this.hideCharts();
                document.getElementById('alerts-panel').style.display = 'none';
                document.getElementById('export-menu').classList.remove('visible');
                break;
                
            case '/':
                // 聚焦搜索框
                e.preventDefault();
                document.getElementById('server-search').focus();
                break;
                
            case 't':
                // 切换主题
                if (!e.ctrlKey && !e.metaKey) {
                    this.toggleTheme();
                }
                break;
                
            case 'l':
                // 切换语言
                if (!e.ctrlKey && !e.metaKey) {
                    const currentLang = this.language === 'zh' ? 'en' : 'zh';
                    this.switchLanguage(currentLang);
                }
                break;
                
            case 'a':
                // 切换告警面板
                if (!e.ctrlKey && !e.metaKey) {
                    document.getElementById('alerts-toggle').click();
                }
                break;
                
            case 's':
                // 切换统计面板
                if (!e.ctrlKey && !e.metaKey) {
                    document.getElementById('toggle-stats').click();
                }
                break;
                
            case 'g':
                // 打开分组管理
                if (!e.ctrlKey && !e.metaKey) {
                    this.showGroupModal();
                }
                break;
                
            case 'v':
                // 切换视图模式
                if (!e.ctrlKey && !e.metaKey) {
                    this.toggleViewMode();
                }
                break;
                
            case '1':
            case '2':
            case '3':
            case '4':
                // 切换网格布局
                if (!e.ctrlKey && !e.metaKey) {
                    const gridTypes = ['auto', 'grid-3x3', 'grid-4x4', 'grid-5x5'];
                    const index = parseInt(e.key) - 1;
                    if (index >= 0 && index < gridTypes.length) {
                        const button = document.querySelector(`[data-grid="${gridTypes[index]}"]`);
                        if (button) {
                            button.click();
                        }
                    }
                }
                break;
        }
    }
    
    // 初始化时显示快捷键提示
    showKeyboardShortcutsHelp() {
        const shortcuts = {
            zh: {
                title: '键盘快捷键',
                shortcuts: [
                    'Ctrl/Cmd + E: 导出数据',
                    'Ctrl/Cmd + F: 搜索服务器',
                    'Ctrl/Cmd + D: 切换主题',
                    '/: 聚焦搜索框',
                    'T: 切换主题',
                    'L: 切换语言',
                    'A: 切换告警面板',
                    'S: 切换统计面板',
                    '1-4: 切换网格布局',
                    'ESC: 关闭弹窗'
                ]
            },
            en: {
                title: 'Keyboard Shortcuts',
                shortcuts: [
                    'Ctrl/Cmd + E: Export data',
                    'Ctrl/Cmd + F: Search servers',
                    'Ctrl/Cmd + D: Toggle theme',
                    '/: Focus search box',
                    'T: Toggle theme',
                    'L: Switch language',
                    'A: Toggle alerts panel',
                    'S: Toggle stats panel',
                    '1-4: Switch grid layout',
                    'ESC: Close dialogs'
                ]
            }
        };
        
        const lang = shortcuts[this.language];
        const message = `${lang.title}:\n\n${lang.shortcuts.join('\n')}`;
        
        this.showNotification('info', lang.title, message.replace(/\n/g, '<br>'));
    }
    
    setupHelpButton() {
        const helpBtn = document.getElementById('help-toggle');
        if (helpBtn) {
            helpBtn.addEventListener('click', () => {
                this.showKeyboardShortcutsHelp();
            });
        }
    }
    
    // 分组管理功能
    setupGroupManagement() {
        // 管理分组按钮
        const manageBtn = document.getElementById('manage-groups-btn');
        if (manageBtn) {
            manageBtn.addEventListener('click', () => {
                this.showGroupModal();
            });
        }
        
        // 视图切换按钮
        const viewToggle = document.getElementById('view-toggle');
        if (viewToggle) {
            viewToggle.addEventListener('click', () => {
                this.toggleViewMode();
            });
        }
        
        // 分组筛选
        const groupFilter = document.getElementById('group-filter');
        if (groupFilter) {
            groupFilter.addEventListener('change', (e) => {
                this.groupFilter = e.target.value;
                this.applyFilters();
            });
        }
        
        // 模态框事件
        this.setupGroupModal();
        
        // 右键菜单
        this.setupContextMenu();
    }
    
    setupGroupModal() {
        const modal = document.getElementById('group-modal');
        const closeBtn = document.getElementById('close-group-modal');
        const form = document.getElementById('group-form');
        
        // 关闭模态框
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                modal.style.display = 'none';
            });
        }
        
        // 点击背景关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });
        
        // 创建分组表单
        if (form) {
            form.addEventListener('submit', (e) => {
                e.preventDefault();
                this.createGroup();
            });
        }
        
        // 批量分配
        const bulkAssignBtn = document.getElementById('bulk-assign-btn');
        if (bulkAssignBtn) {
            bulkAssignBtn.addEventListener('click', () => {
                this.bulkAssignServers();
            });
        }
        
        // 导出分组
        const exportGroupsBtn = document.getElementById('export-groups-btn');
        if (exportGroupsBtn) {
            exportGroupsBtn.addEventListener('click', () => {
                this.exportGroups();
            });
        }
        
        // 导入分组
        const importGroupsBtn = document.getElementById('import-groups-btn');
        const importGroupsFile = document.getElementById('import-groups-file');
        
        if (importGroupsBtn && importGroupsFile) {
            importGroupsBtn.addEventListener('click', () => {
                importGroupsFile.click();
            });
            
            importGroupsFile.addEventListener('change', (e) => {
                const file = e.target.files[0];
                if (file) {
                    this.importGroups(file);
                }
            });
        }
    }
    
    setupContextMenu() {
        const contextMenu = document.getElementById('server-context-menu');
        
        // 隐藏右键菜单
        document.addEventListener('click', () => {
            contextMenu.style.display = 'none';
        });
        
        // 右键菜单事件
        document.addEventListener('contextmenu', (e) => {
            const serverCard = e.target.closest('.server-card');
            if (serverCard) {
                e.preventDefault();
                this.showContextMenu(e, serverCard);
            }
        });
    }
    
    showGroupModal() {
        const modal = document.getElementById('group-modal');
        modal.style.display = 'flex';
        this.updateGroupModal();
    }
    
    updateGroupModal() {
        this.updateGroupsList();
        this.updateServerAssignmentList();
        this.updateAssignmentGroupSelect();
    }
    
    createGroup() {
        const name = document.getElementById('group-name').value.trim();
        const type = document.getElementById('group-type').value;
        const description = document.getElementById('group-description').value.trim();
        const color = document.getElementById('group-color').value;
        
        if (!name) {
            this.showNotification('error', 'Error', 'Group name is required');
            return;
        }
        
        const groupId = `group-${Date.now()}`;
        const group = {
            id: groupId,
            name,
            type,
            description,
            color,
            created: new Date().toISOString(),
            servers: []
        };
        
        this.groups.set(groupId, group);
        this.saveGroups();
        
        // 清空表单
        document.getElementById('group-form').reset();
        document.getElementById('group-color').value = '#667eea';
        
        this.showNotification('success', this.t('groups.groupCreated'));
        this.updateGroupModal();
        this.updateGroupFilter();
    }
    
    deleteGroup(groupId) {
        if (confirm(this.t('groups.confirmDelete'))) {
            // 移除服务器分组关系
            for (const [hostname, assignedGroupId] of this.serverGroups.entries()) {
                if (assignedGroupId === groupId) {
                    this.serverGroups.delete(hostname);
                }
            }
            
            this.groups.delete(groupId);
            this.saveGroups();
            this.saveServerGroups();
            
            this.showNotification('success', this.t('groups.groupDeleted'));
            this.updateGroupModal();
            this.updateGroupFilter();
            this.applyFilters();
        }
    }
    
    updateGroupsList() {
        const container = document.getElementById('groups-container');
        
        if (this.groups.size === 0) {
            container.innerHTML = `<div class="no-groups">${this.t('groups.noGroups')}</div>`;
            return;
        }
        
        container.innerHTML = Array.from(this.groups.values()).map(group => {
            const serverCount = Array.from(this.serverGroups.values()).filter(gId => gId === group.id).length;
            const stats = this.calculateGroupStats(group.id);
            
            return `
                <div class="group-item">
                    <div class="group-info">
                        <div class="group-color" style="background-color: ${group.color}"></div>
                        <div class="group-details">
                            <h5>${group.name}</h5>
                            <p class="group-meta">${this.t(`groups.${group.type}`)} • ${serverCount} ${this.t('groups.servers')}</p>
                            ${group.description ? `<p class="group-description">${group.description}</p>` : ''}
                        </div>
                    </div>
                    <div class="group-stats">
                        <div class="group-stats-grid">
                            <div class="group-stat">
                                <span class="group-stat-label">Online:</span>
                                <span class="group-stat-value online">${stats.onlineCount}/${stats.totalCount}</span>
                            </div>
                            <div class="group-stat">
                                <span class="group-stat-label">Avg CPU:</span>
                                <span class="group-stat-value">${stats.avgCpu.toFixed(1)}%</span>
                            </div>
                            <div class="group-stat">
                                <span class="group-stat-label">Avg Memory:</span>
                                <span class="group-stat-value">${stats.avgMemory.toFixed(1)}%</span>
                            </div>
                            <div class="group-stat">
                                <span class="group-stat-label">Critical:</span>
                                <span class="group-stat-value ${stats.criticalCount > 0 ? 'critical' : ''}">${stats.criticalCount}</span>
                            </div>
                        </div>
                    </div>
                    <div class="group-actions">
                        <button class="group-btn edit" onclick="window.gpuMonitor.editGroup('${group.id}')">${this.t('groups.edit')}</button>
                        <button class="group-btn delete" onclick="window.gpuMonitor.deleteGroup('${group.id}')">${this.t('groups.delete')}</button>
                    </div>
                </div>
            `;
        }).join('');
    }
    
    updateServerAssignmentList() {
        const container = document.getElementById('servers-assignment-list');
        const servers = Array.from(this.servers.values());
        
        container.innerHTML = servers.map(server => {
            const groupId = this.serverGroups.get(server.hostname);
            const group = groupId ? this.groups.get(groupId) : null;
            
            return `
                <div class="assignment-server">
                    <input type="checkbox" id="server-${server.hostname}" value="${server.hostname}">
                    <div class="assignment-server-info">
                        <span class="assignment-server-name">${server.hostname}</span>
                        <span class="assignment-server-group">
                            ${group ? group.name : this.t('groups.ungrouped')}
                        </span>
                    </div>
                </div>
            `;
        }).join('');
    }
    
    updateAssignmentGroupSelect() {
        const select = document.getElementById('assignment-group');
        
        select.innerHTML = `
            <option value="">${this.t('groups.selectGroup')}</option>
            ${Array.from(this.groups.values()).map(group => 
                `<option value="${group.id}">${group.name}</option>`
            ).join('')}
        `;
    }
    
    bulkAssignServers() {
        const groupId = document.getElementById('assignment-group').value;
        if (!groupId) {
            this.showNotification('error', 'Error', 'Please select a group');
            return;
        }
        
        const checkedServers = document.querySelectorAll('#servers-assignment-list input[type="checkbox"]:checked');
        let assignedCount = 0;
        
        checkedServers.forEach(checkbox => {
            const hostname = checkbox.value;
            this.serverGroups.set(hostname, groupId);
            checkbox.checked = false;
            assignedCount++;
        });
        
        if (assignedCount > 0) {
            this.saveServerGroups();
            this.showNotification('success', `${assignedCount} ${this.t('groups.serversAssigned')}`);
            this.updateServerAssignmentList();
            this.applyFilters();
        }
    }
    
    showContextMenu(event, serverCard) {
        const contextMenu = document.getElementById('server-context-menu');
        const hostname = serverCard.dataset.hostname;
        this.contextMenuTarget = hostname;
        
        // 更新分组子菜单
        this.updateGroupSubmenu();
        
        // 定位菜单
        contextMenu.style.left = event.pageX + 'px';
        contextMenu.style.top = event.pageY + 'px';
        contextMenu.style.display = 'block';
        
        // 菜单事件
        const assignToGroup = document.getElementById('assign-to-group');
        const removeFromGroup = document.getElementById('remove-from-group');
        
        removeFromGroup.onclick = () => {
            this.removeServerFromGroup(hostname);
            contextMenu.style.display = 'none';
        };
    }
    
    updateGroupSubmenu() {
        const submenu = document.getElementById('group-submenu');
        
        submenu.innerHTML = Array.from(this.groups.values()).map(group => `
            <div class="submenu-item" onclick="window.gpuMonitor.assignServerToGroup('${this.contextMenuTarget}', '${group.id}')">
                <div class="submenu-color" style="background-color: ${group.color}"></div>
                <span>${group.name}</span>
            </div>
        `).join('');
    }
    
    assignServerToGroup(hostname, groupId) {
        this.serverGroups.set(hostname, groupId);
        this.saveServerGroups();
        
        const group = this.groups.get(groupId);
        this.showNotification('success', `${hostname} assigned to ${group.name}`);
        
        document.getElementById('server-context-menu').style.display = 'none';
        this.applyFilters();
    }
    
    removeServerFromGroup(hostname) {
        this.serverGroups.delete(hostname);
        this.saveServerGroups();
        
        this.showNotification('success', `${hostname} removed from group`);
        this.applyFilters();
    }
    
    // 数据持久化
    saveGroups() {
        const groupsData = Array.from(this.groups.entries());
        localStorage.setItem('server-groups-config', JSON.stringify(groupsData));
    }
    
    loadGroups() {
        try {
            const saved = localStorage.getItem('server-groups-config');
            if (saved) {
                const groupsData = JSON.parse(saved);
                this.groups = new Map(groupsData);
            }
            
            const savedServerGroups = localStorage.getItem('server-groups-assignments');
            if (savedServerGroups) {
                const serverGroupsData = JSON.parse(savedServerGroups);
                this.serverGroups = new Map(serverGroupsData);
            }
            
            this.updateGroupFilter();
        } catch (error) {
            console.error('Error loading groups:', error);
        }
    }
    
    saveServerGroups() {
        const serverGroupsData = Array.from(this.serverGroups.entries());
        localStorage.setItem('server-groups-assignments', JSON.stringify(serverGroupsData));
    }
    
    exportGroups() {
        try {
            const exportData = {
                groups: Array.from(this.groups.entries()),
                serverGroups: Array.from(this.serverGroups.entries()),
                exportDate: new Date().toISOString(),
                version: '1.0'
            };
            
            const jsonString = JSON.stringify(exportData, null, 2);
            const timestamp = new Date().toISOString().split('T')[0];
            const filename = `server-groups-${timestamp}.json`;
            
            this.downloadFile(jsonString, filename, 'application/json');
            this.showNotification('success', 'Groups exported successfully');
        } catch (error) {
            console.error('Export error:', error);
            this.showNotification('error', 'Failed to export groups');
        }
    }
    
    importGroups(file) {
        const reader = new FileReader();
        reader.onload = (e) => {
            try {
                const importData = JSON.parse(e.target.result);
                
                // Validate the import data structure
                if (!importData.groups || !importData.serverGroups || !Array.isArray(importData.groups) || !Array.isArray(importData.serverGroups)) {
                    throw new Error('Invalid file format');
                }
                
                // Show confirmation dialog
                const importConfirm = confirm(`Import ${importData.groups.length} groups and ${importData.serverGroups.length} server assignments?\n\nThis will replace existing groups. Continue?`);
                
                if (importConfirm) {
                    // Clear existing data
                    this.groups.clear();
                    this.serverGroups.clear();
                    
                    // Import groups
                    importData.groups.forEach(([id, group]) => {
                        // Validate group structure
                        if (group.id && group.name && group.color) {
                            this.groups.set(id, group);
                        }
                    });
                    
                    // Import server assignments
                    importData.serverGroups.forEach(([hostname, groupId]) => {
                        // Only assign if the group exists
                        if (this.groups.has(groupId)) {
                            this.serverGroups.set(hostname, groupId);
                        }
                    });
                    
                    // Save to localStorage
                    this.saveGroups();
                    this.saveServerGroups();
                    
                    // Update UI
                    this.updateGroupModal();
                    this.updateGroupFilter();
                    this.applyFilters();
                    
                    this.showNotification('success', `Successfully imported ${this.groups.size} groups`);
                }
            } catch (error) {
                console.error('Import error:', error);
                this.showNotification('error', 'Failed to import groups: Invalid file format');
            }
            
            // Reset file input
            document.getElementById('import-groups-file').value = '';
        };
        
        reader.readAsText(file);
    }
    
    calculateGroupStats(groupId) {
        const groupServers = Array.from(this.servers.values()).filter(server => 
            this.serverGroups.get(server.hostname) === groupId
        );
        
        return this.calculateGroupStatsFromServers(groupServers);
    }
    
    calculateGroupStatsFromServers(servers) {
        if (servers.length === 0) {
            return {
                totalCount: 0,
                onlineCount: 0,
                avgCpu: 0,
                avgMemory: 0,
                avgDisk: 0,
                avgTemp: 0,
                criticalCount: 0
            };
        }
        
        const onlineServers = servers.filter(server => server.status === 'online');
        const cpuSum = servers.reduce((sum, server) => sum + (server.cpu_percent || 0), 0);
        const memorySum = servers.reduce((sum, server) => sum + (server.memory_percent || 0), 0);
        const diskSum = servers.reduce((sum, server) => sum + (server.disk_percent || 0), 0);
        const tempSum = servers.reduce((sum, server) => sum + (server.max_temp || 0), 0);
        
        // Count critical servers (high CPU, memory, or disk usage)
        const criticalCount = servers.filter(server => 
            server.cpu_percent > this.alertThresholds.cpu ||
            server.memory_percent > this.alertThresholds.memory ||
            server.disk_percent > this.alertThresholds.disk ||
            server.status === 'offline'
        ).length;
        
        return {
            totalCount: servers.length,
            onlineCount: onlineServers.length,
            avgCpu: servers.length > 0 ? cpuSum / servers.length : 0,
            avgMemory: servers.length > 0 ? memorySum / servers.length : 0,
            avgDisk: servers.length > 0 ? diskSum / servers.length : 0,
            avgTemp: servers.length > 0 ? tempSum / servers.length : 0,
            criticalCount
        };
    }
    
    updateGroupFilter() {
        const select = document.getElementById('group-filter');
        const currentValue = select.value;
        
        // 更新选项
        const defaultOptions = `
            <option value="all">${this.t('groups.allGroups')}</option>
            <option value="ungrouped">${this.t('groups.ungrouped')}</option>
        `;
        
        const groupOptions = Array.from(this.groups.values()).map(group => 
            `<option value="${group.id}">${group.name}</option>`
        ).join('');
        
        select.innerHTML = defaultOptions + groupOptions;
        select.value = currentValue;
    }
    
    toggleViewMode() {
        this.viewMode = this.viewMode === 'cards' ? 'grouped' : 'cards';
        const viewBtn = document.getElementById('view-toggle');
        const viewIcon = document.getElementById('view-icon');
        
        if (this.viewMode === 'grouped') {
            viewBtn.classList.add('grouped');
            viewIcon.innerHTML = `<path d="M3,3H11V5H3V3M13,3H21V5H13V3M3,7H11V9H3V7M13,7H21V9H13V7M3,11H11V13H3V11M13,11H21V13H13V11M3,15H11V17H3V15M13,15H21V17H13V15M3,19H11V21H3V19M13,19H21V21H13V19Z"/>`;
        } else {
            viewBtn.classList.remove('grouped');
            viewIcon.innerHTML = `<path d="M3,5H9V11H3V5M5,7V9H7V7H5M11,7H21V9H11V7M11,15H21V17H11V15M5,20V18H7V20H5M3,17H9V23H3V17M11,11H21V13H11V11Z"/>`;
        }
        
        localStorage.setItem('view-mode', this.viewMode);
        this.renderFilteredServers();
    }
    
    updateElementText(id, text) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = text;
        }
    }
    
    updateTipContents() {
        const tips = document.querySelectorAll('.tip');
        const tipKeys = ['publicPanel', 'backgroundRun', 'networkMonitor', 'stopMonitor', 'checkStatus'];
        const tipDescKeys = ['publicPanelDesc', 'backgroundRunDesc', 'networkMonitorDesc', 'stopMonitorDesc', 'checkStatusDesc'];
        
        tips.forEach((tip, index) => {
            if (index < tipKeys.length) {
                const strong = tip.querySelector('strong');
                const span = tip.querySelector('span');
                if (strong) {
                    strong.textContent = this.t(`tutorial.${tipKeys[index]}`);
                }
                if (span) {
                    span.innerHTML = this.t(`tutorial.${tipDescKeys[index]}`);
                }
            }
        });
    }
    
    updateTutorialLanguage() {
        // 这个方法现在由 updateAllTexts 处理
        // 保留为了兼容性，但主要逻辑已移动到 updateAllTexts
        
        // 更新切换按钮文本
        const toggleBtn = document.getElementById('toggle-tutorial-btn');
        if (toggleBtn) {
            const tutorialSection = document.getElementById('tutorial-section');
            const isCollapsed = tutorialSection && tutorialSection.classList.contains('collapsed');
            toggleBtn.textContent = isCollapsed ? this.t('tutorial.showTutorial') : this.t('tutorial.hideTutorial');
        }
    }

    parseURLParameters() {
        // 配置文件中的访问密钥优先级最高
        if (CONFIG.ACCESS_KEY_MODE && CONFIG.ACCESS_KEY) {
            this.accessKey = CONFIG.ACCESS_KEY;
            this.apiEndpoint = `${this.apiBaseUrl}/api/access/${CONFIG.ACCESS_KEY}/servers`;
            console.log('使用配置的访问密钥模式:', CONFIG.ACCESS_KEY);
            
            // 更新页面标题显示当前访问模式
            this.updatePageTitle('访问密钥模式');
            return;
        }
        
        const urlParams = new URLSearchParams(window.location.search);
        
        // 检查URL中的access key参数
        const accessKey = urlParams.get('access') || urlParams.get('accessKey') || urlParams.get('key');
        if (accessKey) {
            this.accessKey = accessKey;
            this.apiEndpoint = `${this.apiBaseUrl}/api/access/${accessKey}/servers`;
            console.log('使用URL访问密钥模式:', accessKey);
            
            // 更新页面标题显示当前访问模式
            this.updatePageTitle('访问密钥模式');
            return;
        }
        
        // 已移除项目密钥和访问令牌支持，只保留AccessKey访问方式
        
        if (CONFIG.DEBUG) {
            console.log('使用默认模式 (无认证)');
            console.log('API Base URL:', this.apiBaseUrl);
            console.log('API Endpoint:', this.apiEndpoint);
        }
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
            this.handleKeyboardShortcuts(e);
        });
        
        // 访问密钥模态框事件
        this.setupAccessKeyModal();
        
        // 网格布局切换事件
        this.setupGridControls();
    }

    setupAccessKeyModal() {
        // 访问密钥按钮点击事件
        const accessKeyBtn = document.getElementById('access-key-toggle');
        if (accessKeyBtn) {
            accessKeyBtn.addEventListener('click', () => {
                this.showAccessKeyModal();
            });
        }

        // 关闭模态框事件
        const closeBtn = document.getElementById('close-access-key-modal');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                this.hideAccessKeyModal();
            });
        }

        // 模态框背景点击关闭
        const modal = document.getElementById('access-key-modal');
        if (modal) {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.hideAccessKeyModal();
                }
            });
        }

        // 测试连接按钮
        const testBtn = document.getElementById('test-connection-btn');
        if (testBtn) {
            testBtn.addEventListener('click', () => {
                this.testApiConnection();
            });
        }

        // 应用配置表单
        const form = document.getElementById('access-key-form');
        if (form) {
            form.addEventListener('submit', (e) => {
                e.preventDefault();
                this.applyAccessKeyConfig();
            });
        }

        // 重置配置按钮
        const resetBtn = document.getElementById('reset-config-btn');
        if (resetBtn) {
            resetBtn.addEventListener('click', () => {
                this.resetToDefaultConfig();
            });
        }

        // ESC键关闭模态框
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.hideAccessKeyModal();
            }
        });
    }

    showAccessKeyModal() {
        const modal = document.getElementById('access-key-modal');
        if (modal) {
            // 更新当前配置显示
            this.updateCurrentConfigDisplay();
            
            // 从localStorage读取保存的配置
            const savedApiUrl = localStorage.getItem('serverStatus_api_url');
            const savedAccessKey = localStorage.getItem('serverStatus_access_key');
            
            // 填充表单
            document.getElementById('api-url-input').value = savedApiUrl || this.apiBaseUrl;
            document.getElementById('access-key-input').value = savedAccessKey || this.accessKey || '';
            
            // 如果有保存的配置，勾选保存选项
            const saveToLocalCheckbox = document.getElementById('save-to-local-checkbox');
            if (saveToLocalCheckbox) {
                saveToLocalCheckbox.checked = !!savedApiUrl;
            }
            
            // 重置连接状态
            this.updateConnectionStatus(this.t('accessKey.notTested'), '');
            
            // 显示模态框
            modal.style.display = 'block';
            
            // 聚焦到第一个输入框
            setTimeout(() => {
                document.getElementById('api-url-input').focus();
            }, 100);
        }
    }

    hideAccessKeyModal() {
        const modal = document.getElementById('access-key-modal');
        if (modal) {
            modal.style.display = 'none';
        }
    }

    updateCurrentConfigDisplay() {
        // 更新API服务器地址显示
        const apiUrlElement = document.getElementById('current-api-url');
        if (apiUrlElement) {
            apiUrlElement.textContent = this.apiBaseUrl;
        }

        // 更新访问模式显示
        const accessModeElement = document.getElementById('current-access-mode');
        const currentKeyItem = document.getElementById('current-key-item');
        const currentKeyElement = document.getElementById('current-access-key');

        if (this.accessKey) {
            if (accessModeElement) {
                accessModeElement.textContent = 'Access Key Mode';
            }
            if (currentKeyItem) {
                currentKeyItem.style.display = 'block';
            }
            if (currentKeyElement) {
                // 只显示访问密钥的前8个字符
                const maskedKey = this.accessKey.length > 8 
                    ? this.accessKey.substring(0, 8) + '...' 
                    : this.accessKey;
                currentKeyElement.textContent = maskedKey;
            }
        } else {
            if (accessModeElement) {
                accessModeElement.textContent = 'Public Mode';
            }
            if (currentKeyItem) {
                currentKeyItem.style.display = 'none';
            }
        }
    }

    async testApiConnection() {
        const apiUrl = document.getElementById('api-url-input').value.trim();
        const accessKey = document.getElementById('access-key-input').value.trim();

        if (!apiUrl) {
            this.updateConnectionStatus('Error: API URL is required', 'error');
            return;
        }

        this.updateConnectionStatus(this.t('accessKey.testing'), 'warning');

        try {
            let testUrl = `${apiUrl}/api/servers`;
            
            // 如果有访问密钥，使用访问密钥API
            if (accessKey) {
                testUrl = `${apiUrl}/api/access/${accessKey}/servers`;
            }

            const response = await fetch(testUrl);
            
            if (response.ok) {
                const servers = await response.json();
                const successMessage = this.t('accessKey.connectionSuccess').replace('{count}', servers.length);
                this.updateConnectionStatus(successMessage, 'success');
            } else {
                this.updateConnectionStatus(
                    `${this.t('accessKey.connectionFailed')}: HTTP ${response.status} ${response.statusText}`, 
                    'error'
                );
            }
        } catch (error) {
            this.updateConnectionStatus(
                `${this.t('accessKey.connectionFailed')}: ${error.message}`, 
                'error'
            );
        }
    }

    updateConnectionStatus(message, type) {
        const statusElement = document.getElementById('connection-status-text');
        const statusContainer = document.getElementById('connection-status-info');
        
        if (statusElement) {
            statusElement.textContent = message;
        }
        
        if (statusContainer) {
            // 移除之前的状态类
            statusContainer.classList.remove('success', 'error', 'warning');
            
            // 添加新的状态类
            if (type) {
                statusContainer.classList.add(type);
            }
        }
    }

    applyAccessKeyConfig() {
        const apiUrl = document.getElementById('api-url-input').value.trim();
        const accessKey = document.getElementById('access-key-input').value.trim();
        const saveToLocal = document.getElementById('save-to-local-checkbox')?.checked || false;

        if (!apiUrl) {
            this.updateConnectionStatus('Error: API URL is required', 'error');
            return;
        }

        // 如果用户选择保存到本地，则保存API地址到localStorage
        if (saveToLocal) {
            localStorage.setItem('serverStatus_api_url', apiUrl);
            if (CONFIG.DEBUG) {
                console.log('API地址已保存到localStorage:', apiUrl);
            }
        } else {
            // 如果用户取消了保存选项，清除之前保存的API地址
            localStorage.removeItem('serverStatus_api_url');
            if (CONFIG.DEBUG) {
                console.log('已清除localStorage中的API地址');
            }
        }

        // 保存访问密钥到localStorage（如果有的话）
        if (accessKey) {
            localStorage.setItem('serverStatus_access_key', accessKey);
        } else {
            localStorage.removeItem('serverStatus_access_key');
        }

        // 构建新的URL参数
        const params = new URLSearchParams();
        params.set('api', apiUrl);
        
        if (accessKey) {
            params.set('key', accessKey);
        }

        // 重新加载页面，应用新配置
        const newUrl = `${window.location.pathname}?${params.toString()}`;
        window.location.href = newUrl;
    }

    resetToDefaultConfig() {
        // 清除localStorage中保存的配置
        localStorage.removeItem('serverStatus_api_url');
        localStorage.removeItem('serverStatus_access_key');
        
        if (CONFIG.DEBUG) {
            console.log('已清除localStorage中的配置，重置为默认设置');
        }

        // 更新连接状态
        this.updateConnectionStatus('Configuration reset to default', 'success');

        // 重置表单字段为默认值
        document.getElementById('api-url-input').value = 'http://localhost:8080';
        document.getElementById('access-key-input').value = '';
        
        const saveToLocalCheckbox = document.getElementById('save-to-local-checkbox');
        if (saveToLocalCheckbox) {
            saveToLocalCheckbox.checked = false;
        }

        // 如果当前页面有URL参数，刷新页面移除参数
        if (window.location.search) {
            setTimeout(() => {
                window.location.href = window.location.pathname;
            }, 1000); // 1秒后刷新，让用户看到成功消息
        }
    }

    updateAccessKeyTexts() {
        // 更新访问密钥模态框的所有文本
        this.updateElementText('access-key-modal-title', this.t('accessKey.modalTitle'));
        this.updateElementText('current-config-title', this.t('accessKey.currentConfig'));
        this.updateElementText('api-server-label', this.t('accessKey.apiServer'));
        this.updateElementText('access-mode-label', this.t('accessKey.accessMode'));
        this.updateElementText('current-key-label', this.t('accessKey.currentKey'));
        this.updateElementText('update-config-title', this.t('accessKey.updateConfig'));
        this.updateElementText('api-url-label', this.t('accessKey.apiServerUrl'));
        this.updateElementText('access-key-label', this.t('accessKey.accessKeyLabel'));
        this.updateElementText('access-key-help', this.t('accessKey.accessKeyHelp'));
        this.updateElementText('test-connection-btn', this.t('accessKey.testConnection'));
        this.updateElementText('apply-config-btn', this.t('accessKey.applyReload'));
        this.updateElementText('connection-status-title', this.t('accessKey.connectionStatus'));
        this.updateElementText('help-section-title', this.t('accessKey.helpExamples'));
        this.updateElementText('help-public-mode', `<strong>${this.t('accessKey.publicMode')}</strong> ${this.t('accessKey.publicModeDesc')}`);
        this.updateElementText('help-project-key', `<strong>${this.t('accessKey.projectKeyMode')}</strong> ${this.t('accessKey.projectKeyModeDesc')}`);
        this.updateElementText('help-access-key', `<strong>${this.t('accessKey.accessKeyMode')}</strong> ${this.t('accessKey.accessKeyModeDesc')}`);
        this.updateElementText('help-url-params', `<strong>${this.t('accessKey.urlParams')}</strong>`);
        
        // 更新输入框的placeholder
        const apiUrlInput = document.getElementById('api-url-input');
        if (apiUrlInput) {
            apiUrlInput.placeholder = 'http://localhost:8080';
        }
        
        const accessKeyInput = document.getElementById('access-key-input');
        if (accessKeyInput) {
            accessKeyInput.placeholder = this.t('accessKey.accessKeyPlaceholder');
        }
    }
    
    setupThemeSwitching() {
        const themeBtn = document.getElementById('theme-toggle');
        if (themeBtn) {
            themeBtn.addEventListener('click', () => {
                this.toggleTheme();
            });
        }
        
        // 初始化主题
        const savedTheme = localStorage.getItem('preferred-theme') || 'light';
        this.currentTheme = savedTheme;
        this.applyTheme();
    }
    
    toggleTheme() {
        this.currentTheme = this.currentTheme === 'light' ? 'dark' : 'light';
        localStorage.setItem('preferred-theme', this.currentTheme);
        this.applyTheme();
    }
    
    applyTheme() {
        const html = document.documentElement;
        const themeIcon = document.getElementById('theme-icon');
        
        if (this.currentTheme === 'dark') {
            html.setAttribute('data-theme', 'dark');
            if (themeIcon) {
                themeIcon.innerHTML = `
                    <path d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zM2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1zm18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1zM11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1zm0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1zM5.99 4.58c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0s.39-1.03 0-1.41L5.99 4.58zm12.37 12.37c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0 .39-.39.39-1.03 0-1.41l-1.06-1.06zm1.06-10.96c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0l1.06-1.06zM7.05 18.36c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0l1.06-1.06z"/>
                `;
            }
        } else {
            html.removeAttribute('data-theme');
            if (themeIcon) {
                themeIcon.innerHTML = `
                    <path d="M12 18c-.89 0-1.74-.19-2.5-.54C11.56 16.5 13 14.42 13 12s-1.44-4.5-3.5-5.46C10.26 6.19 11.11 6 12 6c3.31 0 6 2.69 6 6s-2.69 6-6 6z"/>
                `;
            }
        }
    }
    
    setupLanguageSwitching() {
        const langButtons = document.querySelectorAll('.lang-btn');
        langButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                const lang = btn.id.includes('zh') ? 'zh' : 'en';
                this.switchLanguage(lang);
            });
        });
    }
    
    switchLanguage(lang) {
        if (this.language === lang) return;
        
        this.language = lang;
        localStorage.setItem('preferred-language', lang);
        
        // 更新语言按钮状态
        document.querySelectorAll('.lang-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.getElementById(`lang-${lang}`).classList.add('active');
        
        // 应用语言更改
        this.applyLanguage();
    }
    
    applyLanguage() {
        // 更新页面语言属性
        document.documentElement.lang = this.language;
        
        // 更新所有文本内容
        this.updateAllTexts();
        this.updateTutorialLanguage();
        
        // 重新渲染服务器卡片以更新文本
        if (this.servers.size > 0) {
            const serversArray = Array.from(this.servers.values());
            this.renderServerCards(serversArray);
        }
    }
    
    setupGridControls() {
        const gridButtons = document.querySelectorAll('.grid-size-btn');
        const serversGrid = document.getElementById('servers-grid');
        
        gridButtons.forEach(button => {
            button.addEventListener('click', () => {
                // 移除所有按钮的active状态
                gridButtons.forEach(btn => btn.classList.remove('active'));
                // 添加当前按钮的active状态
                button.classList.add('active');
                
                // 移除网格的所有布局类
                serversGrid.className = 'servers-grid';
                // 添加新的布局类
                const gridType = button.dataset.grid;
                serversGrid.classList.add(gridType);
                
                // 保存用户选择到localStorage
                localStorage.setItem('gridLayout', gridType);
            });
        });
        
        // 从localStorage恢复用户选择的布局
        const savedLayout = localStorage.getItem('gridLayout') || 'auto';
        const savedButton = document.querySelector(`[data-grid="${savedLayout}"]`);
        if (savedButton) {
            savedButton.click();
        }
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
            button.textContent = this.t('tutorial.copied');
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
            
            button.textContent = this.t('tutorial.selected');
            setTimeout(() => {
                button.textContent = this.t('tutorial.copyBtn');
            }, 2000);
        });
    }

    toggleTutorial() {
        const tutorialSection = document.getElementById('tutorial-section');
        const toggleBtn = tutorialSection.querySelector('.toggle-tutorial');
        
        if (tutorialSection.classList.contains('collapsed')) {
            tutorialSection.classList.remove('collapsed');
            toggleBtn.textContent = this.t('tutorial.hideTutorial');
        } else {
            tutorialSection.classList.add('collapsed');
            toggleBtn.textContent = this.t('tutorial.showTutorial');
        }
    }



    updateConnectionStatus(connected, customMessage = null) {
        const statusElement = document.getElementById('connection-status');
        if (connected) {
            statusElement.textContent = this.t('connectionStatus.online');
            statusElement.className = 'status online';
        } else {
            statusElement.textContent = customMessage || this.t('connectionStatus.offline');
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
                <h3>${this.t('errors.authFailed')}</h3>
                <p>${this.t('errors.authFailedDesc')}</p>
                <p style="font-size: 0.9em; color: #666; margin-top: 20px;">
                    ${this.t('errors.supportedFormats')}<br>
                    • ?access=your-access-key (${this.language === 'zh' ? '访问密钥模式' : 'Access Key Mode'})<br>
                    • ?key=your-team-key (${this.language === 'zh' ? 'API密钥模式' : 'API Key Mode'})<br>
                    • ?token=your-token (${this.language === 'zh' ? '访问令牌模式' : 'Access Token Mode'})
                </p>
            </div>
        `;
    }

    updateServers(serversData) {
        // 更新服务器数据
        serversData.forEach(server => {
            // 检查告警
            const alerts = this.checkServerAlerts(server);
            alerts.forEach(alert => this.addAlert(alert));
            
            this.servers.set(server.hostname, server);
        });

        // 更新服务器计数
        document.getElementById('server-count').textContent = `${serversData.length}${this.t('serverCount')}`;
        
        // 更新性能统计
        this.updatePerformanceStats(serversData);
        
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
        // 更新服务器数据并应用筛选
        this.filteredServers = servers.slice(); // 创建副本
        this.applyFilters();
    }

    createServerCard(server) {
        const isOnline = server.status === 'online';
        const lastSeenText = isOnline ? this.t('server.justNow') : this.formatTimeAgo(server.last_seen);
        
        // 获取分组信息
        const groupId = this.serverGroups.get(server.hostname);
        const group = groupId ? this.groups.get(groupId) : null;
        
        return `
            <div class="server-card ${server.status} ${group ? 'grouped' : ''}" data-hostname="${server.hostname}" data-session="${server.session_id || ''}" ${group ? `style="--group-color: ${group.color}; --group-color-light: ${group.color}08;"` : ''}>
                ${group ? `<div class="group-indicator" style="background-color: ${group.color}" title="${group.name}"></div>` : ''}
                <div class="server-header">
                    <div class="server-name">
                        ${server.hostname}${server.session_id ? ` (${server.session_id.substring(0, 8)})` : ''}
                        ${group ? `<div class="group-label" style="background-color: ${group.color}20; color: ${group.color}; border-color: ${group.color}40;">${group.name}</div>` : ''}
                    </div>
                    <div class="server-status ${server.status}">${isOnline ? this.t('server.online') : this.t('server.offline')}</div>
                </div>
                <div class="metrics">
                    <div class="metric">
                        <div class="metric-label">${this.t('metrics.cpu')}</div>
                        <div class="metric-value cpu ${server.cpu_percent > 80 ? 'high' : ''}">${server.cpu_percent.toFixed(1)}%</div>
                        <div class="progress-bar">
                            <div class="progress-fill cpu ${server.cpu_percent > 80 ? 'high' : ''}" style="width: ${server.cpu_percent}%"></div>
                        </div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">${this.t('metrics.memory')}</div>
                        <div class="metric-value memory ${server.memory_percent > 80 ? 'high' : ''}">${server.memory_percent.toFixed(1)}%</div>
                        <div class="progress-bar">
                            <div class="progress-fill memory ${server.memory_percent > 80 ? 'high' : ''}" style="width: ${server.memory_percent}%"></div>
                        </div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">${this.t('metrics.disk')}</div>
                        <div class="metric-value disk ${server.disk_percent > 80 ? 'high' : ''}">${server.disk_percent.toFixed(1)}%</div>
                        <div class="progress-bar">
                            <div class="progress-fill disk ${server.disk_percent > 80 ? 'high' : ''}" style="width: ${server.disk_percent}%"></div>
                        </div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">${this.t('metrics.temperature')}</div>
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
                    <div class="metric">
                        <div class="metric-label">${this.t('metrics.network')}</div>
                        <div class="network-speeds">
                            <div class="network-speed">
                                <span class="speed-icon">↑</span>
                                <span class="speed-value ${this.getNetworkSpeedClass(server.network_speed_sent || 0)}">${this.formatNetworkSpeed(server.network_speed_sent || 0)}</span>
                            </div>
                            <div class="network-speed">
                                <span class="speed-icon">↓</span>
                                <span class="speed-value ${this.getNetworkSpeedClass(server.network_speed_recv || 0)}">${this.formatNetworkSpeed(server.network_speed_recv || 0)}</span>
                            </div>
                        </div>
                        <div class="network-total">
                            ${this.t('metrics.total')}: ↑${this.formatBytes(server.network_bytes_sent || 0)} ↓${this.formatBytes(server.network_bytes_recv || 0)}
                        </div>
                    </div>
                </div>
                <div class="server-info">
                    ${this.t('server.lastUpdate')}${lastSeenText}
                </div>
                ${group && this.viewMode === 'cards' ? `<div class="group-label">${group.name}</div>` : ''}
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

        if (diffSecs < 60) return `${diffSecs}${this.t('server.secondsAgo')}`;
        if (diffMins < 60) return `${diffMins}${this.t('server.minutesAgo')}`;
        if (diffHours < 24) return `${diffHours}${this.t('server.hoursAgo')}`;
        return time.toLocaleDateString();
    }
    
    formatNetworkSpeed(speedKBps) {
        if (speedKBps < 1) {
            return `${(speedKBps * 1024).toFixed(0)} B/s`;
        } else if (speedKBps < 1024) {
            return `${speedKBps.toFixed(1)} KB/s`;
        } else {
            return `${(speedKBps / 1024).toFixed(1)} MB/s`;
        }
    }
    
    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    getNetworkSpeedClass(speedKBps) {
        if (speedKBps > 10240) { // > 10 MB/s
            return 'very-high';
        } else if (speedKBps > 1024) { // > 1 MB/s
            return 'high';
        }
        return '';
    }

    async showServerDetails(hostname, sessionId) {
        this.selectedServer = hostname;
        this.selectedSessionId = sessionId;
        const displayName = sessionId ? `${hostname} (${sessionId.substring(0, 8)})` : hostname;
        document.getElementById('selected-server-name').textContent = `${displayName} - ${this.t('server.detailed')}`;
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
                url = `${this.apiBaseUrl}/api/access/${this.accessKey}/server-by-session/${sessionId}`;
            } else if (this.accessKey) {
                // 如果只有accessKey，使用hostname-based API
                url = `${this.apiBaseUrl}/api/access/${this.accessKey}/server/${hostname}`;
            } else {
                // 默认API
                url = `${this.apiBaseUrl}/api/server/${hostname}`;
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
                label: `${this.t('metrics.cpuUsage')} (%)`,
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
                label: `${this.t('metrics.memoryUsage')} (%)`,
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
                label: `${this.t('metrics.diskUsage')} (%)`,
                data: history.map(item => item.disk.usage_percent),
                borderColor: '#f59e0b',
                backgroundColor: 'rgba(245, 158, 11, 0.1)',
                tension: 0.4,
                fill: true
            }]
        });

        // 网络速度图表
        this.updateChart('network-speed-chart', {
            labels: labels,
            datasets: [{
                label: `${this.t('metrics.networkSpeed')} ↑ (KB/s)`,
                data: history.map(item => (item.network && item.network.speed_sent) ? item.network.speed_sent : 0),
                borderColor: '#06b6d4',
                backgroundColor: 'rgba(6, 182, 212, 0.1)',
                tension: 0.4,
                fill: true
            }, {
                label: `${this.t('metrics.networkSpeed')} ↓ (KB/s)`,
                data: history.map(item => (item.network && item.network.speed_recv) ? item.network.speed_recv : 0),
                borderColor: '#10b981',
                backgroundColor: 'rgba(16, 185, 129, 0.1)',
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
                        label: `GPU${i + 1} ${this.t('metrics.gpuUsage')} (%)`,
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
                        label: `GPU${i + 1} ${this.t('metrics.gpuMemoryUsage')} (%)`,
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
                label: `CPU ${this.t('metrics.temperature')} (°C)`,
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
                        label: `GPU${i + 1} ${this.t('metrics.temperature')} (°C)`,
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
                label: `GPU ${this.t('metrics.temperature')} (°C)`,
                data: history.map(item => item.temperature ? item.temperature.gpu_temp : 0),
                borderColor: '#8b5cf6',
                backgroundColor: 'rgba(139, 92, 246, 0.1)',
                tension: 0.4,
                fill: false
            });
        }
        if (history.some(item => item.temperature && item.temperature.max_temp > 0)) {
            tempDatasets.push({
                label: `${this.language === 'zh' ? '最高' : 'Max'} ${this.t('metrics.temperature')} (°C)`,
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
            return `${days}${this.t('systemInfo.days')} ${hours}${this.t('systemInfo.hours')} ${minutes}${this.t('systemInfo.minutes')}`;
        };

        const systemInfoHtml = `
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.hostname')}</span>
                <span class="info-value">${systemInfo.hostname}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.os')}</span>
                <span class="info-value">${systemInfo.os.platform} ${systemInfo.os.version}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.arch')}</span>
                <span class="info-value">${systemInfo.os.arch}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.cpuModel')}</span>
                <span class="info-value">${systemInfo.cpu.model_name || 'Unknown'}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.cpuCores')}</span>
                <span class="info-value">${systemInfo.cpu.core_count}${this.t('systemInfo.cores')}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.totalMemory')}</span>
                <span class="info-value">${formatBytes(systemInfo.memory.total)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.usedMemory')}</span>
                <span class="info-value">${formatBytes(systemInfo.memory.used)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.totalDisk')}</span>
                <span class="info-value">${formatBytes(systemInfo.disk.total)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.usedDisk')}</span>
                <span class="info-value">${formatBytes(systemInfo.disk.used)}</span>
            </div>
            <div class="info-item">
                <span class="info-label">${this.t('systemInfo.uptime')}</span>
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
                <span class="info-label">${this.t('systemInfo.lastUpdate')}</span>
                <span class="info-value">${new Date(systemInfo.timestamp).toLocaleString()}</span>
            </div>
            ${systemInfo.temperature && systemInfo.temperature.cpu_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">CPU ${this.t('metrics.temperature')}:</span>
                <span class="info-value">${systemInfo.temperature.cpu_temp.toFixed(1)}°C</span>
            </div>` : ''}
            ${systemInfo.temperature && systemInfo.temperature.gpu_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">GPU ${this.t('metrics.temperature')}:</span>
                <span class="info-value">${systemInfo.temperature.gpu_temp.toFixed(1)}°C</span>
            </div>` : ''}
            ${systemInfo.temperature && systemInfo.temperature.max_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">${this.language === 'zh' ? '最高' : 'Max'} ${this.t('metrics.temperature')}:</span>
                <span class="info-value">${systemInfo.temperature.max_temp.toFixed(1)}°C</span>
            </div>` : ''}
            ${systemInfo.temperature && systemInfo.temperature.avg_temp > 0 ? `
            <div class="info-item">
                <span class="info-label">${this.language === 'zh' ? '平均' : 'Avg'} ${this.t('metrics.temperature')}:</span>
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
            const response = await fetch(this.apiBaseUrl + '/api/uuid-count');
            if (response.ok) {
                const data = await response.json();
                document.getElementById('uuid-count').textContent = `${data.active_uuids}${this.t('uuidDevices')}`;
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