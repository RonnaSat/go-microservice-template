package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ecom/internal/common"
	"ecom/internal/orders/domain"
	"ecom/internal/orders/service"
)

type Consumers struct {
	svc *service.OrderService
	bus *common.Bus
}

func NewConsumers(svc *service.OrderService, bus *common.Bus) *Consumers {
	return &Consumers{svc: svc, bus: bus}
}

func (c *Consumers) Start() {
	go c.consumeCustomers()
	go c.consumeProducts()
}

func (c *Consumers) consumeCustomers() {
	cons, err := c.bus.Consume("orders.customers_cache", "customer.upserted")
	if err != nil {
		log.Fatalf("consume customers: %v", err)
	}

	for d := range cons.Deliveries {
		var evt common.CustomerUpserted
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			log.Printf("bad customer event: %v", err)
			_ = d.Nack(false, false)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := c.svc.UpdateCustomerCache(ctx, &domain.CustomerCache{
			ID:        evt.ID,
			Name:      evt.Name,
			Email:     evt.Email,
			UpdatedAt: evt.UpdatedAt,
		})
		cancel()

		if err != nil {
			log.Printf("cache customer upsert failed: %v", err)
			_ = d.Nack(false, true)
			continue
		}
		_ = d.Ack(false)
	}
}

func (c *Consumers) consumeProducts() {
	cons, err := c.bus.Consume("orders.products_cache", "product.upserted")
	if err != nil {
		log.Fatalf("consume products: %v", err)
	}

	for d := range cons.Deliveries {
		var evt common.ProductUpserted
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			log.Printf("bad product event: %v", err)
			_ = d.Nack(false, false)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := c.svc.UpdateProductCache(ctx, &domain.ProductCache{
			ID:        evt.ID,
			Title:     evt.Title,
			SKU:       evt.SKU,
			Price:     evt.Price,
			UpdatedAt: evt.UpdatedAt,
		})
		cancel()

		if err != nil {
			log.Printf("cache product upsert failed: %v", err)
			_ = d.Nack(false, true)
			continue
		}
		_ = d.Ack(false)
	}
}
