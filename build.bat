@echo off
echo ========================================
echo    Obscura GPU Monitor 编译脚本
echo ========================================
echo.

echo 正在初始化Go模块...
go mod tidy
if %errorlevel% neq 0 (
    echo 错误: Go模块初始化失败
    pause
    exit /b 1
)

echo.
echo 正在创建build目录...
if not exist "build" mkdir build

echo.
echo 正在编译监控端 (monitor-agent)...
go build -o build\monitor-agent.exe .\monitor-agent
if %errorlevel% neq 0 (
    echo 错误: 监控端编译失败
    pause
    exit /b 1
)
echo ✓ 监控端编译成功: build\monitor-agent.exe

echo.
echo 正在编译数据收集端 (data-server)...
go build -o build\data-server.exe .\data-server
if %errorlevel% neq 0 (
    echo 错误: 数据收集端编译失败
    pause
    exit /b 1
)
echo ✓ 数据收集端编译成功: build\data-server.exe

echo.
echo ========================================
echo           编译完成！
echo ========================================
echo.
echo 使用方法:
echo 1. 启动数据收集端: build\data-server.exe
echo 2. 启动监控端: build\monitor-agent.exe
echo 3. 打开浏览器访问: http://localhost:8080
echo.
echo 按任意键退出...
pause >nul