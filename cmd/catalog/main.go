package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"ecom/internal/catalog/domain"
	"ecom/internal/catalog/handler"
	"ecom/internal/catalog/repository"
	"ecom/internal/catalog/service"
	"ecom/internal/common"

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
	svc := service.NewCatalogService(repo)
	h := handler.NewCatalogHandler(svc)

	relay := &Relay{repo: repo, bus: bus}
	go relay.Start()

	app := fiber.New()
	h.RegisterRoutes(app)

	addr := mustEnv("HTTP_ADDR")
	log.Printf("catalog listening on %s", addr)
	log.Fatal(app.Listen(addr))
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS products (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  sku TEXT NOT NULL,
  price BIGINT NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS outbox (
  id BIGSERIAL PRIMARY KEY,
  routing_key TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  processed_at TIMESTAMP
);
`)
	return err
}

type Relay struct {
	repo domain.Repository
	bus  *common.Bus
}

func (r *Relay) Start() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		msgs, err := r.repo.GetUnprocessedMessages(context.Background(), 10)
		if err != nil {
			log.Printf("relay: get messages: %v", err)
			continue
		}

		for _, m := range msgs {
			err := r.bus.Publish(m.RoutingKey, m.Payload)
			if err != nil {
				log.Printf("relay: publish %d: %v", m.ID, err)
				continue
			}

			if err := r.repo.MarkMessageProcessed(context.Background(), m.ID); err != nil {
				log.Printf("relay: mark processed %d: %v", m.ID, err)
			}
		}
	}
}
