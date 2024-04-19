CREATE TABLE users (
    id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(800) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT uk_users_email UNIQUE (email)
);

CREATE TABLE categories (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    sequence INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk_categories PRIMARY KEY (id),
    CONSTRAINT fk_categories_users FOREIGN KEY (user_id) REFERENCES users(id)
);