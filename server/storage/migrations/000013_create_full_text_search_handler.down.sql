DROP TRIGGER IF EXISTS _record_products_changes_products_t
    ON product_s.products;
DROP TRIGGER IF EXISTS _record_products_changes_categories_t
    ON product_s.products_categories;

DROP TRIGGER IF EXISTS _01_set_products_search_on_delete_t
    ON full_text_search_s.products_changes;

DROP FUNCTION IF EXISTS
    full_text_search_s.record_products_changes(),
    full_text_search_s.set_products_search(),
    full_text_search_s.gen_products_search(UUID,TEXT,TEXT,JSONB),
    full_text_search_s.tags_to_text(UUID);

DROP TABLE IF EXISTS full_text_search_s.products_changes;

DROP SCHEMA IF EXISTS full_text_search_s;
