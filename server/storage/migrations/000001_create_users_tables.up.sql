CREATE SCHEMA user_s;

CREATE TABLE user_s.users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    email VARCHAR(256) UNIQUE,
    is_v_email BOOLEAN NOT NULL DEFAULT FALSE,

    phone_number VARCHAR(15) UNIQUE,
    is_v_phone_number BOOLEAN NOT NULL DEFAULT FALSE,

    permission_type INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,

    national_code VARCHAR(10) NOT NULL DEFAULT '',
    first_name VARCHAR(50) NOT NULL DEFAULT '',
    last_name VARCHAR(60) NOT NULL DEFAULT '',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE user_s.shop (
    user_id UUID PRIMARY KEY REFERENCES user_s.users(id) ON DELETE CASCADE,

    brand VARCHAR(100) UNIQUE,
    shop_addr VARCHAR(1000) NOT NULL,
    zip_code VARCHAR(10) UNIQUE,
    business_code VARCHAR(100) UNIQUE NOT NULL,

    phone_number VARCHAR(15) UNIQUE,
    is_v_phone_number BOOLEAN NOT NULL DEFAULT FALSE,

    bio VARCHAR(650) NOT NULL DEFAULT '',
    img_path TEXT NOT NULL DEFAULT '',

    is_shop BOOLEAN NOT NULL DEFAULT FALSE
);
