CREATE TABLE users (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(128) UNIQUE NOT NULL,
    email VARCHAR(256) UNIQUE,
    password TEXT NOT NULL,
    permission_type INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE user_profile (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    img_path TEXT,
    age INTEGER CHECK (age > 0),
    phone_number TEXT,
    bio VARCHAR(450),
    UNIQUE (img_path, phone_number)
);
