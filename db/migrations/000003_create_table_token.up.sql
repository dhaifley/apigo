BEGIN;

CREATE TABLE IF NOT EXISTS token (
    account_id TEXT NOT NULL DEFAULT current_setting('app.account_id')::TEXT,
    FOREIGN KEY (account_id) REFERENCES account (account_id) ON DELETE CASCADE,
    token_id TEXT NOT NULL,
    PRIMARY KEY (account_id, token_id),
    status TEXT NOT NULL DEFAULT 'active',
    expiration TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    FOREIGN KEY (created_by) REFERENCES "user" (user_key) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT,
    FOREIGN KEY (updated_by) REFERENCES "user" (user_key) ON DELETE SET NULL
);

ALTER TABLE IF EXISTS token ENABLE ROW LEVEL SECURITY;

CREATE POLICY account_isolation_policy ON token
    USING (account_id = current_setting('app.account_id')::TEXT);

COMMIT;
