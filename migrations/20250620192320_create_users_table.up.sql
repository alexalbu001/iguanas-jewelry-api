CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,  
    googleid VARCHAR(255) UNIQUE NOT NULL,  -- UNIQUE creates an index
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'customer',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);