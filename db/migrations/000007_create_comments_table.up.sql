CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(128) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    pub_date TIMESTAMP NOT NULL DEFAULT now(),
    parent UUID DEFAULT NULL,
    children_amount INTEGER NOT NULL DEFAULT 0,
    referrer TEXT NOT NULL,
    body TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);
