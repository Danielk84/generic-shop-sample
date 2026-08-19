CREATE INDEX CONCURRENTLY idx_vendors_order
ON order_s.vendors_order(user_id, order_id, product_id, property);
