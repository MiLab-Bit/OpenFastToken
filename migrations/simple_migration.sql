-- FastToken OAuth 精简 - 简单迁移脚本
-- 数据库: SQLite
-- 执行时间: 2026-05-30

-- 删除自定义 OAuth Provider 配置表
DROP TABLE IF EXISTS custom_oauth_providers;

-- 删除用户 OAuth 绑定关系表
DROP TABLE IF EXISTS user_oauth_bindings;

-- 验证：查看剩余的表
-- SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;
