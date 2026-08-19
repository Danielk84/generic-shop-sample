CREATE INDEX CONCURRENTLY idx_user_id_parent_referrer
ON user_s.comments(parent, referrer, is_active);
