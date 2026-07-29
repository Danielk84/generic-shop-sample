package background

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/cache_query"
	"generic-shop-sample/storage/queries"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	ordersProcessingChannel = "order-processing-channel"
)

type OrdersProcess struct {
	Cache            cache.CacheClient
	Log              logger.Logger
	OrderStore       queries.OrderStore
	OrderItemStore   queries.OrderItemsStore
	ProductStore     queries.ProductStore
	VendorStore      queries.VendorOrderStore
	VendorCacheStore cache_query.VendorCacheStore
	Pagination       int
}

func (o *OrdersProcess) Start(ctx context.Context) {
	sub := o.Cache.Subscribe(ctx, ordersProcessingChannel)
	defer func() { _ = sub.Close() }()

	ch := sub.Channel()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

OuterLoop:
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				o.Log.Warn("orders processing Subscription closed")
				return
			}
			o.Log.Debug("OrdersProcess.Start", "debug", "start of sub channel")

			var input queries.OrderID
			if err := json.NewDecoder(bytes.NewBufferString(msg.Payload)).Decode(&input); err != nil {
				o.Log.Error("failed decode OrderID", "error", err)
				continue
			}
			order, err := o.OrderStore.Get(ctx, input)
			if err != nil {
				o.Log.Error("OrdersProcess.Start", "error", err)
				continue
			}
			if !order.IsPaid {
				o.Log.Warn("OrderProcess.Start",
					"error", "invalid not paid order on order process queue",
					"id", input)
				continue
			}
			if err := o.process(ctx, input); err != nil {
				o.Log.Error("failed to process order", "error", err, "id", input.ID)
				continue
			}
		case <-ticker.C:
			o.Log.Debug("OrdersProcess.Start", "debug", "start of ticker")

		InnerLoop:
			for page := 1; ; page++ {
				items, err := o.OrderStore.NotConfirmedList(ctx, o.Pagination, page)
				if err != nil {
					if err != pgx.ErrNoRows {
						o.Log.Error("failed to get NotConfirmedList", "error", err)
					}
					continue OuterLoop
				}
				for _, i := range items {
					if err := o.process(ctx, i.OrderID); err != nil {
						o.Log.Error("failed to process order", "error", err, "id", i.ID)
						continue InnerLoop
					}
				}
			}
		case <-ctx.Done():
			o.Log.Debug("OrdersProcess.Start", "debug", "end of OrdersProcess task")
			return
		}
	}
}

func (o *OrdersProcess) process(ctx context.Context, id queries.OrderID) error {
	for page := 1; ; page++ {
		items, err := o.OrderItemStore.FullList(ctx, id.ID, o.Pagination, page)
		if err != nil {
			if err == pgx.ErrNoRows {
				o.Log.Debug("OrdersProcess.process",
					"debug", "end of order items",
					"page", page)
				break
			}
			return err
		}
		for _, i := range items {
			var remainder int32 = 1
			confirmedVendor := []queries.ProductVendor{}

		InnertLoop:
			for remainder > 0 {
				remainder = i.ItemsTotal - i.ProcessedItems
				if remainder < 1 {
					o.Log.Debug("OrdersProcess.process",
						"debug", "not remainder order item",
						"remainder", remainder,
						"page", page)
					continue InnertLoop
				}
				vendor, err := o.VendorCacheStore.Turn(ctx, i.ProductID, i.Property)
				if err != nil {
					return err
				}
				o.Log.Debug("OrdersProcess.process",
					"debug", "get turn",
					"productID", i.ProductID,
					"property", i.Property,
					"turn", vendor,
					"page", page)

				quantity, err := o.ProductStore.GetQuantity(ctx, i.ProductID, i.Property, vendor)
				if err != nil {
					return err
				}
				o.Log.Debug("OrdersProcess.process",
					"debug", "get quantity",
					"quantity", quantity,
					"page", page)

				orderable := min(quantity, remainder)
				o.Log.Debug("OrdersProcess.process",
					"debug", "min(quantity, remainder)",
					"orderable", orderable,
					"page", page)

				err = o.VendorStore.Create(ctx, queries.VendorOrder{
					VendorOrderDelivere: queries.VendorOrderDelivere{
						UserID:      vendor,
						OrderID:     i.OrderID,
						ProductID:   i.ProductID,
						Property:    i.Property,
						IsDelivered: false,
					},
					Quantity:  orderable,
					TotalBill: int64(orderable) * i.Price,
				})
				if err != nil {
					return err
				}
				err = o.ProductStore.SetVendor(ctx, queries.UpdateProductVendor{
					ProductIDRequest:       queries.ProductIDRequest{ID: i.ProductID},
					ProductVendor:          queries.ProductVendor{UserID: vendor, Quantity: quantity - orderable},
					ProductPropertyRequest: queries.ProductPropertyRequest{Property: i.Property},
				})
				if err != nil {
					return err
				}
				confirmedVendor = append(confirmedVendor, queries.ProductVendor{
					UserID:   vendor,
					Quantity: orderable,
				})
				i.ProcessedItems += orderable
			}
			o.Log.Debug("OrdersProcess.process",
				"confirmedVendor", confirmedVendor,
				"page", page)

			err = o.OrderItemStore.SetConfirmedVendors(ctx,
				queries.OrderItemID{
					OrderItem: queries.OrderItem{
						OrderID:   i.OrderID,
						ProductID: i.ProductID,
					},
					UserID: id.UserID,
				},
				confirmedVendor)
			if err != nil {
				return err
			}
			if err = o.OrderStore.SetConfirmed(ctx, id, true); err != nil {
				return err
			}
		}
	}
	return nil
}

var _ BackgroundTask = (*OrdersProcess)(nil)

func SendOrderDone(ctx context.Context, cache cache.CacheClient, order queries.OrderID) error {
	msg, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode OrderID, %s", err)
	}
	if err := cache.Publish(ctx, ordersProcessingChannel, msg).Err(); err != nil {
		return err
	}
	return nil
}
