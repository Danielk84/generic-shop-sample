CREATE UNIQUE INDEX CONCURRENTLY idx_name_is_available_and_active
ON product_s.products(id, is_available, is_active);
