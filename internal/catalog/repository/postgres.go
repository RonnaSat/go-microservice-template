package repository

import (
	"context"
	"database/sql"
	"errors"
	"ecom/internal/catalog/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveWithOutbox(ctx context.Context, p *domain.Product, msg *domain.OutboxMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO products(id, title, sku, price, updated_at) VALUES ($1,$2,$3,$4,$5)`,
		p.ID, p.Title, p.SKU, p.Price, p.UpdatedAt,
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

func (r *PostgresRepository) Get(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, sku, price FROM products WHERE id=$1`, id,
	).Scan(&p.ID, &p.Title, &p.SKU, &p.Price)

	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) GetBatch(ctx context.Context, ids []string) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		var p domain.Product
		err := r.db.QueryRowContext(ctx,
			`SELECT id, title, sku, price FROM products WHERE id=$1`, id,
		).Scan(&p.ID, &p.Title, &p.SKU, &p.Price)
		if err == nil {
			out = append(out, p)
		}
	}
	return out, nil
}
