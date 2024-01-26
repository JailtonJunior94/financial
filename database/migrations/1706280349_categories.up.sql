CREATE TABLE categories (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sequence INT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NULL,
    active BOOLEAN NOT NULL,

    foreign key (user_id) references users(id)
);