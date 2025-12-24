package handler

import (
	"context"
	"time"

	"ecom/internal/customers/service"
	"github.com/gofiber/fiber/v2"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) RegisterRoutes(app fiber.Router) {
	app.Post("/customers", h.Create)
	app.Get("/customers/:id", h.Get)
}

func (h *CustomerHandler) Create(c *fiber.Ctx) error {
	var in struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	cust, err := h.svc.Create(ctx, in.Name, in.Email)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(cust)
}

func (h *CustomerHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	cust, err := h.svc.Get(ctx, id)
	if err != nil {
		if err.Error() == "not found" {
			return c.Status(404).SendString("not found")
		}
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(cust)
}