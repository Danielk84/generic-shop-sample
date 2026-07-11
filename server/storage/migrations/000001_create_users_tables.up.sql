CREATE TABLE users (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(128) UNIQUE,
    email VARCHAR(256) UNIQUE,
    is_v_email BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number VARCHAR(15) UNIQUE,
    is_v_phone_number BOOLEAN NOT NULL DEFAULT FALSE,
    password TEXT,
    permission_type INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE user_profile (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    img_path TEXT UNIQUE,
    birthday DATE,
    bio VARCHAR(450)
);
