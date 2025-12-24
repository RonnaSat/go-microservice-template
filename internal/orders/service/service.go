package service

import (
	"context"
	"time"

	"ecom/internal/orders/domain"
	"github.com/google/uuid"
)

type OrderService struct {
	repo domain.Repository
}

func NewOrderService(repo domain.Repository) *OrderService {
	return &OrderService{repo: repo}
}

type CreateOrderInput struct {
	CustomerID string
	Items      []struct {
		ProductID string
		Quantity  int
		UnitPrice int64
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, in CreateOrderInput) (string, error) {
	orderID := uuid.NewString()
	now := time.Now().UTC()

	items := make([]domain.OrderItem, len(in.Items))
	for i, item := range in.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	order := &domain.Order{
		ID:         orderID,
		CustomerID: in.CustomerID,
		CreatedAt:  now,
		Status:     "CREATED",
		Items:      items,
	}

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return "", err
	}
	return orderID, nil
}

func (s *OrderService) GetOrderView(ctx context.Context, orderID string) (*domain.OrderView, error) {
	return s.repo.GetOrderView(ctx, orderID)
}

// These are called by consumers
func (s *OrderService) UpdateCustomerCache(ctx context.Context, c *domain.CustomerCache) error {
	return s.repo.UpsertCustomerCache(ctx, c)
}

func (s *OrderService) UpdateProductCache(ctx context.Context, p *domain.ProductCache) error {
	return s.repo.UpsertProductCache(ctx, p)
}
