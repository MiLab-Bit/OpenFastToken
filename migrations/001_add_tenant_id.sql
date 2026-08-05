-- FastToken Phase 1: Multi-Tenancy (tenant_id migration) v2
-- Date: 2026-07-31

BEGIN;

-- Step 1: Add tenant_id to tokens table
ALTER TABLE tokens ADD COLUMN IF NOT EXISTS tenant_id bigint;
CREATE INDEX IF NOT EXISTS idx_tokens_tenant_id ON tokens(tenant_id);

-- Step 2: Add tenant_id to channels table
ALTER TABLE channels ADD COLUMN IF NOT EXISTS tenant_id bigint;
CREATE INDEX IF NOT EXISTS idx_channels_tenant_id ON channels(tenant_id);

-- Step 3: Add tenant_id to logs table
ALTER TABLE logs ADD COLUMN IF NOT EXISTS tenant_id bigint;
CREATE INDEX IF NOT EXISTS idx_logs_tenant_id ON logs(tenant_id);

-- Step 4: Add tenant_id to top_ups table
ALTER TABLE top_ups ADD COLUMN IF NOT EXISTS tenant_id bigint;
CREATE INDEX IF NOT EXISTS idx_top_ups_tenant_id ON top_ups(tenant_id);

-- Step 5: Migrate existing data from enterprise_user mapping
-- tokens: update where user belongs to an enterprise
UPDATE tokens SET tenant_id = eu.enterprise_id
FROM enterprise_user eu
WHERE eu.user_id = tokens.user_id AND tokens.tenant_id IS NULL;

-- channels: currently global/shared, keep as NULL
-- New channels should be explicitly assigned to a tenant.

-- logs: update where user belongs to an enterprise
UPDATE logs SET tenant_id = eu.enterprise_id
FROM enterprise_user eu
WHERE eu.user_id = logs.user_id AND logs.tenant_id IS NULL;

-- top_ups: update where user belongs to an enterprise
UPDATE top_ups SET tenant_id = eu.enterprise_id
FROM enterprise_user eu
WHERE eu.user_id = top_ups.user_id AND top_ups.tenant_id IS NULL;

COMMIT;
