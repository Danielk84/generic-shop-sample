CREATE UNIQUE INDEX CONCURRENTLY idx_user_info_users
ON user_s.users(email, permission_type, is_active);
