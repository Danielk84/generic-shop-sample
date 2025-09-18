CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    pub_date TIMESTAMP DEFAULT NOW()
    parent UUID DEFAULT NULL,
    children_amount INTEGER,
    referrer text,
    body text
);
