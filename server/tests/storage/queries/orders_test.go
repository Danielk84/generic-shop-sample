package queries_test

import (
	"context"
	"fmt"
	"generic-shop-sample/app/background"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/cache_query"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	tu "generic-shop-sample/tests/internal/testutils"
	"reflect"
	"testing"
	"time"
)

// because of order and order-items and vendor-order have relation,
// so test combined together.
func TestOrders(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	db := tu.DBTestSetup(ctx, config)
	defer db.Close()
	session := db.GetSession()
	log := logger.GetLogger()
	s := testOrders{
		ctx:     ctx,
		session: session,
		config:  config,

		userStore:        queries.NewUserStore(session, log),
		productStore:     queries.NewProductStore(session, log),
		orderStore:       queries.NewOrderStore(session, log),
		orderItemStore:   queries.NewOrderItemsStore(session, log),
		vendorOrderStore: queries.NewVendorOrderStore(session, log),
	}

	s.setup(t)

	s.orders_create(t)
	s.orders_customerList(t)
	s.orders_fullList(t)
	s.orders_BeforeNotConfirmedList(t)
	s.orders_get(t)
	s.orders_setUserInfo(t)
	s.orders_verifyUserInfo(t)

	s.orderItems_Create(t)
	s.orderItems_CustomerList(t)
	s.orderItems_AdminList(t)
	s.orderItems_FullList(t)
	s.orderItems_Delete(t)
	s.orderItems_SetItemsTotal(t)
	s.orders_setPaymentStatus(t)

	s.orderItems_SetConfirmedVendors(t)

	s.vendorOrder_list(t)
	s.vendorOrder_setIsDelivered(t)

	s.orders_setConfirmed(t)
	s.orders_deleteExpiredOrders(t)
}

type testOrders struct {
	ctx     context.Context
	session database.Session
	config  config.Config

	userStore        queries.UserStore
	productStore     queries.ProductStore
	orderStore       queries.OrderStore
	orderItemStore   queries.OrderItemsStore
	vendorOrderStore queries.VendorOrderStore

	user      queries.UserResponse
	productID string
	orderID   string
	property  queries.ProductProperty
	vendorID  string
}

func (o *testOrders) setup(t *testing.T) {
	var err error
	o.user, err = o.userStore.Get(o.ctx, "customerUser")
	if err != nil {
		t.Fatalf("failed to get customer user: %s", err)
	}
	vendor, err := o.userStore.Get(o.ctx, "vendorUser")
	if err != nil {
		t.Fatalf("failed to get vendor user: %s", err)
	}
	o.vendorID = vendor.ID

	_, err = o.session.Exec(o.ctx, "TRUNCATE order_s.order_items, order_s.orders, product_s.products CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables in order setup: %s", err)
	}
	err = o.productStore.Create(o.ctx, queries.CreateProductRequest{
		Name:        "test order product",
		Description: "hello",
		CommonDetail: queries.ProductProperty{
			"info": "some info",
		},
	})
	if err != nil {
		t.Fatalf("failed to create product: %s", err)
	}
	productList, err := o.productStore.AdminList(o.ctx, 1, 1)
	if err != nil {
		t.Fatalf("failed to get list of product: %s", err)
	}
	if l := len(productList); l != 1 {
		t.Fatalf("expected one product but got: %d", l)
	}
	o.productID = productList[0].ID

	o.property = queries.ProductProperty{"info": "some info"}
	err = o.productStore.SetVariantDetail(o.ctx, o.productID, []queries.ProductVariantDetail{
		{
			ProductPropertyRequest: queries.ProductPropertyRequest{
				Property: o.property,
			},
			Vendors: []queries.ProductVendor{},
			Price:   123,
		},
	})
	if err != nil {
		t.Fatalf("failed to set variant detail: %s", err)
	}

	err = o.productStore.SetVendor(o.ctx, queries.UpdateProductVendor{
		ProductIDRequest: queries.ProductIDRequest{
			ID: o.productID,
		},
		ProductVendor: queries.ProductVendor{
			UserID:   vendor.ID,
			Quantity: 10,
		},
		ProductPropertyRequest: queries.ProductPropertyRequest{
			Property: o.property,
		},
	})
	if err != nil {
		t.Fatalf(`failed to set vendor on product: "%s", vendor: "%s"`, o.productID, vendor.ID)
	}
}

func (o *testOrders) orders_create(t *testing.T) {
	orderID_1, err := o.orderStore.Create(o.ctx, o.user.ID)
	if err != nil {
		t.Fatalf("failed to create order_1: %s", err)
	}
	orderID_2, err := o.orderStore.Create(o.ctx, o.user.ID)
	if err != nil {
		t.Fatalf("failed to create order_2: %s", err)
	}
	if orderID_1 != orderID_2 {
		t.Fatalf(`expected same orders id but orderID_1="%s", orderID_2="%s"`, orderID_1, orderID_2)
	}
	o.orderID = orderID_1
}

func (o *testOrders) orders_customerList(t *testing.T) {
	cList, err := o.orderStore.CustomerList(o.ctx, o.user.ID, 1, 1)
	if err != nil {
		t.Fatalf("failed to get customer list orders: %s", err)
	}
	if l := len(cList); l != 1 {
		t.Fatalf(`excepted one order but got: "%d"`, l)
	}
	if orderID := cList[0].ID; orderID != o.orderID {
		t.Fatalf(`expected same orders id but o.orderID="%s", orderID="%s"`, o.orderID, orderID)
	}
}

func (o *testOrders) orders_fullList(t *testing.T) {
	fList, err := o.orderStore.FullList(o.ctx, 1, 1)
	if err != nil {
		t.Errorf("failed to get full list: %s", err)
	}
	if l := len(fList); l != 1 {
		t.Fatalf(`excepted one order but got: "%d"`, l)
	}
	if orderID := fList[0].ID; orderID != o.orderID {
		t.Fatalf(`expected same orders id but o.orderID="%s", orderID="%s"`, o.orderID, orderID)
	}
}

func (o *testOrders) orders_BeforeNotConfirmedList(t *testing.T) {
	noList, err := o.orderStore.FullList(o.ctx, 1, 1)
	if err != nil {
		t.Errorf("failed to get full list: %s", err)
	}
	if l := len(noList); l != 1 {
		t.Fatalf(`excepted one order but got: "%d"`, l)
	}
	if orderID := noList[0].ID; orderID != o.orderID {
		t.Fatalf(`expected same orders id but o.orderID="%s", orderID="%s"`, o.orderID, orderID)
	}
}

func (o *testOrders) orders_get(t *testing.T) {
	order, err := o.orderStore.Get(o.ctx, queries.OrderID{
		ID:     o.orderID,
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to get order base on user: %s", err)
	}
	if order.ID != o.orderID || order.IsConfirmed || order.IsPaid {
		t.Fatalf(`invalid order, o.orderID="%s", order="%v"`, o.orderID, order)
	}
}

func (o *testOrders) orders_setUserInfo(t *testing.T) {
	info := queries.OrderUserInfo{
		ZipCode: "12345677",
		Address: "1, 2, 3 ,4, Thanks - Thanks", // :)
	}
	err := o.orderStore.SetUserInfo(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		info)
	if err != nil {
		t.Fatalf("failed to set user info: %s", err)
	}
	order, err := o.orderStore.Get(o.ctx, queries.OrderID{
		ID:     o.orderID,
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to get order base on user: %s", err)
	}
	if info.Address != order.Address ||
		info.ZipCode != order.ZipCode ||
		order.IsConfirmed ||
		order.IsDelivered {
		t.Fatalf(`invalid order, info="%v", order="%v"`, info, order)
	}
}

func (o *testOrders) orders_verifyUserInfo(t *testing.T) {
	err := o.orderStore.VerifyUserInfo(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		true)
	if err != nil {
		t.Fatalf("failed to verify user info: %s", err)
	}
	order, err := o.orderStore.Get(o.ctx, queries.OrderID{
		ID:     o.orderID,
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to get order base on user: %s", err)
	}
	if !order.IsVerified {
		t.Fatalf("invalid IsVerified field")
	}
}

func (o *testOrders) orders_setPaymentStatus(t *testing.T) {
	payment := queries.PaymentStatus{
		PaymentSummary: queries.ProductProperty{"info": "some info"},
		IsPaid:         true,
	}
	err := o.orderStore.SetPaymentStatus(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		payment)
	if err != nil {
		t.Fatalf("failed to set payment status: %s", err)
	}
	order, err := o.orderStore.Get(o.ctx, queries.OrderID{
		ID:     o.orderID,
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to get order base on user: %s", err)
	}
	if !order.IsPaid {
		t.Fatal("invalid payment status: IsPaid=false")
	}
	if !reflect.DeepEqual(order.PaymentSummary[0], payment.PaymentSummary) {
		t.Fatalf(`invalid payment summary, PaymentSummary: "%v", payment "%v"`,
			order.PaymentSummary,
			payment)
	}
}

func (o *testOrders) orders_setConfirmed(t *testing.T) {
	err := o.orderStore.SetConfirmed(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		true)
	if err != nil {
		t.Fatalf("failed to verify user info: %s", err)
	}
	order, err := o.orderStore.Get(o.ctx, queries.OrderID{
		ID:     o.orderID,
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to get order base on user: %s", err)
	}
	if !order.IsConfirmed {
		t.Fatalf("invalid IsConfirmed field")
	}
}

func (o *testOrders) orders_deleteExpiredOrders(t *testing.T) {
	orderID, err := o.orderStore.Create(o.ctx, o.user.ID)
	if err != nil {
		t.Fatalf("failed to create order: %s", err)
	}
	if orderID == o.orderID {
		t.Fatalf("failed to create new order after pay")
	}
	_, err = o.session.Exec(o.ctx, `UPDATE order_s.orders
		SET started_at = (NOW() - INTERVAL '24 hours')
		WHERE id = $1::UUID`,
		orderID)
	if err != nil {
		t.Fatalf("failed to update started_at: %s", err)
	}
	fList, err := o.orderStore.FullList(o.ctx, 2, 1)
	if err != nil {
		t.Fatalf("failed to get full list of order: %s", err)
	}
	logger.GetLogger().Info("testOrders.orders_deleteExpiredOrders", "info", fmt.Sprintf("%+v", fList))
	err = o.orderStore.DeleteExpiredOrders(o.ctx)
	if err != nil {
		t.Fatalf("failed to delete expired orders: %s", err)
	}
	fList, err = o.orderStore.FullList(o.ctx, 2, 1)
	if err != nil {
		t.Fatalf("failed to get full list of order: %s", err)
	}
	if l := len(fList); l != 1 {
		t.Fatalf(`expected one paid order in list but got="%d"`, l)
	}
	if id := fList[0].ID; id == orderID {
		t.Fatalf(`expected orderID="%s", but got="%s", o.orderID="%s", items="%+v"`, orderID, id, o.orderID, fList)
	}
}

func (o *testOrders) orderItems_Create(t *testing.T) {
	beforeOrderProduct, err := o.productStore.Get(o.ctx, o.productID)
	if err != nil {
		t.Fatalf("failed to get product for before create order: %s", err)
	}

	err = o.orderItemStore.Create(o.ctx, o.user.ID,
		queries.OrderItemRequest{
			OrderItem: queries.OrderItem{
				OrderID:   o.orderID,
				ProductID: o.productID,
			},
			Price:    123,
			Property: o.property,
		})
	if err != nil {
		t.Fatalf("failed to create order item: %s", err)
	}

	afterOrderProduct, err := o.productStore.Get(o.ctx, o.productID)
	if err != nil {
		t.Fatalf("failed to get product for after create order: %s", err)
	}

	if beforeOrderProduct.AvailableQuantity != afterOrderProduct.AvailableQuantity+1 {
		t.Fatalf(`invalid available quantity after creating order item, before="%d", after="%d"`,
			beforeOrderProduct.AvailableQuantity,
			afterOrderProduct.AvailableQuantity)
	}
}

func (o *testOrders) orderItems_CustomerList(t *testing.T) {
	cList, err := o.orderItemStore.CustomerList(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		1, 1)
	if err != nil {
		t.Fatalf("failed to get order items customer list: %s", err)
	}
	if l := len(cList); l != 1 {
		t.Fatalf("invalid list len: %d", l)
	}
	item := cList[0]
	if item.ItemsTotal != 1 {
		t.Fatalf(`invalid item total, got="%d"`, item.ItemsTotal)
	}
}

func (o *testOrders) orderItems_AdminList(t *testing.T) {
	aList, err := o.orderItemStore.AdminList(o.ctx, o.orderID, 1, 1)
	if err != nil {
		t.Fatalf("failed to get order items admin list: %s", err)
	}
	if l := len(aList); l != 1 {
		t.Fatalf("invalid list len: %d", l)
	}
	item := aList[0]
	if item.Price != 123 {
		t.Fatalf(`invalid price, got="%d"`, item.Price)
	}
}

func (o *testOrders) orderItems_FullList(t *testing.T) {
	fList, err := o.orderItemStore.FullList(o.ctx, o.orderID, 1, 1)
	if err != nil {
		t.Fatalf("failed to get order items full list: %s", err)
	}
	if l := len(fList); l != 1 {
		t.Fatalf("invalid list len: %d", l)
	}
	item := fList[0]
	if !reflect.DeepEqual(item.Property, o.property) {
		t.Fatalf(`invalid property, got="%v"`, item.Property)
	}
}

func (o *testOrders) orderItems_Delete(t *testing.T) {
	beforeOrderProduct, err := o.productStore.Get(o.ctx, o.productID)
	if err != nil {
		t.Fatalf("failed to get product for before create order: %s", err)
	}

	err = o.orderItemStore.Delete(o.ctx, queries.OrderItemID{
		OrderItem: queries.OrderItem{
			OrderID:   o.orderID,
			ProductID: o.productID,
		},
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to create order item: %s", err)
	}

	afterOrderProduct, err := o.productStore.Get(o.ctx, o.productID)
	if err != nil {
		t.Fatalf("failed to get product for after delete order: %s", err)
	}

	if beforeOrderProduct.AvailableQuantity+1 != afterOrderProduct.AvailableQuantity {
		t.Fatalf(`invalid available quantity after deleting order item, before="%d", after="%d"`,
			beforeOrderProduct.AvailableQuantity,
			afterOrderProduct.AvailableQuantity)
	}

	o.orderItems_Create(t)
}

func (o *testOrders) orderItems_SetItemsTotal(t *testing.T) {
	err := o.orderItemStore.SetItemsTotal(o.ctx,
		queries.OrderItemID{
			OrderItem: queries.OrderItem{
				OrderID:   o.orderID,
				ProductID: o.productID,
			},
			UserID: o.user.ID,
		}, 5)
	if err != nil {
		t.Fatalf("failed to set items total: %s", err)
	}
	cList, err := o.orderItemStore.CustomerList(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		1, 1)
	if err != nil {
		t.Fatalf("failed to get order items customer list: %s", err)
	}
	if l := len(cList); l != 1 {
		t.Fatalf("invalid list len: %d", l)
	}
	item := cList[0]
	if item.ItemsTotal != 5 {
		t.Fatalf(`invalid items total after set items total, got="%d"`, item.ItemsTotal)
	}
	product, err := o.productStore.Get(o.ctx, o.productID)
	if err != nil {
		t.Fatalf("failed to get product after set item total: %s", err)
	}
	if product.AvailableQuantity != 5 {
		t.Fatalf(`invalid available quantity in product after set items total, got="%d"`,
			product.AvailableQuantity)
	}
}

func (o *testOrders) orderItems_SetConfirmedVendors(t *testing.T) {
	log := logger.GetLogger()
	c := tu.CacheTestSetup(o.ctx, o.config)
	client := c.GetCache(cache.OrdersCache)
	ctx, cancel := context.WithCancel(o.ctx)
	defer cancel()
	go func() {
		task := background.OrdersProcess{
			Cache:          client,
			Log:            log,
			OrderStore:     o.orderStore,
			OrderItemStore: o.orderItemStore,
			ProductStore:   o.productStore,
			VendorStore:    o.vendorOrderStore,
			VendorCacheStore: cache_query.NewVendorCacheStore(
				client,
				log,
				o.productStore),
			Pagination: 20,
		}
		task.Start(ctx)
	}()
	time.Sleep(5 * time.Second)
	err := background.SendOrderDone(ctx, client, queries.OrderID{
		ID:     o.orderID,
		UserID: o.user.ID,
	})
	if err != nil {
		t.Fatalf("failed to send order done in cache publisher: %s", err)
	}
	time.Sleep(5 * time.Second)
	cList, err := o.orderItemStore.CustomerList(o.ctx,
		queries.OrderID{ID: o.orderID, UserID: o.user.ID},
		1, 1)
	if err != nil {
		t.Fatalf("failed to get order items customer list: %s", err)
	}
	if l := len(cList); l != 1 {
		t.Fatalf("invalid list len: %d", l)
	}
	item := cList[0]
	if l := len(item.ConfirmedVendors); l != 1 {
		t.Fatalf(`invalid confirmed vendors list in order item, got="%v"`, item)
	}
	vendor := item.ConfirmedVendors[0]
	if vendor.UserID != o.vendorID {
		t.Fatalf(`invalid vendor id after process done order, expect="%s", got="%s"`,
			o.vendorID,
			vendor.UserID)
	}
}

func (o *testOrders) vendorOrder_list(t *testing.T) {
	voList, err := o.vendorOrderStore.List(o.ctx, o.vendorID, 1, 1)
	if err != nil {
		t.Fatalf("failed to get vendor order list: %s", err)
	}
	if l := len(voList); l != 1 {
		t.Fatalf(`invalid vendor order list len, got="%v"`, voList)
	}
	item := voList[0]
	if item.Quantity != 5 {
		t.Fatalf(`invalid vendor order quantity, got="%v"`, item)
	}
}

func (o *testOrders) vendorOrder_setIsDelivered(t *testing.T) {
	err := o.vendorOrderStore.SetIsDelivered(o.ctx,
		queries.VendorOrderDelivere{
			UserID:      o.vendorID,
			OrderID:     o.orderID,
			ProductID:   o.productID,
			Property:    o.property,
			IsDelivered: true,
		})
	if err != nil {
		t.Fatalf("failed to set is delivered in vendor order store: %s", err)
	}
	voList, err := o.vendorOrderStore.List(o.ctx, o.vendorID, 1, 1)
	if err != nil {
		t.Fatalf("failed to get vendor order list: %s", err)
	}
	if l := len(voList); l != 1 {
		t.Fatalf(`invalid vendor order list len, got="%d"`, l)
	}
	item := voList[0]
	if !item.IsDelivered {
		t.Fatalf(`invalid isDelivered in vendor order, got="%v"`, item)
	}
}
