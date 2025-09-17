CREATE UNIQUE INDEX CONCURRENTLY idx_name_is_available_and_active
ON products(name, is_available, is_active);
