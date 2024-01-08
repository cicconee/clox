CREATE TABLE files (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    directory_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (directory_id) REFERENCES directories(id) ON DELETE CASCADE
);