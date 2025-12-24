# E-Commerce Microservices (Go)

## Architecture
```mermaid
flowchart LR
  client[Client / Postman]

  subgraph services["Services (Fiber HTTP)"]
    customers["customers :8081"]
    catalog["catalog :8082"]
    orders["orders :8083"]
  end

  subgraph infra["Infrastructure (docker-compose)"]
    rabbit["RabbitMQ topic exchange: domain.events"]
    qCustomers["queue: orders.customers_cache"]
    qProducts["queue: orders.products_cache"]

    pg[(PostgreSQL)]
    customers_db[(customers_db)]
    catalog_db[(catalog_db)]
    orders_db[(orders_db)]
  end

  client -->|HTTP| customers
  client -->|HTTP| catalog
  client -->|HTTP| orders

  customers -->|SQL| customers_db
  catalog -->|SQL| catalog_db
  orders -->|SQL orders + local caches| orders_db

  customers -->|publish customer.upserted| rabbit
  catalog -->|publish product.upserted| rabbit
  rabbit -->|route customer.upserted| qCustomers
  rabbit -->|route product.upserted| qProducts
  qCustomers -->|consume| orders
  qProducts -->|consume| orders

  pg --- customers_db
  pg --- catalog_db
  pg --- orders_db
```

Notes:
- `customers` and `catalog` publish domain events on RabbitMQ.
- `orders` consumes those events to keep local cache tables up to date for `/orders/:id/view`.

## Services
- `customers`: `http://localhost:8081`
- `catalog`: `http://localhost:8082`
- `orders`: `http://localhost:8083`

## Run
- `docker compose up --build`
- Optional UIs: RabbitMQ management `http://localhost:15672` (guest/guest)
