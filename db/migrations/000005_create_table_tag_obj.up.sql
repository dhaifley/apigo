BEGIN;

CREATE SEQUENCE IF NOT EXISTS tag_obj_key_seq;

CREATE TABLE IF NOT EXISTS tag_obj (
    account_id TEXT NOT NULL DEFAULT current_setting('app.account_id')::TEXT,
    FOREIGN KEY (account_id) REFERENCES account (account_id) ON DELETE CASCADE,
    tag_obj_key BIGINT NOT NULL
        DEFAULT nextval('tag_obj_key_seq') UNIQUE,
    PRIMARY KEY (account_id, tag_obj_key),
    tag_type TEXT NOT NULL,
    tag_obj_id TEXT NOT NULL,
    tag_key TEXT NOT NULL,
    tag_val TEXT,
    UNIQUE (account_id, tag_type, tag_obj_id, tag_key, tag_val),
    FOREIGN KEY (account_id, tag_key, tag_val)
        REFERENCES tag (account_id, tag_key, tag_val) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    FOREIGN KEY (created_by) REFERENCES "user" (user_key) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT,
    FOREIGN KEY (updated_by) REFERENCES "user" (user_key) ON DELETE SET NULL
);

ALTER TABLE IF EXISTS tag_obj ENABLE ROW LEVEL SECURITY;

CREATE POLICY account_isolation_policy ON tag_obj
    USING (account_id = current_setting('app.account_id')::TEXT);

COMMIT;
