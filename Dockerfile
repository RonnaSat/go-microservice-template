FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o customers ./cmd/customers
RUN CGO_ENABLED=0 go build -o catalog   ./cmd/catalog
RUN CGO_ENABLED=0 go build -o orders    ./cmd/orders

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /src/customers /customers
COPY --from=build /src/catalog /catalog
COPY --from=build /src/orders /orders
EXPOSE 8081 8082 8083
