ALTER TABLE product_images DROP COLUMN image_url;
ALTER TABLE product_images ADD COLUMN image_key TEXT NOT NULL;