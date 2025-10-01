CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    details JSONB,
    price BIGINT NOT NULL DEFAULT 0,
    pub_date TIMESTAMP NOT NULL DEFAULT now(),
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tag TEXT UNIQUE NOT NULL
);

CREATE TABLE products_categories (
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    tag TEXT REFERENCES categories(tag) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT product_categorie_pkey PRIMARY KEY (product_id, tag)
);

CREATE TABLE product_images (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    img_path TEXT NOT NULL UNIQUE
);
