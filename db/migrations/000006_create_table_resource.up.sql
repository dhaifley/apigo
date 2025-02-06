BEGIN;

CREATE SEQUENCE IF NOT EXISTS resource_key_seq;

CREATE TABLE IF NOT EXISTS resource (
    account_id TEXT NOT NULL DEFAULT current_setting('app.account_id')::TEXT,
    FOREIGN KEY (account_id) REFERENCES account (account_id) ON DELETE CASCADE,
    resource_key BIGINT NOT NULL DEFAULT nextval('resource_key_seq') UNIQUE,
    PRIMARY KEY (account_id, resource_key),
    resource_id UUID NOT NULL,
    UNIQUE (account_id, resource_id),
    name TEXT NOT NULL,
    version TEXT,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'new',
    status_data JSONB,
    key_field TEXT NOT NULL,
    key_regex TEXT,
    clear_condition TEXT,
    clear_after BIGINT NOT NULL DEFAULT 60*60*24*30,
    clear_delay BIGINT NOT NULL DEFAULT 0,
    data JSONB,
    source TEXT,
    commit_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    FOREIGN KEY (created_by) REFERENCES "user" (user_key) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT,
    FOREIGN KEY (updated_by) REFERENCES "user" (user_key) ON DELETE SET NULL
);

ALTER TABLE IF EXISTS resource ENABLE ROW LEVEL SECURITY;

CREATE POLICY account_isolation_policy ON resource
    USING (account_id = current_setting('app.account_id')::TEXT);

COMMIT;
