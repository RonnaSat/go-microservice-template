package domain

import (
	"context"
	"time"
)

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customerId"`
	CreatedAt  time.Time   `json:"createdAt"`
	Status     string      `json:"status"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unitPrice"`
}

// Read Models / Caches

type CustomerCache struct {
	ID        string
	Name      string
	Email     string
	UpdatedAt time.Time
}

type ProductCache struct {
	ID        string
	Title     string
	SKU       string
	Price     int64
	UpdatedAt time.Time
}

type OrderView struct {
	Order struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
		Status    string    `json:"status"`
		Customer  any       `json:"customer"`
	} `json:"order"`
	Items []any `json:"items"`
}

type Repository interface {
	SaveOrder(ctx context.Context, o *Order) error
	GetOrderView(ctx context.Context, orderID string) (*OrderView, error)
	
	// Cache updates
	UpsertCustomerCache(ctx context.Context, c *CustomerCache) error
	UpsertProductCache(ctx context.Context, p *ProductCache) error
}
