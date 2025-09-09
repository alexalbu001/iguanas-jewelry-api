-- Revert back to CASCADE constraints (if needed)
-- This would restore the CASCADE behavior for foreign key constraints

-- Drop the RESTRICT foreign key constraint on order_items
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;

-- Add the CASCADE foreign key constraint
ALTER TABLE order_items
ADD CONSTRAINT order_items_product_id_fkey
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

-- Drop the RESTRICT foreign key constraint on cart_items
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;

-- Add the CASCADE foreign key constraint
ALTER TABLE cart_items
ADD CONSTRAINT cart_items_product_id_fkey
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;
