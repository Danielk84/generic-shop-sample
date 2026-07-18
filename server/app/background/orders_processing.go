package background

import (
	"bytes"
	"context"
	"encoding/json"
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

type ordersProcess struct {
	client           cache.CacheClient
	log              logger.Logger
	orderStore       queries.OrderStore
	orderItemStore   queries.OrderItemsStore
	productStore     queries.ProductStore
	vendorStore      queries.VendorOrderStore
	vendorCacheStore cache_query.VendorCacheStore
	pagination       int
}

func (o *ordersProcess) start(ctx context.Context) {
	sub := o.client.Subscribe(ctx, ordersProcessingChannel)
	defer sub.Close()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

OuterLoop:
	for {
		ch := sub.Channel()
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				o.log.Warn("orders processing Subscription closed")
				return
			}
			var input queries.OrderID
			if err := json.NewDecoder(bytes.NewBufferString(msg.Payload)).Decode(&input); err != nil {
				o.log.Error("failed decode OrderID", "error", err)
				continue
			}
			if err := o.process(ctx, input); err != nil {
				o.log.Error("failed to process order", "error", err, "id", input.ID)
				continue
			}
		case <-ticker.C:
			page := 0
		InnerLoop:
			for {
				items, err := o.orderStore.NotConfirmedList(ctx, o.pagination, page)
				if err != nil {
					if err != pgx.ErrNoRows {
						o.log.Error("failed to get NotConfirmedList", "error", err)
					}
					continue OuterLoop
				}
				for _, i := range items {
					if err := o.process(ctx, i.OrderID); err != nil {
						o.log.Error("failed to process order", "error", err, "id", i.ID)
						continue InnerLoop
					}
				}
				page += 1
			}
		}
	}
}

func (o *ordersProcess) process(ctx context.Context, id queries.OrderID) error {
	page := 0
	for {
		items, err := o.orderItemStore.FullList(ctx, id.ID, o.pagination, page)
		if err != nil {
			if err == pgx.ErrNoRows {
				break
			}
			return err
		}
	InnerLoop:
		for _, i := range items {
			var remainder int32 = 1
			confirmedVendor := []queries.ProductVendor{}
			for remainder < 1 {
				remainder = i.ItemsTotal - i.ProcessedItems
				if remainder < 1 {
					continue InnerLoop
				}
				vendor, err := o.vendorCacheStore.Turn(ctx, i.ProductID, i.Property)
				if err != nil {
					return err
				}
				quantity, err := o.productStore.GetQuantity(ctx, i.ProductID, i.Property, vendor)
				if err != nil {
					return err
				}
				orderable := min(quantity, remainder)
				err = o.vendorStore.Create(ctx, queries.VendorOrder{
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
				err = o.productStore.SetVendor(ctx, i.ProductID, i.Property, queries.ProductVendor{
					UserID:   vendor,
					Quantity: quantity - orderable,
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
			err = o.orderItemStore.SetConfirmedVendors(ctx,
				queries.OrderItemID{
					OrderItem: queries.OrderItem{
						OrderID:   i.OrderID,
						ProductID: i.ProductID,
					},
					UserID: "",
				},
				confirmedVendor)
			if err != nil {
				return err
			}
			if err = o.orderStore.SetConfirmed(ctx, id, true); err != nil {
				return err
			}
		}
		page += 1
	}
	return nil
}
