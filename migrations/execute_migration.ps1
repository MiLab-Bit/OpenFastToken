# FastToken OAuth 精简 - 直接 SQL 执行脚本
# 使用 DB Browser for SQLite 或其他工具执行以下 SQL

$dbPath = "Z:\Dev\FastToken\one-api.db"
$backupPath = "Z:\Dev\FastToken\one-api-backup-20260530.db"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "FastToken OAuth 精简 - 数据库迁移" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

# 检查数据库文件
if (Test-Path $dbPath) {
    $fileInfo = Get-Item $dbPath
    Write-Host "✅ 数据库文件: $dbPath" -ForegroundColor Green
    Write-Host "   大小: $($fileInfo.Length) 字节" -ForegroundColor Gray
} else {
    Write-Host "❌ 数据库文件不存在" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "📝 需要执行的 SQL 语句:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  DROP TABLE IF EXISTS custom_oauth_providers;" -ForegroundColor White
Write-Host "  DROP TABLE IF EXISTS user_oauth_bindings;" -ForegroundColor White
Write-Host ""

# 检查是否有可用的工具
Write-Host "🔍 检查可用的 SQLite 工具..." -ForegroundColor Cyan

$tools = @()

# 检查 sqlite3 命令
try {
    $sqlite3 = Get-Command sqlite3 -ErrorAction SilentlyContinue
    if ($sqlite3) {
        $tools += "sqlite3 (命令行)"
    }
} catch {}

# 检查 DB Browser
$dbBrowserPaths = @(
    "${env:ProgramFiles}\DB Browser for SQLite\DB Browser for SQLite.exe",
    "${env:ProgramFiles(x86)}\DB Browser for SQLite\DB Browser for SQLite.exe",
    "${env:LocalAppData}\Programs\DB Browser for SQLite\DB Browser for SQLite.exe"
)

foreach ($path in $dbBrowserPaths) {
    if (Test-Path $path) {
        $tools += "DB Browser for SQLite"
        break
    }
}

# 检查 Python + sqlite3
try {
    $python = Get-Command python -ErrorAction SilentlyContinue
    if ($python) {
        $result = python -c "import sqlite3; print('ok')" 2>$null
        if ($result -eq "ok") {
            $tools += "Python + sqlite3"
        }
    }
} catch {}

if ($tools.Count -gt 0) {
    Write-Host "✅ 找到可用工具:" -ForegroundColor Green
    foreach ($tool in $tools) {
        Write-Host "   - $tool" -ForegroundColor Gray
    }
    Write-Host ""
    
    # 如果有 Python，直接执行
    if ($tools -contains "Python + sqlite3") {
        Write-Host "🚀 使用 Python 执行迁移..." -ForegroundColor Cyan
        Write-Host ""
        
        $pythonScript = @"
import sqlite3
import os

db_path = r'$dbPath'

print('=' * 60)
print('FastToken OAuth 精简 - 数据库迁移脚本')
print('=' * 60)
print()

# 连接数据库
print('📡 连接数据库...')
conn = sqlite3.connect(db_path)
cursor = conn.cursor()

print('✅ 数据库连接成功')
print()

# 检查要删除的表
print('🔍 检查要删除的表...')
cursor.execute('''
    SELECT name FROM sqlite_master 
    WHERE type='table' 
    AND name IN ('custom_oauth_providers', 'user_oauth_bindings')
''')
existing_tables = [row[0] for row in cursor.fetchall()]

if not existing_tables:
    print('ℹ️  要删除的表不存在，无需迁移')
    print('   - custom_oauth_providers: 不存在')
    print('   - user_oauth_bindings: 不存在')
    conn.close()
    print()
    print('=' * 60)
    print('✅ 迁移完成（无需操作）')
    print('=' * 60)
    exit(0)

print(f'📋 找到 {len(existing_tables)} 个要删除的表:')
for table in existing_tables:
    print(f'   - {table}')
print()

# 执行删除
print('🗑️  开始删除表...')

for table in existing_tables:
    print(f'   删除 {table}...')
    cursor.execute(f'DROP TABLE IF EXISTS {table}')
    print(f'   ✅ {table} 已删除')

conn.commit()
print()
print('✅ 表删除成功！')
print()

# 验证
print('🔍 验证删除结果...')
cursor.execute('''
    SELECT name FROM sqlite_master 
    WHERE type='table' 
    AND name IN ('custom_oauth_providers', 'user_oauth_bindings')
''')
remaining = cursor.fetchall()

if not remaining:
    print('✅ 验证成功：表已被完全删除')
else:
    print(f'⚠️  警告：仍有残留表: {remaining}')

# 显示所有表
print()
print('📊 当前数据库中的所有表:')
cursor.execute('SELECT name FROM sqlite_master WHERE type="table" ORDER BY name')
tables = cursor.fetchall()

for i, table in enumerate(tables, 1):
    print(f'   {i}. {table[0]}')

conn.close()

print()
print('=' * 60)
print('✅ 迁移完成！')
print('=' * 60)
print()
print('📝 后续步骤:')
print('   1. 重启 FastToken 服务')
print('   2. 测试登录功能')
print('   3. 检查用户数据是否正常')
print()
"@
        
        $pythonScript | python
    }
} else {
    Write-Host "⚠️  未找到 SQLite 工具" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "请手动执行以下步骤:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "1. 下载 DB Browser for SQLite:" -ForegroundColor White
    Write-Host "   https://sqlitebrowser.org/dl/" -ForegroundColor Gray
    Write-Host ""
    Write-Host "2. 打开数据库文件:" -ForegroundColor White
    Write-Host "   $dbPath" -ForegroundColor Gray
    Write-Host ""
    Write-Host "3. 执行 SQL 标签，输入:" -ForegroundColor White
    Write-Host "   DROP TABLE IF EXISTS custom_oauth_providers;" -ForegroundColor Gray
    Write-Host "   DROP TABLE IF EXISTS user_oauth_bindings;" -ForegroundColor Gray
    Write-Host ""
    Write-Host "4. 点击 Execute，然后 Write Changes" -ForegroundColor White
    Write-Host ""
}

Write-Host "============================================================" -ForegroundColor Cyan
