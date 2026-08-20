SET LOCAL lock_timeout = '5s';
ALTER TABLE IF EXISTS product_s.products
    DROP COLUMN IF EXISTS __search;
