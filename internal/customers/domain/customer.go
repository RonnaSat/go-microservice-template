package domain

import (
	"context"
	"time"
)

type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository interface {
	SaveWithOutbox(ctx context.Context, c *Customer, msg *OutboxMessage) error
	Get(ctx context.Context, id string) (*Customer, error)

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
