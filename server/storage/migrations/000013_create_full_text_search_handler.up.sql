CREATE SCHEMA full_text_search_s;

CREATE TABLE full_text_search_s.products_changes (
    product_id UUID PRIMARY KEY REFERENCES product_s.products(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION full_text_search_s.tags_to_text(product_id UUID) RETURNS TEXT
AS $$
    DECLARE
        context TEXT[] := '{}';
    BEGIN
        SELECT ARRAY_AGG(p.tag) INTO context
        FROM products_s.products_categories AS p 
        WHERE p.product_id = product_id;

        RETURN array_to_string(context, ',');
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION full_text_search_s.gen_products_search(
    id UUID,
    name TEXT,
    description TEXT,
    common_detail JSONB
) RETURNS TSVECTOR
AS $$
    BEGIN
        RETURN
            setweight(to_tsvector('simple', full_text_search_s.tags_to_text(id)), 'A') ||
            setweight(to_tsvector('simple', name), 'B') ||
            setweight(to_tsvector('simple', description), 'C') ||
            setweight(jsonb_to_tsvector('simple', common_detail, '["string", "numeric"]'), 'D');
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION full_text_search_s.set_products_search() RETURNS TRIGGER
AS $$
    BEGIN
        UPDATE product_s.products
        SET __search = full_text_search_s.gen_products_search(id, name, description, common_detail)
        WHERE id = OLD.product_id;

        RETURN OLD;
    END;
$$ LANGUAGE PLPGSQL;

-- on removing rows in full_text_search_s.change_products product_s.products.__search must update.
CREATE OR REPLACE TRIGGER _01_set_products_search_on_delete_t
    BEFORE DELETE ON full_text_search_s.products_changes
    FOR EACH ROW
    EXECUTE FUNCTION full_text_search_s.set_products_search();

CREATE OR REPLACE FUNCTION full_text_search_s.record_products_changes() RETURNS TRIGGER
AS $$
    DECLARE
        p_id UUID;
    BEGIN
        IF TG_TABLE_SCHEMA = 'product_s' AND TG_TABLE_NAME = 'products' THEN
            p_id := NEW.id;
        ELSIF TG_TABLE_SCHEMA = 'product_s' AND TG_TABLE_NAME = 'products_categories' THEN
            p_id := NEW.product_id;
        ELSE
            RAISE EXCEPTION 'invalid table for recording products changes';
        END IF;

        INSERT INTO full_text_search_s.products_changes(product_id)
            VALUES (p_id)
            ON CONFLICT DO NOTHING;

        RETURN NEW;
    END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE TRIGGER _record_products_changes_products_t
    AFTER INSERT OR UPDATE ON product_s.products
    FOR EACH ROW
    EXECUTE FUNCTION full_text_search_s.record_products_changes();

CREATE OR REPLACE TRIGGER _record_products_changes_categories_t
    AFTER INSERT OR UPDATE ON product_s.products_categories
    FOR EACH ROW
    EXECUTE FUNCTION full_text_search_s.record_products_changes();
