CREATE UNIQUE INDEX CONCURRENTLY idx_user_info_users
ON user_s.users(username, email, is_active);
