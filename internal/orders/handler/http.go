package handler

import (
	"context"
	"time"

	"ecom/internal/orders/service"
	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) RegisterRoutes(app fiber.Router) {
	app.Post("/orders", h.Create)
	app.Get("/orders/:id/view", h.GetView)
}

func (h *OrderHandler) Create(c *fiber.Ctx) error {
	var req struct {
		CustomerID string `json:"customerId"`
		Items      []struct {
			ProductID string `json:"productId"`
			Quantity  int    `json:"quantity"`
			UnitPrice int64  `json:"unitPrice"`
		} `json:"items"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	if req.CustomerID == "" || len(req.Items) == 0 {
		return c.Status(400).SendString("customerId and items required")
	}

	in := service.CreateOrderInput{
		CustomerID: req.CustomerID,
		Items: make([]struct {
			ProductID string
			Quantity  int
			UnitPrice int64
		}, len(req.Items)),
	}
	for i, item := range req.Items {
		in.Items[i].ProductID = item.ProductID
		in.Items[i].Quantity = item.Quantity
		in.Items[i].UnitPrice = item.UnitPrice
	}

	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	id, err := h.svc.CreateOrder(ctx, in)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(map[string]any{"orderId": id})
}

func (h *OrderHandler) GetView(c *fiber.Ctx) error {
	orderID := c.Params("id")

	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	view, err := h.svc.GetOrderView(ctx, orderID)
	if err != nil {
		if err.Error() == "not found" {
			return c.Status(404).SendString("not found")
		}
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(view)
}