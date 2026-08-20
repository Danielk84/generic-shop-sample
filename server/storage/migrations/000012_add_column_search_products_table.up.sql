-- add full search options
SET LOCAL lock_timeout = '5s';
ALTER TABLE IF EXISTS product_s.products
    ADD COLUMN IF NOT EXISTS __search TSVECTOR;
