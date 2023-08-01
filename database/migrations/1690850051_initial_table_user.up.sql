CREATE TABLE users (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(800) NOT NULL,
    created_at TIME NOT NULL,
    updated_at TIME NULL,
    active BOOLEAN NOT NULL
);