CREATE SCHEMA user_s;

CREATE TABLE user_s.users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    username VARCHAR(128) UNIQUE,
    email VARCHAR(256) UNIQUE,
    is_v_email BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number VARCHAR(15) UNIQUE,
    is_v_phone_number BOOLEAN NOT NULL DEFAULT FALSE,
    password TEXT,
    permission_type INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE user_s.user_profile (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    img_path TEXT UNIQUE,
    birthday DATE,
    bio VARCHAR(450)
);
