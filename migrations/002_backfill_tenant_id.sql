-- FastToken Phase 1: tenant_id backfill (data only)
-- Date: 2026-07-31
-- Strategy (collision-safe):
--   * Enterprise members  -> tenant_id = enterprise_id (shared org tenant)
--   * Solo users           -> tenant_id = 0 (personal, scoped by user_id in middleware)
--   * We NEVER put a user_id into tenant_id, so no id collision with enterprise_id (2,3).

BEGIN;

-- Step 1: enterprise members' resources get their enterprise_id
UPDATE tokens
SET tenant_id = eu.enterprise_id
FROM enterprise_user eu
WHERE eu.user_id = tokens.user_id
  AND tokens.tenant_id IS DISTINCT FROM eu.enterprise_id;

UPDATE logs
SET tenant_id = eu.enterprise_id
FROM enterprise_user eu
WHERE eu.user_id = logs.user_id
  AND logs.tenant_id IS DISTINCT FROM eu.enterprise_id;

UPDATE top_ups
SET tenant_id = eu.enterprise_id
FROM enterprise_user eu
WHERE eu.user_id = top_ups.user_id
  AND top_ups.tenant_id IS DISTINCT FROM eu.enterprise_id;

-- Step 2: solo users' resources -> tenant_id = 0 (platform default / personal)
UPDATE tokens  SET tenant_id = 0 WHERE tenant_id IS NULL;
UPDATE logs    SET tenant_id = 0 WHERE tenant_id IS NULL;
UPDATE top_ups SET tenant_id = 0 WHERE tenant_id IS NULL;
UPDATE channels SET tenant_id = 0 WHERE tenant_id IS NULL;

COMMIT;

-- Verification (run separately)
-- SELECT 'tokens', count(*) FILTER (WHERE tenant_id IS NULL) FROM tokens
-- UNION ALL SELECT 'logs', count(*) FILTER (WHERE tenant_id IS NULL) FROM logs
-- UNION ALL SELECT 'top_ups', count(*) FILTER (WHERE tenant_id IS NULL) FROM top_ups
-- UNION ALL SELECT 'channels', count(*) FILTER (WHERE tenant_id IS NULL) FROM channels;
