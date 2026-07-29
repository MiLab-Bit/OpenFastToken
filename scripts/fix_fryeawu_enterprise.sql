-- ============================================================
-- 一次性脚本：补写 FryeaWu 企业认证回写
-- 背景：旧代码审批通过时只改了 enterprise 表，未回写 users 表，
--       导致用户个人资料无变化、未关联企业、企业版功能未开启。
-- 目标：user.id=1 (FreyaID00 / abovetigers@gmail.com)
--       <-> enterprise.id=1 (上海咫策科技有限公司, contact_name=FryeaWu)
-- 规则：写入企业关联；会员等级取 max(现有, 企业授予)，不降级。
-- 运行：psql -h localhost -U postgres -d fasttoken -f fix_fryeawu_enterprise.sql
-- ============================================================

BEGIN;

-- 0) 安全校验：确认目标记录存在且匹配，避免误改其它账号
DO $$
DECLARE
  v_user_ok boolean;
  v_ent_ok  boolean;
BEGIN
  SELECT EXISTS(SELECT 1 FROM users WHERE id = 1 AND email ILIKE 'abovetigers@gmail.com')
    INTO v_user_ok;
  SELECT EXISTS(SELECT 1 FROM enterprise WHERE id = 1 AND contact_name = 'FryeaWu')
    INTO v_ent_ok;
  IF NOT v_user_ok OR NOT v_ent_ok THEN
    RAISE EXCEPTION '目标用户/企业不匹配，终止脚本。user_ok=%, ent_ok=%', v_user_ok, v_ent_ok;
  END IF;
END $$;

-- 1) 企业记录回写提交者 user_id（保持前后一致，便于后续逻辑）
UPDATE enterprise
SET user_id = 1,
    updated_at = EXTRACT(EPOCH FROM NOW())::bigint
WHERE id = 1 AND (user_id = 0 OR user_id IS NULL);

-- 2) 用户回写 enterprise_id + 会员等级（取较高者，不降级）
--    注意：users 表无 updated_at 列，仅更新业务字段。
UPDATE users u
SET enterprise_id = 1,
    membership_level = CASE
      WHEN (CASE u.membership_level WHEN 'platinum' THEN 3 WHEN 'gold' THEN 2 WHEN 'silver' THEN 1 ELSE 0 END)
           >= (SELECT CASE e.membership_level WHEN 'platinum' THEN 3 WHEN 'gold' THEN 2 WHEN 'silver' THEN 1 ELSE 0 END
                 FROM enterprise e WHERE e.id = 1)
      THEN u.membership_level
      ELSE (SELECT e.membership_level FROM enterprise e WHERE e.id = 1)
    END
WHERE u.id = 1;

-- 3) 建立企业-用户关联（管理员 / active），已存在则跳过
INSERT INTO enterprise_user (enterprise_id, user_id, role, status, joined_at, created_at, updated_at)
SELECT 1, 1, 'admin', 'active',
       EXTRACT(EPOCH FROM NOW())::bigint,
       EXTRACT(EPOCH FROM NOW())::bigint,
       EXTRACT(EPOCH FROM NOW())::bigint
WHERE NOT EXISTS (
  SELECT 1 FROM enterprise_user WHERE enterprise_id = 1 AND user_id = 1
);

-- 4) 校验结果回显
SELECT
  u.id            AS user_id,
  u.username,
  u.membership_level,
  u.enterprise_id,
  e.id            AS enterprise_id,
  e.name          AS enterprise_name,
  e.membership_level AS granted_level,
  (SELECT COUNT(*) FROM enterprise_user eu WHERE eu.enterprise_id = 1 AND eu.user_id = 1) AS ent_user_link
FROM users u, enterprise e
WHERE u.id = 1 AND e.id = 1;

COMMIT;
