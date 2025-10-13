CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    started_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),

    items_total INTEGER NOT NULL CHECK (items_total >= 0) DEFAULT 0,
    total_bill BIGINT NOT NULL CHECK (total_bill > 0) DEFAULT 0,

    payment_summary TEXT,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,

    address TEXT,
    zip_code TEXT,
    is_confirmed BOOLEAN NOT NULL DEFAULT FALSE,

    delivery_details JSONB,
    is_delivered BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE order_items (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    items_total INTEGER NOT NULL CHECK (items_total > 0) DEFAULT 1,
    price BIGINT NOT NULL DEFAULT 0,
    is_packed BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, order_id, product_id)
);

CREATE OR REPLACE FUNCTION update_order_after_add_order_items() RETURNS TRIGGER
AS $$
    DECLARE
        product_price BIGINT;
    BEGIN
        SELECT price INTO product_price FROM products WHERE id = NEW.product_id AND price = NEW.price LIMIT 1;
        IF product_price != NEW.price THEN
            RAISE EXCEPTION 'invalid price';
        END IF;

        UPDATE products SET available_quantity = available_quantity - 1 WHERE id = NEW.product_id;

        UPDATE orders SET items_total = items_total + 1, total_bill = total_bill + product_price
            WHERE id = NEW.order_id;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER update_order_after_add_order_items_trigger
    AFTER INSERT ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_order_after_add_order_items();

CREATE OR REPLACE FUNCTION update_order_after_update_order_items() RETURNS TRIGGER
AS $$
    BEGIN
        UPDATE products SET available_quantity = (available_quantity + OLD.items_total) - NEW.items_total
            WHERE id = NEW.product_id;

        UPDATE orders
            SET
                items_total = (items_total - OLD.items_total) + NEW.items_total,
                total_bill = (total_bill - (OLD.items_total * OLD.price)) + (NEW.items_total * NEW.price)
            WHERE id = NEW.order_id;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER update_order_after_update_order_items_trigger
    AFTER UPDATE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_order_after_update_order_items();

CREATE OR REPLACE FUNCTION update_order_after_delete_order_items() RETURNS TRIGGER
AS $$
    BEGIN
        UPDATE orders SET items_total = items_total - 1, total_bill = total_bill - OLD.price
            WHERE id = OLD.order_id;

        UPDATE products SET available_quantity = available_quantity + 1 WHERE id = OLD.product_id;
        RETURN OLD;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER update_order_after_delete_order_items_trigger
    AFTER DELETE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_order_after_delete_order_items();
