CREATE TABLE categories (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    parent_id UUID DEFAULT NULL,
    name VARCHAR(255) NOT NULL,
    sequence INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk_categories PRIMARY KEY (id),
    CONSTRAINT fk_categories_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories(id)
);
