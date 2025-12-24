package repository

import (
	"context"
	"database/sql"
	"errors"
	"ecom/internal/orders/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveOrder(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders(id, customer_id, created_at, status) VALUES ($1,$2,$3,'CREATED')`,
		o.ID, o.CustomerID, o.CreatedAt,
	)
	if err != nil {
		return err
	}

	for _, it := range o.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items(order_id, product_id, quantity, unit_price)
			 VALUES ($1,$2,$3,$4)`,
			o.ID, it.ProductID, it.Quantity, it.UnitPrice,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresRepository) UpsertCustomerCache(ctx context.Context, c *domain.CustomerCache) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO customers_cache(id, name, email, updated_at)
VALUES ($1,$2,$3,$4)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    email = EXCLUDED.email,
    updated_at = EXCLUDED.updated_at
`, c.ID, c.Name, c.Email, c.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpsertProductCache(ctx context.Context, p *domain.ProductCache) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO products_cache(id, title, sku, price, updated_at)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    sku = EXCLUDED.sku,
    price = EXCLUDED.price,
    updated_at = EXCLUDED.updated_at
`, p.ID, p.Title, p.SKU, p.Price, p.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetOrderView(ctx context.Context, orderID string) (*domain.OrderView, error) {
	var out domain.OrderView
	var customerID string
	var name, email sql.NullString

	err := r.db.QueryRowContext(ctx, `
SELECT o.id, o.created_at, o.status, o.customer_id,
       c.name, c.email
FROM orders o
LEFT JOIN customers_cache c ON c.id = o.customer_id
WHERE o.id = $1
`, orderID).Scan(&out.Order.ID, &out.Order.CreatedAt, &out.Order.Status, &customerID, &name, &email)

	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	}
	if err != nil {
		return nil, err
	}

	out.Order.Customer = map[string]any{
		"id":    customerID,
		"name":  nullToAny(name),
		"email": nullToAny(email),
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT oi.product_id, oi.quantity, oi.unit_price,
       p.title, p.sku, p.price
FROM order_items oi
LEFT JOIN products_cache p ON p.id = oi.product_id
WHERE oi.order_id = $1
`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			productID  string
			qty        int
			unitPrice  int64
			title, sku sql.NullString
			price      sql.NullInt64
		)
		if err := rows.Scan(&productID, &qty, &unitPrice, &title, &sku, &price); err != nil {
			return nil, err
		}
		out.Items = append(out.Items, map[string]any{
			"productId": productID,
			"quantity":  qty,
			"unitPrice": unitPrice,
			"product": map[string]any{
				"id":    productID,
				"title": nullToAny(title),
				"sku":   nullToAny(sku),
				"price": nullToAnyInt64(price),
			},
		})
	}
	return &out, nil
}

func nullToAny(s sql.NullString) any {
	if s.Valid {
		return s.String
	}
	return nil
}
func nullToAnyInt64(n sql.NullInt64) any {
	if n.Valid {
		return n.Int64
	}
	return nil
}
