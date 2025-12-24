package service

import (
	"context"
	"encoding/json"
	"time"

	"ecom/internal/common"
	"ecom/internal/customers/domain"
	"github.com/google/uuid"
)

type CustomerService struct {
	repo domain.Repository
}

func NewCustomerService(repo domain.Repository) *CustomerService {
	return &CustomerService{
		repo: repo,
	}
}

func (s *CustomerService) Create(ctx context.Context, name, email string) (*domain.Customer, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	c := &domain.Customer{
		ID:        id,
		Name:      name,
		Email:     email,
		UpdatedAt: now,
	}

	evt := common.CustomerUpserted{
		ID:        id,
		Name:      name,
		Email:     email,
		UpdatedAt: now,
	}
	payload, _ := json.Marshal(evt)

	msg := &domain.OutboxMessage{
		RoutingKey: "customer.upserted",
		Payload:    payload,
	}

	if err := s.repo.SaveWithOutbox(ctx, c, msg); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *CustomerService) Get(ctx context.Context, id string) (*domain.Customer, error) {
	return s.repo.Get(ctx, id)
}
