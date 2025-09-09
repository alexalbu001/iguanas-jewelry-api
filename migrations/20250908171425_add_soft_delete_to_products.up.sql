-- Add soft delete support to products table
-- This allows products to be marked as deleted without removing them from the database
-- This preserves order history and maintains data integrity

ALTER TABLE products ADD COLUMN deleted_at TIMESTAMP NULL;

-- Create an index on deleted_at for better query performance
CREATE INDEX idx_products_deleted_at ON products(deleted_at);

-- Add a comment to document the soft delete pattern
COMMENT ON COLUMN products.deleted_at IS 'Timestamp when the product was soft deleted. NULL means the product is active.';
