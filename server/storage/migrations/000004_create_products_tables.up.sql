CREATE SCHEMA product_s;

CREATE TABLE product_s.products (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(256) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL CHECK (price >= 0) DEFAULT 0,
    pub_date TIMESTAMP NOT NULL DEFAULT now(),
    available_quantity INTEGER NOT NULL CHECK (available_quantity >= 0) DEFAULT 0,
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    common_detail JSONB NOT NULL CHECK (jsonb_typeof(common_detail) = 'object') DEFAULT '{}'::JSONB,
    variant_detail JSONB NOT NULL CHECK (jsonb_typeof(variant_detail) = 'array') DEFAULT '[]'::JSONB
    -- variante-detail schema
    -- Array<{
    --     property: Map<{ [key: string]: string }>
    --     price: number,
    --     vendors: Array<{
    --        user_id: string,
    --        quantity: number,
    --     }>,
    -- }>
);

CREATE TABLE product_s.categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tag TEXT UNIQUE NOT NULL
);

CREATE TABLE product_s.products_categories (
    product_id UUID REFERENCES product_s.products(id) ON DELETE CASCADE,
    tag TEXT REFERENCES product_s.categories(tag) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT product_categorie_key PRIMARY KEY (product_id, tag)
);

CREATE TABLE product_s.product_images (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7(),
    product_id UUID NOT NULL REFERENCES product_s.products(id) ON DELETE CASCADE,
    img_path TEXT NOT NULL UNIQUE
);

CREATE OR REPLACE FUNCTION product_s.check_duplicate_text_array(arr TEXT[]) RETURNS INTEGER
AS $$
    DECLARE
        dup_count INTEGER;
    BEGIN
        SELECT COUNT(*) - COUNT(DISTINCT x)
        INTO dup_count
        FROM unnest(arr) AS x;

        RETURN dup_count;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION product_s.products_variant_detail_validator() RETURNS TRIGGER
AS $$
    DECLARE
        item JSONB;
        vendor JSONB;
        user_ids TEXT[] := '{}';
        properties TEXT[] := '{}';
    BEGIN
        IF jsonb_typeof(NEW.variant_detail) != 'array' THEN
            RAISE EXCEPTION 'variant_detail must be a JSON array';
        END IF;

        FOR item IN (
            SELECT value
            FROM jsonb_array_elements(NEW.variant_detail)
        )
        LOOP
            IF jsonb_typeof(item->'property') != 'object' THEN
                RAISE EXCEPTION 'property in variant_detail must be a JSON object';
            END IF;
            BEGIN
                properties := array_append(properties, (item->>'property'));
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE EXCEPTION 'failed to process property in variant_detail, error msg: %', SQLERRM;
            END;

            IF jsonb_typeof(item->'vendors') != 'array' THEN
                RAISE EXCEPTION 'vendors in variant_detail must be a JSON array';
            END IF;

            IF NOT (item ? 'price') OR ((item->>'price')::BIGINT) < 0 then
                RAISE EXCEPTION 'variant_detail has price field with number type and most be greater than zero';
            END IF;
            FOR vendor IN (
                SELECT value
                FROM jsonb_array_elements(item->'vendors')
            )
            LOOP
                IF NOT (vendor ? 'user_id') THEN
                    RAISE EXCEPTION 'user_id in vendors in variant_detail most be specified';
                END IF;

                IF NOT (vendor ? 'quantity') OR ((vendor->>'quantity')::INTEGER) < 0 THEN
                    RAISE EXCEPTION 'vendor in variant_detail has quantity field with number type and most be greater than zero';
                END IF;

                BEGIN
                    user_ids := array_append(user_ids, (vendor->>'user_id'));
                EXCEPTION
                    WHEN OTHERS THEN
                        RAISE EXCEPTION 'failed to process user_id "%" in variant_detail, error msg: %', vendor->>'user_id', SQLERRM;
                END;
            END LOOP;
        END LOOP;

        -- check vendors be UNIQUE
        IF product_s.check_duplicate_text_array(user_ids) > 0 THEN
            RAISE EXCEPTION 'duplicated user_id found in vendors array';
        END IF;

        -- check properties be UNIQUE
        IF product_s.check_duplicate_text_array(properties) > 0 THEN
            RAISE EXCEPTION 'duplicated user_id found in vendors array';
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER _01_products_variant_detail_validator_t
    BEFORE INSERT OR UPDATE ON product_s.products
    FOR EACH ROW
    EXECUTE FUNCTION product_s.products_variant_detail_validator();

-- this function set products.price and products.quantity,
-- base on products.variant_detail.
CREATE OR REPLACE FUNCTION product_s.set_depend_fields() RETURNS TRIGGER
AS $$
    DECLARE
        new_len INTEGER;
        i INTEGER;
        j INTEGER;
        item JSONB;
        price BIGINT;
    BEGIN
        new_len := jsonb_array_length(NEW.variant_detail);
        IF new_len = 0 THEN
            NEW.available_quantity := 0;
            NEW.price := 0;
            RETURN new;
        END IF;

        j := 0;
        FOR i IN 0..(new_len-1)
        LOOP
            price := (NEW.variant_detail->i->>'price')::BIGINT;
            IF new.price > price OR price > 0 THEN
                NEW.price = price;
            END IF;

            FOR item in (
                SELECT value
                FROM jsonb_array_elements(NEW.variant_detail->i->'vendors')
            )
            LOOP
                j = j + (item->>'quantity')::INTEGER;
            END LOOP;
        END LOOP;

        IF TG_OP = 'UPDATE' THEN 
            IF OLD.available_quantity != NEW.available_quantity THEN
                IF NEW.available_quantity < j THEN
                    NEW.available_quantity = (j - (j - NEW.available_quantity));
                END IF;
            ELSE
                NEW.available_quantity := j;
            END IF;
        ELSE
            NEW.available_quantity := j;
        END IF;
        IF NEW.available_quantity > j THEN
            NEW.available_quantity := j;
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER _02_set_depend_fields_t
    BEFORE INSERT OR UPDATE ON product_s.products
    FOR EACH ROW
    EXECUTE FUNCTION product_s.set_depend_fields();

CREATE OR REPLACE FUNCTION product_s.set_is_available() RETURNS TRIGGER
AS $$
    BEGIN
        IF NEW.available_quantity > 0 THEN
            NEW.is_available := TRUE;
        ELSE
            NEW.is_available := FALSE;
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER _03_set_is_available_t
    BEFORE INSERT OR UPDATE ON product_s.products
    FOR EACH ROW
    EXECUTE FUNCTION product_s.set_is_available();

CREATE OR REPLACE FUNCTION product_s.get_variant_detail(product_id UUID) RETURNS JSONB
AS $$
    DECLARE
        variant_detail JSONB;
    BEGIN
        SELECT p.variant_detail INTO variant_detail
        FROM product_s.products p
        WHERE p.id = product_id
        LIMIT 1;

        IF variant_detail IS NULL THEN
            RAISE EXCEPTION 'invalid product_id=%', product_id;
        END IF;

        RETURN variant_detail;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION product_s.get_price(product_id UUID, property JSONB) RETURNS BIGINT
AS $$
    DECLARE
        variant_detail JSONB;
        item JSONB;
        price BIGINT;
    BEGIN
        BEGIN
            variant_detail := product_s.get_variant_detail(product_id);
        EXCEPTION
            WHEN OTHERS THEN
                RAISE EXCEPTION 'failed to get_variant_detail in get_price: %', SQLERRM;
        END;

        << loop_block >>
        FOR item IN (
            SELECT value
            FROM jsonb_array_elements(variant_detail)
        )
        LOOP
            IF item->'property' = property THEN
                price := (item->>'price')::BIGINT;
                EXIT loop_block;
            END IF;
        END LOOP loop_block;

        IF price IS NULL THEN
            RAISE EXCEPTION 'invalid property';
        END IF;

        RETURN price;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION product_s.get_vendors(product_id UUID, property JSONB) RETURNS TEXT[]
AS $$
    DECLARE
        variant_detail JSONB;
        vendor JSONB;
        item JSONB;
        user_ids TEXT[] := '{}';
    BEGIN
        BEGIN
            variant_detail := product_s.get_variant_detail(product_id);
        EXCEPTION
            WHEN OTHERS THEN
                RAISE EXCEPTION 'failed to get_variant_detail in get_vendors: %', SQLERRM;
        END;

        << loop_block >>
        FOR item IN (
            SELECT value
            FROM jsonb_array_elements(variant_detail)
        )
        LOOP
            IF item->'property' = property THEN
                FOR vendor IN (
                    SELECT value
                    FROM jsonb_array_elements(item->'vendors')
                )
                LOOP
                    user_ids := array_append(user_ids, (vendor->>'user_id'));
                END LOOP;
                EXIT loop_block;
            END IF;
        END LOOP loop_block;

        IF array_length(user_ids, 1) = 0 THEN
            RAISE EXCEPTION 'invalid property';
        END IF;

        RETURN user_ids;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION product_s.get_quantity(product_id UUID, property JSONB, user_id TEXT) RETURNS INTEGER
AS $$
    DECLARE
        variant_detail JSONB;
        vendor JSONB;
        item JSONB;
    BEGIN
        BEGIN
            variant_detail := product_s.get_variant_detail(product_id);
        EXCEPTION
            WHEN OTHERS THEN
                RAISE EXCEPTION 'failed to get_variant_detail in get_quantity: %', SQLERRM;
        END;

        FOR item IN (
            SELECT value
            FROM jsonb_array_elements(variant_detail)
        )
        LOOP
            IF item->'property' = property THEN
                FOR vendor IN (
                    SELECT value
                    FROM jsonb_array_elements(item->'vendors')
                )
                LOOP
                    IF vendor->>'user_id' = user_id THEN
                        RETURN (vendor->>'quantity')::INTEGER;
                    END IF;
                END LOOP;
            END IF;
        END LOOP;

        RAISE EXCEPTION 'invalid property or user_id';
    END; 
$$ LANGUAGE PLPGSQL;

