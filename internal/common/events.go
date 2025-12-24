package common

import "time"

const ExchangeName = "domain.events"

type CustomerUpserted struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProductUpserted struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	SKU       string    `json:"sku"`
	Price     int64     `json:"price"` // cents
	UpdatedAt time.Time `json:"updatedAt"`
}
