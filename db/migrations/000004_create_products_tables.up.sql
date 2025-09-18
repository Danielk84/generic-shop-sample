CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(256),
    description TEXT,
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE categories (
    id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category TEXT UNIQUE
);

CREATE TABLE products_categories (
    product_id UUID REFERENCES products(id) ON UPDATE CASCADE ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT product_categorie_pkey PRIMARY KEY (product_id, category_id)
);

CREATE TABLE product_images (
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    img_path TEXT NOT NULL UNIQUE
);

CREATE TABLE product_info (
    product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    info JSONB
);
