DROP TABLE IF EXISTS order_items;

DROP TABLE IF EXISTS orders;

DROP FUNCTION IF EXISTS update_order_after_delete_order_items(), update_order_after_update_order_items(), update_order_after_add_order_items();
