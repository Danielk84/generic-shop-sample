CREATE INDEX CONCURRENTLY idx_user_id_parent_referrer
ON comments(username, parent, referrer, is_active);
