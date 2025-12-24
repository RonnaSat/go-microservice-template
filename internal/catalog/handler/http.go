package handler

import (
	"context"
	"time"

	"ecom/internal/catalog/service"

	"github.com/gofiber/fiber/v2"
)

type CatalogHandler struct {
	svc *service.CatalogService
}

func NewCatalogHandler(svc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) RegisterRoutes(app fiber.Router) {
	app.Post("/products", h.Create)
	app.Post("/products/batch", h.GetBatch)
	app.Get("/products/:id", h.Get)
}

func (h *CatalogHandler) Create(c *fiber.Ctx) error {
	var in struct {
		Title string `json:"title"`
		SKU   string `json:"sku"`
		Price int64  `json:"price"`
	}
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	p, err := h.svc.CreateProduct(ctx, in.Title, in.SKU, in.Price)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(p)
}

func (h *CatalogHandler) GetBatch(c *fiber.Ctx) error {
	var in struct {
		IDs []string `json:"ids"`
	}
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	products, err := h.svc.GetProductsBatch(ctx, in.IDs)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(products)
}

func (h *CatalogHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	p, err := h.svc.GetProduct(ctx, id)
	if err != nil {
		if err.Error() == "not found" {
			return c.Status(404).SendString("not found")
		}
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(p)
}
