DROP TABLE IF EXISTS order_s.order_items;

DROP TABLE IF EXISTS order_s.orders;

DROP FUNCTION IF EXISTS
    order_s.update_order_after_delete_order_items(),
    order_s.update_order_after_update_order_items(),
    order_s.update_order_after_add_order_items();

DROP SCHEMA IF EXISTS order_s CASCADE;
