package repository

import (
	"context"
	"database/sql"
	"errors"
	"ecom/internal/customers/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveWithOutbox(ctx context.Context, c *domain.Customer, msg *domain.OutboxMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO customers(id, name, email, updated_at) VALUES ($1,$2,$3,$4)`,
		c.ID, c.Name, c.Email, c.UpdatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO outbox(routing_key, payload) VALUES ($1,$2)`,
		msg.RoutingKey, msg.Payload,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetUnprocessedMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, routing_key, payload, created_at FROM outbox WHERE processed_at IS NULL ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.OutboxMessage
	for rows.Next() {
		var m domain.OutboxMessage
		if err := rows.Scan(&m.ID, &m.RoutingKey, &m.Payload, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *PostgresRepository) MarkMessageProcessed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE outbox SET processed_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (*domain.Customer, error) {
	var c domain.Customer
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, email FROM customers WHERE id=$1`, id,
	).Scan(&c.ID, &c.Name, &c.Email)

	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
