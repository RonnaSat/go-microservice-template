package main

import (
	"database/sql"
	"log"
	"os"

	"ecom/internal/common"
	"ecom/internal/orders/app"
	"ecom/internal/orders/handler"
	"ecom/internal/orders/repository"
	"ecom/internal/orders/service"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}

func main() {
	db, err := sql.Open("pgx", mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := initSchema(db); err != nil {
		log.Fatal(err)
	}

	bus, err := common.Connect(mustEnv("AMQP_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	// Dependency Injection
	repo := repository.NewPostgresRepository(db)
	svc := service.NewOrderService(repo)
	h := handler.NewOrderHandler(svc)
	consumers := app.NewConsumers(svc, bus)

	// Start Consumers
	consumers.Start()

	appFiber := fiber.New()
	h.RegisterRoutes(appFiber)

	addr := mustEnv("HTTP_ADDR")
	log.Printf("orders listening on %s", addr)
	log.Fatal(appFiber.Listen(addr))
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS orders (
  id TEXT PRIMARY KEY,
  customer_id TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
  order_id TEXT NOT NULL,
  product_id TEXT NOT NULL,
  quantity INT NOT NULL,
  unit_price BIGINT NOT NULL,
  PRIMARY KEY(order_id, product_id)
);

-- local read model caches (for joining)
CREATE TABLE IF NOT EXISTS customers_cache (
  id TEXT PRIMARY KEY,
  name TEXT,
  email TEXT,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS products_cache (
  id TEXT PRIMARY KEY,
  title TEXT,
  sku TEXT,
  price BIGINT,
  updated_at TIMESTAMP NOT NULL
);
`)
	return err
}
