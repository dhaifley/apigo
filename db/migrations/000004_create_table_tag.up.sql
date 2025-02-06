BEGIN;

CREATE TABLE IF NOT EXISTS tag (
    account_id TEXT NOT NULL DEFAULT current_setting('app.account_id')::TEXT,
    FOREIGN KEY (account_id) REFERENCES account (account_id) ON DELETE CASCADE,
    tag_key TEXT NOT NULL,
    tag_val TEXT,
    PRIMARY KEY (account_id, tag_key, tag_val),
    status TEXT NOT NULL DEFAULT 'active',
    status_data JSONB,
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    FOREIGN KEY (created_by) REFERENCES "user" (user_key) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT,
    FOREIGN KEY (updated_by) REFERENCES "user" (user_key) ON DELETE SET NULL
);

ALTER TABLE IF EXISTS tag ENABLE ROW LEVEL SECURITY;

CREATE POLICY account_isolation_policy ON tag
    USING (account_id = current_setting('app.account_id')::TEXT);

COMMIT;
