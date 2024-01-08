CREATE TABLE user_tokens (
    token_id VARCHAR(255) PRIMARY KEY,
    token_name VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    last_used TIMESTAMPTZ,
    user_id VARCHAR(255) REFERENCES users(id)
);