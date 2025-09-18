CREATE INDEX CONCURRENTLY idx_user_id_parent_referrer
ON comments(user_id, parent, referrer);
