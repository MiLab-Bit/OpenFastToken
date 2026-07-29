-- FastToken 语义缓存表迁移脚本
-- 用于支持向量相似度搜索

-- 1. 启用 pgvector 扩展（如果尚未启用）
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. 创建语义缓存表
CREATE TABLE IF NOT EXISTS semantic_cache_entries (
    id SERIAL PRIMARY KEY,
    model_name VARCHAR(255) NOT NULL,
    prompt TEXT NOT NULL,
    prompt_vector vector(1536) NOT NULL,  -- text-embedding-3-small 维度
    request_body JSONB NOT NULL,
    response_body JSONB NOT NULL,
    created_at BIGINT NOT NULL,
    expires_at BIGINT,
    user_group VARCHAR(50),
    ttl BIGINT
);

-- 3. 创建向量索引（性能优化）
-- 方案A：HNSW索引（推荐，性能更好）
-- 需要约 100MB 内存用于构建，适合 >10万条记录
CREATE INDEX IF NOT EXISTS idx_prompt_vector_hnsw
ON semantic_cache_entries
USING hnsw (prompt_vector vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- 方案B：IVFFlat索引（备选，兼容性更好）
-- 适合 10万条以下记录
-- CREATE INDEX IF NOT EXISTS idx_prompt_vector_ivfflat
-- ON semantic_cache_entries
-- USING ivfflat (prompt_vector vector_cosine_ops)
-- WITH (lists = 100);

-- 4. 创建其他索引
CREATE INDEX IF NOT EXISTS idx_model_created
ON semantic_cache_entries(model_name, created_at);

CREATE INDEX IF NOT EXISTS idx_expires
ON semantic_cache_entries(expires_at)
WHERE expires_at > 0;

CREATE INDEX IF NOT EXISTS idx_user_group
ON semantic_cache_entries(user_group);

-- 5. 添加注释
COMMENT ON TABLE semantic_cache_entries IS '语义缓存表，用于存储AI请求的向量化和响应';
COMMENT ON COLUMN semantic_cache_entries.prompt_vector IS '提示文本的向量表示（1536维）';
COMMENT ON COLUMN semantic_cache_entries.expires_at IS '过期时间戳（Unix），0表示永不过期';

-- 6. 授权（根据实际用户名调整）
-- GRANT SELECT, INSERT, UPDATE, DELETE ON semantic_cache_entries TO fasttoken_user;

-- 验证迁移
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'semantic_cache_entries';

-- 显示表结构
\d semantic_cache_entries
