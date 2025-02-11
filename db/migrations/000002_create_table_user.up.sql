BEGIN;

CREATE SEQUENCE IF NOT EXISTS user_key_seq;

CREATE TABLE IF NOT EXISTS "user" (
    user_key BIGINT NOT NULL DEFAULT nextval('user_key_seq') PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    password TEXT,
    email TEXT,
    last_name TEXT,
    first_name TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT
);

COMMIT;
