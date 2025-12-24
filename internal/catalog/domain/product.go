package domain

import (
	"context"
	"time"
)

type Product struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	SKU       string    `json:"sku"`
	Price     int64     `json:"price"` // cents
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository interface {
	SaveWithOutbox(ctx context.Context, p *Product, msg *OutboxMessage) error
	Get(ctx context.Context, id string) (*Product, error)
	GetBatch(ctx context.Context, ids []string) ([]Product, error)

	// Outbox methods
	GetUnprocessedMessages(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkMessageProcessed(ctx context.Context, id int64) error
}

type OutboxMessage struct {
	ID         int64
	RoutingKey string
	Payload    []byte
	CreatedAt  time.Time
}
