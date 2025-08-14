CREATE TABLE IF NOT EXISTS payment (
    id VARCHAR(50) PRIMARY KEY,
    order_id VARCHAR(50) REFERENCES orders(id),
    stripe_payment_id VARCHAR(50),
    amount INTEGER NOT NULL CHECK (amount > 0),
    status VARCHAR(50) NOT NULL,
    failure_reason TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);