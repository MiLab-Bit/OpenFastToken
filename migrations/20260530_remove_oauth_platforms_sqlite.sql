-- FastToken OAuth 精简 - SQLite 数据库迁移脚本
-- 执行时间: 2026-05-30
-- 数据库类型: SQLite
-- 说明: 删除不需要的 OAuth 平台相关字段和表

-- ============================================
-- 1. 删除自定义 OAuth Provider 相关表
-- ============================================

-- 删除自定义 OAuth Provider 配置表
DROP TABLE IF EXISTS custom_oauth_providers;

-- 删除用户 OAuth 绑定表
DROP TABLE IF EXISTS user_oauth_bindings;

-- ============================================
-- 2. SQLite 字段删除说明
-- ============================================

-- SQLite 不支持 DROP COLUMN 语句
-- 但好消息是：这些字段不影响程序运行！
-- 
-- 原因：
-- 1. Go 代码已经删除了这些字段的引用
-- 2. GORM 会自动忽略代码中不存在的字段
-- 3. 旧字段只是保留在数据库中，不会被访问
--
-- 影响：
-- - 程序运行：无影响
-- - 数据安全：废弃字段的数据不会被读取或更新
-- - 存储空间：字段占用的空间可以忽略不计
--
-- 如果确实需要删除字段，可以：
-- 1. 导出数据
-- 2. 重建表（不包含废弃字段）
-- 3. 导入数据
-- 但这通常不是必须的

-- ============================================
-- 3. 可选：完整重建 users 表（如需彻底清理）
-- ============================================

-- 注意：执行前请先备份数据库！
-- 以下步骤会重建 users 表，删除废弃字段

-- 步骤 1: 创建新表（不包含废弃字段）
-- CREATE TABLE users_new (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     username TEXT UNIQUE NOT NULL,
--     password TEXT NOT NULL,
--     display_name TEXT,
--     role INTEGER DEFAULT 1,
--     status INTEGER DEFAULT 1,
--     email TEXT,
--     github_id TEXT,
--     wechat_id TEXT,
--     -- 保留其他必要字段...
--     created_at INTEGER,
--     last_login_at INTEGER
--     -- 根据实际表结构调整
-- );

-- 步骤 2: 从旧表复制数据到新表
-- INSERT INTO users_new 
-- SELECT id, username, password, display_name, role, status, email, 
--        github_id, wechat_id, created_at, last_login_at
-- FROM users;

-- 步骤 3: 删除旧表
-- DROP TABLE users;

-- 步骤 4: 重命名新表
-- ALTER TABLE users_new RENAME TO users;

-- 步骤 5: 重建索引
-- CREATE INDEX idx_users_username ON users(username);
-- CREATE INDEX idx_users_email ON users(email);
-- CREATE INDEX idx_users_github_id ON users(github_id);
-- CREATE INDEX idx_users_wechat_id ON users(wechat_id);

-- ============================================
-- 4. 验证迁移结果
-- ============================================

-- 查看所有表
SELECT name FROM sqlite_master WHERE type='table';

-- 查看 users 表结构
PRAGMA table_info(users);

-- 检查保留的 OAuth 字段数据
SELECT COUNT(*) as github_count FROM users WHERE github_id IS NOT NULL AND github_id != '';
SELECT COUNT(*) as wechat_count FROM users WHERE wechat_id IS NOT NULL AND wechat_id != '';

-- ============================================
-- 5. 备份建议
-- ============================================

-- Windows PowerShell 备份命令:
-- Copy-Item "C:\Users\Administrator\WorkBuddy\2026-05-27-05-31-58\one-api.db" "C:\Users\Administrator\WorkBuddy\2026-05-27-05-31-58\one-api-backup-20260530.db"

-- 或使用 SQLite 命令行工具备份:
-- sqlite3 one-api.db ".backup one-api-backup-20260530.db"

-- ============================================
-- 6. 执行步骤（推荐）
-- ============================================

-- 推荐方案：只删除自定义 OAuth 表，保留 users 表结构
-- 原因：
-- 1. 删除表是安全的，不会影响现有用户
-- 2. users 表的废弃字段不影响程序运行
-- 3. 避免数据迁移风险

-- 只需执行这两行：
-- DROP TABLE IF EXISTS custom_oauth_providers;
-- DROP TABLE IF EXISTS user_oauth_bindings;

-- ============================================
-- 7. 执行命令
-- ============================================

-- 方式 1: 使用 sqlite3 命令行工具
-- sqlite3 "C:\Users\Administrator\WorkBuddy\2026-05-27-05-31-58\one-api.db" "DROP TABLE IF EXISTS custom_oauth_providers; DROP TABLE IF EXISTS user_oauth_bindings;"

-- 方式 2: 使用 PowerShell 调用 SQLite
-- $dbPath = "C:\Users\Administrator\WorkBuddy\2026-05-27-05-31-58\one-api.db"
-- sqlite3 $dbPath "DROP TABLE IF EXISTS custom_oauth_providers; DROP TABLE IF EXISTS user_oauth_bindings;"
