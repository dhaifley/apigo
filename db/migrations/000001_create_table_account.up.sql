BEGIN;

CREATE TABLE IF NOT EXISTS account (
    account_id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    status_data JSONB,
    repo TEXT,
    repo_status TEXT NOT NULL DEFAULT 'inactive',
    repo_status_data JSONB,
    secret TEXT NOT NULL DEFAULT gen_random_uuid(),
    data JSONB,
    resource_commit_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE IF EXISTS account ENABLE ROW LEVEL SECURITY;

CREATE POLICY account_isolation_policy ON account
USING (current_setting('app.account_id')::TEXT = 'sys' OR 
    account_id = current_setting('app.account_id')::TEXT);

COMMIT;
