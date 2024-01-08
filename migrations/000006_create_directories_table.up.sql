CREATE TABLE directories (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id UUID NULL, -- NULL indicates a root directory.
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    last_write TIMESTAMPTZ NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES directories(id) ON DELETE CASCADE 
);