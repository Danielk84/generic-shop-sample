CREATE UNIQUE INDEX CONCURRENTLY idx_user_info_users
ON users(username, email, is_active);
