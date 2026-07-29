CREATE SCHEMA order_s;

CREATE TABLE order_s.orders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES user_s.users(id) ON DELETE RESTRICT,
    started_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),

    items_total INTEGER NOT NULL CHECK (items_total >= 0) DEFAULT 0,
    total_bill BIGINT NOT NULL CHECK (total_bill >= 0) DEFAULT 0,

    payment_summary JSONB NOT NULL CHECK (jsonb_typeof(payment_summary) = 'array') DEFAULT '[]'::JSONB,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,

    address TEXT,
    zip_code TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,

    is_confirmed BOOLEAN NOT NULL DEFAULT FALSE,

    delivery_details JSONB,
    is_delivered BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE order_s.order_items (
    user_id UUID NOT NULL REFERENCES user_s.users(id) ON DELETE RESTRICT,
    order_id UUID NOT NULL REFERENCES order_s.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES product_s.products(id) ON DELETE RESTRICT,
    items_total INTEGER NOT NULL CHECK (items_total > 0) DEFAULT 1,
    processed_items INTEGER NOT NULL CHECK (processed_items >= 0) DEFAULT 0,
    property JSONB NOT NULL CHECK (jsonb_typeof(property) = 'object') DEFAULT '{}'::JSONB,
    price BIGINT NOT NULL DEFAULT 0,
    confirmed_vendors JSONB NOT NULL CHECK (jsonb_typeof(confirmed_vendors) = 'array') DEFAULT '[]'::JSONB,
    PRIMARY KEY (user_id, order_id, product_id)
);

CREATE OR REPLACE FUNCTION order_s.update_order_after_add_order_items() RETURNS TRIGGER
AS $$
    DECLARE
        product_price BIGINT;
    BEGIN
        BEGIN
            product_price := product_s.get_price(new.product_id, new.property);
        EXCEPTION
            WHEN OTHERS THEN
                RAISE EXCEPTION 'failed to get price with property: %', SQLERRM;
        END;
        IF product_price != NEW.price THEN
            RAISE EXCEPTION 'invalid price';
        END IF;

        UPDATE product_s.products
            SET available_quantity = (available_quantity - NEW.items_total)
            WHERE id = NEW.product_id;

        UPDATE order_s.orders
            SET
                items_total = (items_total + NEW.items_total),
                total_bill = (total_bill + (NEW.items_total * product_price))
            WHERE id = NEW.order_id;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER update_order_after_add_order_items_t
    AFTER INSERT ON order_s.order_items
    FOR EACH ROW
    EXECUTE FUNCTION order_s.update_order_after_add_order_items();

CREATE OR REPLACE FUNCTION order_s.update_order_after_update_order_items() RETURNS TRIGGER
AS $$
    BEGIN
        UPDATE product_s.products
            SET available_quantity = (available_quantity + OLD.items_total) - NEW.items_total
            WHERE id = NEW.product_id;

        UPDATE order_s.orders
            SET
                items_total = (items_total - OLD.items_total) + NEW.items_total,
                total_bill = (total_bill - (OLD.items_total * OLD.price)) + (NEW.items_total * NEW.price)
            WHERE id = NEW.order_id;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER update_order_after_update_order_items_t
    AFTER UPDATE ON order_s.order_items
    FOR EACH ROW
    EXECUTE FUNCTION order_s.update_order_after_update_order_items();

CREATE OR REPLACE FUNCTION order_s.update_order_after_delete_order_items() RETURNS TRIGGER
AS $$
    BEGIN
        UPDATE order_s.orders
            SET
                items_total = (items_total - OLD.items_total),
                total_bill = total_bill - (OLD.items_total * OLD.price)
            WHERE id = OLD.order_id;

        UPDATE product_s.products
            SET available_quantity = (available_quantity +  OLD.items_total)
            WHERE id = OLD.product_id;

        RETURN OLD;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER update_order_after_delete_order_items_t
    AFTER DELETE ON order_s.order_items
    FOR EACH ROW
    EXECUTE FUNCTION order_s.update_order_after_delete_order_items();
