CREATE INDEX CONCURRENTLY idx_orders_user_id
ON orders(user_id, started_at, is_paid, is_delivered);
