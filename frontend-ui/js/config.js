// ServerStatus Frontend Configuration
const CONFIG = {
    // API服务器配置
    API_BASE_URL: 'http://localhost:8080',  // 后端API服务器地址
    
    // 轮询配置
    POLL_INTERVAL: 3000,                    // 数据轮询间隔(毫秒)
    
    // UI配置
    DEFAULT_LANGUAGE: 'auto',               // 默认语言 'zh', 'en', 'auto'
    DEFAULT_THEME: 'light',                 // 默认主题 'light', 'dark'
    DEFAULT_GRID_SIZE: 'auto',              // 默认网格大小 'auto', 'grid-3x3', 'grid-4x4', 'grid-5x5'
    
    // 告警阈值
    ALERT_THRESHOLDS: {
        cpu: 85,                            // CPU使用率告警阈值(%)
        memory: 90,                         // 内存使用率告警阈值(%)
        disk: 95,                           // 磁盘使用率告警阈值(%)
        temperature: 75                     // 温度告警阈值(°C)
    },
    
    // 缓存配置
    CACHE_DURATION: 60000,                  // 缓存持续时间(毫秒)
    
    // 网络监控
    NETWORK_SPEED_UNITS: 'auto',            // 网络速度单位 'KB/s', 'MB/s', 'auto'
    HIGH_SPEED_THRESHOLD: 10240,            // 高速网络阈值(KB/s, 10MB/s)
    VERY_HIGH_SPEED_THRESHOLD: 102400,      // 超高速网络阈值(KB/s, 100MB/s)
    
    // 访问密钥模式
    ACCESS_KEY_MODE: false,                 // 是否启用访问密钥模式
    ACCESS_KEY: '',                         // 访问密钥(从URL参数获取)
    
    // 调试模式
    DEBUG: false,                           // 调试模式开关
    
    // 功能开关
    FEATURES: {
        serverGroups: true,                 // 服务器分组功能
        dataExport: true,                   // 数据导出功能
        networkMonitoring: true,            // 网络监控功能
        gpuMonitoring: true,                // GPU监控功能
        temperatureMonitoring: true,        // 温度监控功能
        alerts: true,                       // 告警功能
        keyboardShortcuts: true,            // 键盘快捷键
        themeSwitching: true,               // 主题切换
        languageSwitching: true             // 语言切换
    }
};

// 从URL参数和localStorage读取配置
(function initializeConfig() {
    const urlParams = new URLSearchParams(window.location.search);
    
    // 读取API服务器地址（优先级：URL参数 > localStorage > 默认值）
    const apiUrl = urlParams.get('api') || urlParams.get('server');
    const savedApiUrl = localStorage.getItem('serverStatus_api_url');
    
    if (apiUrl) {
        // URL参数优先级最高
        CONFIG.API_BASE_URL = apiUrl.replace(/\/$/, ''); // 移除末尾斜杠
        if (CONFIG.DEBUG) {
            console.log('从URL参数设置API地址:', CONFIG.API_BASE_URL);
        }
    } else if (savedApiUrl) {
        // 其次使用localStorage中保存的地址
        CONFIG.API_BASE_URL = savedApiUrl.replace(/\/$/, ''); // 移除末尾斜杠
        if (CONFIG.DEBUG) {
            console.log('从localStorage读取API地址:', CONFIG.API_BASE_URL);
        }
    }
    
    // 读取访问密钥（优先级：URL参数 > localStorage > 默认值）
    const accessKey = urlParams.get('key') || urlParams.get('access_key');
    const savedAccessKey = localStorage.getItem('serverStatus_access_key');
    
    if (accessKey) {
        // URL参数优先级最高
        CONFIG.ACCESS_KEY_MODE = true;
        CONFIG.ACCESS_KEY = accessKey;
        if (CONFIG.DEBUG) {
            console.log('从URL参数设置访问密钥:', CONFIG.ACCESS_KEY);
        }
    } else if (savedAccessKey) {
        // 其次使用localStorage中保存的访问密钥
        CONFIG.ACCESS_KEY_MODE = true;
        CONFIG.ACCESS_KEY = savedAccessKey;
        if (CONFIG.DEBUG) {
            console.log('从localStorage读取访问密钥:', CONFIG.ACCESS_KEY);
        }
    }
    
    // 读取调试模式
    const debug = urlParams.get('debug');
    if (debug === 'true' || debug === '1') {
        CONFIG.DEBUG = true;
    }
    
    // 读取语言设置
    const lang = urlParams.get('lang') || urlParams.get('language');
    if (lang && ['zh', 'en'].includes(lang)) {
        CONFIG.DEFAULT_LANGUAGE = lang;
    }
    
    // 读取主题设置
    const theme = urlParams.get('theme');
    if (theme && ['light', 'dark'].includes(theme)) {
        CONFIG.DEFAULT_THEME = theme;
    }
})();

// 调试日志
if (CONFIG.DEBUG) {
    console.log('ServerStatus Frontend Config:', CONFIG);
}

// 导出配置
window.CONFIG = CONFIG;