UPDATE cart_items SET quantity = 1 WHERE quantity IS NULL;
ALTER TABLE cart_items ALTER COLUMN quantity SET NOT NULL;
