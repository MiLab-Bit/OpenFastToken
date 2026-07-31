-- 数据库迁移: 重命名 users 表的 aff_history 列为 aff_history_quota
-- 日期: 2026-06-30
-- 说明: AffHistoryQuota 字段的数据库列名应该从 aff_history 改为 aff_history_quota 以保持一致性

-- 检查列名并重命名
DO $$
BEGIN
    -- 检查 aff_history 列是否存在
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'users' 
        AND column_name = 'aff_history'
    ) THEN
        -- 重命名列
        ALTER TABLE users RENAME COLUMN aff_history TO aff_history_quota;
        RAISE NOTICE 'Column aff_history renamed to aff_history_quota';
    ELSE
        RAISE NOTICE 'Column aff_history does not exist, skipping';
    END IF;
    
    -- 检查 aff_history_quota 列是否已存在（避免重复）
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'users' 
        AND column_name = 'aff_history_quota'
    ) THEN
        -- 如果重命名失败，添加新列
        ALTER TABLE users ADD COLUMN aff_history_quota int DEFAULT 0;
        RAISE NOTICE 'Column aff_history_quota added';
    END IF;
END $$;

-- 验证修改
SELECT column_name, data_type, column_default 
FROM information_schema.columns 
WHERE table_name = 'users' 
AND column_name LIKE 'aff_history%';
