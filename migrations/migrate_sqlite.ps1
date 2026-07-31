# FastToken OAuth 精简 - PowerShell 数据库迁移脚本
# 使用 System.Data.SQLite 或纯文本操作

$ErrorActionPreference = "Stop"

# 数据库路径
$dbPath = "Z:\Dev\FastToken\one-api.db"
$backupPath = "Z:\Dev\FastToken\one-api-backup-20260530.db"

Write-Host "=" * 60
Write-Host "FastToken OAuth 精简 - 数据库迁移脚本"
Write-Host "=" * 60
Write-Host ""

# 检查数据库文件
if (-not (Test-Path $dbPath)) {
    Write-Host "❌ 错误: 数据库文件不存在: $dbPath" -ForegroundColor Red
    exit 1
}

$fileInfo = Get-Item $dbPath
Write-Host "✅ 找到数据库文件: $dbPath"
Write-Host "   文件大小: $($fileInfo.Length) 字节"
Write-Host ""

# 检查备份
if (Test-Path $backupPath) {
    Write-Host "✅ 备份文件已存在: $backupPath" -ForegroundColor Green
} else {
    Write-Host "⚠️  警告: 备份文件不存在" -ForegroundColor Yellow
}
Write-Host ""

# 读取数据库文件（二进制方式）
Write-Host "📖 读取数据库文件..."
$dbBytes = [System.IO.File]::ReadAllBytes($dbPath)
$dbText = [System.Text.Encoding]::UTF8.GetString($dbBytes)

Write-Host "✅ 数据库已加载到内存"
Write-Host ""

# 检查要删除的表
Write-Host "🔍 检查要删除的表..."

$tablesToDelete = @("custom_oauth_providers", "user_oauth_bindings")
$foundTables = @()

foreach ($table in $tablesToDelete) {
    if ($dbText -match [regex]::Escape("CREATE TABLE $table")) {
        $foundTables += $table
        Write-Host "   找到: $table" -ForegroundColor Yellow
    } else {
        Write-Host "   不存在: $table" -ForegroundColor Gray
    }
}

Write-Host ""

if ($foundTables.Count -eq 0) {
    Write-Host "ℹ️  要删除的表不存在，无需迁移" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "=" * 60
    Write-Host "✅ 迁移完成（无需操作）" -ForegroundColor Green
    Write-Host "=" * 60
    exit 0
}

Write-Host "📋 找到 $($foundTables.Count) 个要删除的表"
Write-Host ""

# 由于 SQLite 是二进制格式，我们需要使用 sqlite3 工具
# 检查是否有可用的 sqlite3

Write-Host "🔍 查找 sqlite3 工具..."

$sqlitePaths = @(
    "sqlite3.exe",
    "C:\ProgramData\chocolatey\bin\sqlite3.exe",
    "C:\tools\sqlite3\sqlite3.exe",
    "$env:USERPROFILE\scoop\apps\sqlite\current\sqlite3.exe",
    "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\sqlite3.exe"
)

$sqliteExe = $null
foreach ($path in $sqlitePaths) {
    if (Test-Path $path) {
        $sqliteExe = $path
        Write-Host "✅ 找到 sqlite3: $path" -ForegroundColor Green
        break
    }
}

# 尝试从 PATH 中查找
if (-not $sqliteExe) {
    try {
        $sqliteExe = (Get-Command sqlite3 -ErrorAction SilentlyContinue).Source
        if ($sqliteExe) {
            Write-Host "✅ 找到 sqlite3: $sqliteExe" -ForegroundColor Green
        }
    } catch {}
}

if (-not $sqliteExe) {
    Write-Host ""
    Write-Host "❌ 未找到 sqlite3 工具" -ForegroundColor Red
    Write-Host ""
    Write-Host "请安装 sqlite3 工具后重试:" -ForegroundColor Yellow
    Write-Host "  方式 1: choco install sqlite -y"
    Write-Host "  方式 2: scoop install sqlite"
    Write-Host "  方式 3: 下载 https://www.sqlite.org/download.html"
    Write-Host ""
    Write-Host "或者使用图形化工具:" -ForegroundColor Yellow
    Write-Host "  DB Browser for SQLite: https://sqlitebrowser.org/"
    Write-Host ""
    
    # 提供手动迁移 SQL
    Write-Host "手动迁移 SQL 语句:" -ForegroundColor Cyan
    Write-Host "  DROP TABLE IF EXISTS custom_oauth_providers;" -ForegroundColor White
    Write-Host "  DROP TABLE IF EXISTS user_oauth_bindings;" -ForegroundColor White
    Write-Host ""
    
    exit 1
}

# 执行迁移
Write-Host ""
Write-Host "🗑️  开始删除表..." -ForegroundColor Cyan

$sql = "DROP TABLE IF EXISTS custom_oauth_providers; DROP TABLE IF EXISTS user_oauth_bindings;"

try {
    $result = & $sqliteExe $dbPath $sql 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ 表删除成功！" -ForegroundColor Green
    } else {
        Write-Host "❌ 删除失败: $result" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ 执行失败: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 验证结果
Write-Host "🔍 验证删除结果..."

$verifySql = "SELECT name FROM sqlite_master WHERE type='table' AND name IN ('custom_oauth_providers', 'user_oauth_bindings');"
$verifyResult = & $sqliteExe $dbPath $verifySql 2>&1

if ([string]::IsNullOrWhiteSpace($verifyResult)) {
    Write-Host "✅ 验证成功：表已被完全删除" -ForegroundColor Green
} else {
    Write-Host "⚠️  警告：仍有残留表: $verifyResult" -ForegroundColor Yellow
}

# 显示所有表
Write-Host ""
Write-Host "📊 当前数据库中的所有表:"

$listSql = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"
$tables = & $sqliteExe $dbPath $listSql 2>&1

$i = 1
foreach ($table in $tables -split "`n") {
    if (-not [string]::IsNullOrWhiteSpace($table)) {
        Write-Host "   $i. $table"
        $i++
    }
}

Write-Host ""
Write-Host "=" * 60
Write-Host "✅ 迁移完成！" -ForegroundColor Green
Write-Host "=" * 60
Write-Host ""
Write-Host "📝 后续步骤:"
Write-Host "   1. 重启 FastToken 服务"
Write-Host "   2. 测试登录功能（邮箱、手机号、微信、GitHub）"
Write-Host "   3. 检查用户数据是否正常"
Write-Host ""
