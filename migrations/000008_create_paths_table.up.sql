CREATE TABLE paths (
    parent_id UUID NOT NULL,
    child_id UUID NOT NULL,
    depth INTEGER NOT NULL,
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES directories(id) ON DELETE CASCADE,
    FOREIGN KEY (child_id) REFERENCES directories(id) ON DELETE CASCADE
);