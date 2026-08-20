CREATE INDEX CONCURRENTLY idx_products_search
ON product_s.products USING gin (__search);
