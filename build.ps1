# GPU Monitor Auto Build Script (PowerShell)
# Supports Windows, Linux and macOS compilation

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "GPU Monitor Auto Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host

Write-Host "Cleaning old build files..." -ForegroundColor Yellow
if (Test-Path "release\data-server.exe") { Remove-Item "release\data-server.exe" }
if (Test-Path "release\monitor-agent.exe") { Remove-Item "release\monitor-agent.exe" }
if (Test-Path "release\data-server-linux") { Remove-Item "release\data-server-linux" }
if (Test-Path "release\monitor-agent-linux") { Remove-Item "release\monitor-agent-linux" }
if (Test-Path "release\data-server-darwin") { Remove-Item "release\data-server-darwin" }
if (Test-Path "release\monitor-agent-darwin") { Remove-Item "release\monitor-agent-darwin" }
if (Test-Path "release\data-server-darwin-arm64") { Remove-Item "release\data-server-darwin-arm64" }
if (Test-Path "release\monitor-agent-darwin-arm64") { Remove-Item "release\monitor-agent-darwin-arm64" }

Write-Host "Creating release directory..." -ForegroundColor Yellow
if (!(Test-Path "release")) { New-Item -ItemType Directory -Path "release" | Out-Null }

Write-Host
Write-Host "[1/8] Building Windows data-server..." -ForegroundColor Green
Set-Location "data-server"
$result = & go build -o "../release/data-server.exe"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Windows data-server build failed" -ForegroundColor Red
    Set-Location ".."
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: Windows data-server built" -ForegroundColor Green

Write-Host
Write-Host "[2/8] Building Windows monitor-agent..." -ForegroundColor Green
Set-Location "monitor-agent"
$result = & go build -o "../release/monitor-agent.exe"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Windows monitor-agent build failed" -ForegroundColor Red
    Set-Location ".."
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: Windows monitor-agent built" -ForegroundColor Green

Write-Host
Write-Host "[3/8] Setting Linux environment and building data-server..." -ForegroundColor Green
& go env -w GOOS=linux
Set-Location "data-server"
$result = & go build -o "../release/data-server-linux" "./main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Linux data-server build failed" -ForegroundColor Red
    Set-Location ".."
    & go env -w GOOS=windows
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: Linux data-server built" -ForegroundColor Green

Write-Host
Write-Host "[4/8] Building Linux monitor-agent..." -ForegroundColor Green
Set-Location "monitor-agent"
$result = & go build -o "../release/monitor-agent-linux" "./main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Linux monitor-agent build failed" -ForegroundColor Red
    Set-Location ".."
    & go env -w GOOS=windows
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: Linux monitor-agent built" -ForegroundColor Green

Write-Host
Write-Host "[5/8] Setting macOS environment and building data-server (amd64)..." -ForegroundColor Green
& go env -w GOOS=darwin GOARCH=amd64
Set-Location "data-server"
$result = & go build -o "../release/data-server-darwin" "./main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: macOS data-server build failed" -ForegroundColor Red
    Set-Location ".."
    & go env -w GOOS=windows GOARCH=amd64
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: macOS data-server built" -ForegroundColor Green

Write-Host
Write-Host "[6/8] Building macOS monitor-agent (amd64)..." -ForegroundColor Green
Set-Location "monitor-agent"
$result = & go build -o "../release/monitor-agent-darwin" "./main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: macOS monitor-agent build failed" -ForegroundColor Red
    Set-Location ".."
    & go env -w GOOS=windows GOARCH=amd64
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: macOS monitor-agent built" -ForegroundColor Green

Write-Host
Write-Host "[7/8] Building macOS data-server (arm64)..." -ForegroundColor Green
& go env -w GOOS=darwin GOARCH=arm64
Set-Location "data-server"
$result = & go build -o "../release/data-server-darwin-arm64" "./main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: macOS ARM64 data-server build failed" -ForegroundColor Red
    Set-Location ".."
    & go env -w GOOS=windows GOARCH=amd64
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: macOS ARM64 data-server built" -ForegroundColor Green

Write-Host
Write-Host "[8/8] Building macOS monitor-agent (arm64)..." -ForegroundColor Green
Set-Location "monitor-agent"
$result = & go build -o "../release/monitor-agent-darwin-arm64" "./main.go"
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: macOS ARM64 monitor-agent build failed" -ForegroundColor Red
    Set-Location ".."
    & go env -w GOOS=windows GOARCH=amd64
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ".."
Write-Host "SUCCESS: macOS ARM64 monitor-agent built" -ForegroundColor Green

Write-Host
Write-Host "Restoring Windows environment..." -ForegroundColor Yellow
& go env -w GOOS=windows GOARCH=amd64

Write-Host
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Complete! Files location:" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Get-ChildItem "release" | Select-Object Name
Write-Host
Write-Host "All programs successfully built to release directory" -ForegroundColor Green
Write-Host "Includes multi-GPU monitoring functionality" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Read-Host "Press Enter to exit"