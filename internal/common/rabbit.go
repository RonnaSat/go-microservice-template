package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Bus struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func Connect(amqpURL string) (*Bus, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	// topic exchange for domain events
	if err := ch.ExchangeDeclare(
		ExchangeName, "topic", true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Bus{Conn: conn, Ch: ch}, nil
}

func (b *Bus) Close() {
	if b.Ch != nil {
		_ = b.Ch.Close()
	}
	if b.Conn != nil {
		_ = b.Conn.Close()
	}
}

func (b *Bus) Publish(routingKey string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return b.Ch.PublishWithContext(ctx,
		ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

func (b *Bus) PublishJSON(routingKey string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return b.Ch.PublishWithContext(ctx,
		ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

type Consumer struct {
	Deliveries <-chan amqp.Delivery
}

func (b *Bus) Consume(queueName string, routingKeys ...string) (*Consumer, error) {
	q, err := b.Ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	for _, key := range routingKeys {
		if err := b.Ch.QueueBind(q.Name, key, ExchangeName, false, nil); err != nil {
			return nil, fmt.Errorf("bind %s: %w", key, err)
		}
	}
	deliveries, err := b.Ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Consumer{Deliveries: deliveries}, nil
}
