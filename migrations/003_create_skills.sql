-- FastToken Phase 1b: Agent Marketplace L1（技能注册中心）
-- Date: 2026-07-31
-- 说明：
--   1. created_at/updated_at 使用 bigint epoch 秒，对齐仓库既有约定（Token/Log/Enterprise 全用 int64）。
--   2. sha256 使用 VARCHAR(64) 而非 CHAR(64)：PostgreSQL 下 CHAR 会补空格，破坏等值比较。
--   3. user_id/tenant_id 为 L2「组织私有技能」预留，本期全部写 0（公开技能）。
--      撞 id 铁律：tenant_id 只写 enterprise_id，永不写 user_id。
--   4. 本脚本与 GORM AutoMigrate(&Skill{}) 等价，幂等可重复执行；服务启动会自动建表，
--      此文件供 DBA 审计与手工预建表使用。

BEGIN;

CREATE TABLE IF NOT EXISTS skills (
    id           BIGSERIAL     PRIMARY KEY,
    name         VARCHAR(128)  NOT NULL,
    version      VARCHAR(32)   NOT NULL,
    description  TEXT          NOT NULL DEFAULT '',
    author       VARCHAR(128)  NOT NULL DEFAULT '',
    category     VARCHAR(64)   NOT NULL DEFAULT 'general',
    download_url VARCHAR(512)  NOT NULL,
    sha256       VARCHAR(64)   NOT NULL,
    size_bytes   BIGINT        NOT NULL DEFAULT 0,
    downloads    BIGINT        NOT NULL DEFAULT 0,
    status       VARCHAR(20)   NOT NULL DEFAULT 'draft',
    user_id      INTEGER       NOT NULL DEFAULT 0,
    tenant_id    INTEGER       NOT NULL DEFAULT 0,
    created_at   BIGINT        NOT NULL DEFAULT 0,
    updated_at   BIGINT        NOT NULL DEFAULT 0
);

-- 版本唯一：同名技能同版本只能存在一条（发布去重的数据库级保障）
CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_name_version ON skills (name, version);
CREATE INDEX IF NOT EXISTS idx_skills_category ON skills (category);
CREATE INDEX IF NOT EXISTS idx_skills_status   ON skills (status);
CREATE INDEX IF NOT EXISTS idx_skills_tenant   ON skills (tenant_id);

COMMIT;
