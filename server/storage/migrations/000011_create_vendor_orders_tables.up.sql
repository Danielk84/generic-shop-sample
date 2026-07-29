CREATE TABLE order_s.vendors_order (
    user_id UUID NOT NULL REFERENCES user_s.users(id) ON DELETE RESTRICT,
    order_id UUID NOT NULL REFERENCES order_s.orders(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES product_s.products(id) ON DELETE RESTRICT,
    property JSONB NOT NULL CHECK(jsonb_typeof(property) = 'object') DEFAULT '{}'::JSONB,
    quantity INTEGER NOT NULL CHECK(quantity > 0) DEFAULT 1,
    total_bill BIGINT NOT NULL CHECK(total_bill >= 0) DEFAULT 0,
    is_delivered BOOLEAN NOT NULL DEFAULT FALSE
);
