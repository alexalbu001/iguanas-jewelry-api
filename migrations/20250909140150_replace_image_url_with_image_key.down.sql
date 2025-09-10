ALTER TABLE product_images ADD COLUMN image_url VARCHAR(255) NOT NULL;
ALTER TABLE product_images DROP COLUMN image_key;