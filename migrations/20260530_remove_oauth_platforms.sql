-- FastToken OAuth 精简数据库迁移脚本
-- 执行时间: 2026-05-30
-- 说明: 删除不需要的 OAuth 平台相关字段和表

-- ============================================
-- 1. 删除自定义 OAuth Provider 相关表
-- ============================================

-- 删除自定义 OAuth Provider 配置表
DROP TABLE IF EXISTS `custom_oauth_providers`;

-- 删除用户 OAuth 绑定表
DROP TABLE IF EXISTS `user_oauth_bindings`;

-- ============================================
-- 2. 删除 User 表中不需要的 OAuth 字段
-- ============================================

-- 删除 Discord 相关字段
ALTER TABLE `users` DROP COLUMN IF EXISTS `discord_id`;

-- 删除 OIDC 相关字段
ALTER TABLE `users` DROP COLUMN IF EXISTS `oidc_id`;

-- 删除 Telegram 相关字段
ALTER TABLE `users` DROP COLUMN IF EXISTS `telegram_id`;

-- 删除 LinuxDO 相关字段
ALTER TABLE `users` DROP COLUMN IF EXISTS `linux_do_id`;

-- ============================================
-- 3. 保留的字段（不要删除）
-- ============================================

-- 保留 github_id (GitHub OAuth)
-- 保留 wechat_id (微信登录)
-- 保留 email (邮箱注册)
-- 保留 phone (手机号注册)

-- ============================================
-- 4. 验证迁移结果
-- ============================================

-- 查看用户表结构
DESCRIBE `users`;

-- 检查是否还有残留的 OAuth 绑定数据
SELECT COUNT(*) FROM `users` WHERE `github_id` IS NOT NULL OR `github_id` != '';
SELECT COUNT(*) FROM `users` WHERE `wechat_id` IS NOT NULL OR `wechat_id` != '';

-- ============================================
-- 5. 清理说明
-- ============================================

-- 此脚本会删除以下内容：
-- 1. custom_oauth_providers 表 - 自定义 OAuth Provider 配置
-- 2. user_oauth_bindings 表 - 用户 OAuth 绑定关系
-- 3. users 表中的 discord_id 字段
-- 4. users 表中的 oidc_id 字段
-- 5. users 表中的 telegram_id 字段
-- 6. users 表中的 linux_do_id 字段

-- 迁移完成后，系统将只支持：
-- 1. 邮箱 + 密码登录
-- 2. 手机号注册（如果已实现）
-- 3. 微信扫码登录
-- 4. GitHub OAuth 登录

-- ============================================
-- 6. 备份建议（执行前请先备份）
-- ============================================

-- 备份整个数据库
-- mysqldump -u root -p fasttoken > fasttoken_backup_20260530.sql

-- 或者只备份受影响的表
-- mysqldump -u root -p fasttoken users custom_oauth_providers user_oauth_bindings > oauth_backup_20260530.sql

-- ============================================
-- 7. PostgreSQL 版本（如使用 PostgreSQL）
-- ============================================

-- DROP TABLE IF EXISTS custom_oauth_providers;
-- DROP TABLE IF EXISTS user_oauth_bindings;
-- ALTER TABLE users DROP COLUMN IF EXISTS discord_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS oidc_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS telegram_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS linux_do_id;

-- ============================================
-- 8. SQLite 版本（如使用 SQLite）
-- ============================================

-- DROP TABLE IF EXISTS custom_oauth_providers;
-- DROP TABLE IF EXISTS user_oauth_bindings;
-- SQLite 不支持 DROP COLUMN，需要重建表：
-- CREATE TABLE users_new AS SELECT id, username, password, display_name, role, status, email, github_id, wechat_id, ... FROM users;
-- DROP TABLE users;
-- ALTER TABLE users_new RENAME TO users;
