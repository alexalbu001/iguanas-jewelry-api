CREATE TABLE IF NOT EXISTS user_favorites (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(id),
    product_id VARCHAR(50) REFERENCES products(id),
    created_at TIMESTAMP NOT NULL,
    UNIQUE (user_id, product_id)
);