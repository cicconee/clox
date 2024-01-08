CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    picture_url VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255),
    register_status VARCHAR(255) NOT NULL
);