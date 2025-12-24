package service

import (
	"context"
	"encoding/json"
	"time"

	"ecom/internal/catalog/domain"
	"ecom/internal/common"
	"github.com/google/uuid"
)

type CatalogService struct {
	repo domain.Repository
}

func NewCatalogService(repo domain.Repository) *CatalogService {
	return &CatalogService{
		repo: repo,
	}
}

func (s *CatalogService) CreateProduct(ctx context.Context, title, sku string, price int64) (*domain.Product, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	p := &domain.Product{
		ID:        id,
		Title:     title,
		SKU:       sku,
		Price:     price,
		UpdatedAt: now,
	}

	evt := common.ProductUpserted{
		ID:        id,
		Title:     title,
		SKU:       sku,
		Price:     price,
		UpdatedAt: now,
	}
	payload, _ := json.Marshal(evt)

	msg := &domain.OutboxMessage{
		RoutingKey: "product.upserted",
		Payload:    payload,
	}

	if err := s.repo.SaveWithOutbox(ctx, p, msg); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *CatalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.Get(ctx, id)
}

func (s *CatalogService) GetProductsBatch(ctx context.Context, ids []string) ([]domain.Product, error) {
	return s.repo.GetBatch(ctx, ids)
}
