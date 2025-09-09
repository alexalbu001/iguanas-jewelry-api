-- Revert CASCADE foreign key constraints back to RESTRICT
-- This prevents automatic deletion of order items when products are soft-deleted
-- With soft delete, we want to preserve order history even when products are "deleted"

-- Drop the CASCADE foreign key constraint on order_items
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;

-- Re-add the foreign key constraint with default (RESTRICT) behavior
ALTER TABLE order_items
ADD CONSTRAINT order_items_product_id_fkey
FOREIGN KEY (product_id) REFERENCES products(id);

-- Drop the CASCADE foreign key constraint on cart_items
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;

-- Re-add the foreign key constraint with default (RESTRICT) behavior
ALTER TABLE cart_items
ADD CONSTRAINT cart_items_product_id_fkey
FOREIGN KEY (product_id) REFERENCES products(id);
