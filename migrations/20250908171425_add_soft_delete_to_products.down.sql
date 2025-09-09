-- Remove soft delete support from products table
-- This reverts the soft delete functionality

DROP INDEX IF EXISTS idx_products_deleted_at;
ALTER TABLE products DROP COLUMN IF EXISTS deleted_at;
