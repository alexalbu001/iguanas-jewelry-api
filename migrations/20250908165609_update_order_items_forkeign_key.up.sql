-- Update foreign key constraint to allow product deletion
-- This preserves order history while allowing products to be deleted

-- Drop the existing foreign key constraint
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;

-- Add the new foreign key constraint with CASCADE DELETE
-- This will delete order_items when a product is deleted
-- Alternatively, we could use SET NULL to keep order_items but set product_id to NULL
ALTER TABLE order_items 
ADD CONSTRAINT order_items_product_id_fkey 
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

-- Also update cart_items foreign key for consistency
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;
ALTER TABLE cart_items 
ADD CONSTRAINT cart_items_product_id_fkey 
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;
