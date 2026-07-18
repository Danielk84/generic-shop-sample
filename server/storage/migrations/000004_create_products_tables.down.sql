DROP TABLE IF EXISTS product_s.product_images;
DROP TABLE IF EXISTS product_s.products_categories CASCADE;
DROP TABLE IF EXISTS product_s.categories CASCADE;
DROP TABLE IF EXISTS product_s.products CASCADE;

DROP FUNCTION IF EXISTS
    product_s.products_variant_detail_validator(),
    product_s.set_depend_fields(),
    product_s.check_duplicate_text_array(TEXT[]),
    product_s.set_is_available(),
    product_s.get_price(UUID, JSONB),
    product_s.get_vendors(UUID, JSONB),
    product_s.get_quantity(UUID, JSONB, TEXT),
    product_s.get_variant_detail(UUID);

DROP SCHEMA IF EXISTS product_s CASCADE;
