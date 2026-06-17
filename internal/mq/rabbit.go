package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"arch-oyu-lab3/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Имена как в методичке: topic exchange + очередь + ключ user.created.
const (
	ExchangeName        = "user.exchange"
	QueueName           = "user.events"
	BindingRoutingKey   = "user.#"
	PublishRoutingKey   = "user.created"
)

// Client — обёртка над одним AMQP-каналом (проще, чем тащить conn/ch везде отдельно).
type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Connect(amqpURL string) (*Client, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("подключение к RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("открытие канала: %w", err)
	}

	return &Client{conn: conn, ch: ch}, nil
}

// SetupTopology объявляет exchange, queue и binding.
// Вызывается при старте обоих сервисов — повторный вызов безопасен (идемпотентно).
func (c *Client) SetupTopology() error {
	if err := c.ch.ExchangeDeclare(
		ExchangeName,
		"topic",
		true,  // durable
		false, // auto-delete
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("exchange %s: %w", ExchangeName, err)
	}

	if _, err := c.ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("queue %s: %w", QueueName, err)
	}

	if err := c.ch.QueueBind(QueueName, BindingRoutingKey, ExchangeName, false, nil); err != nil {
		return fmt.Errorf("binding: %w", err)
	}

	return nil
}

// PublishUserCreated кладёт JSON-событие в exchange.
// routing key user.created попадает под binding user.# и уходит в очередь user.events.
func (c *Client) PublishUserCreated(ctx context.Context, event events.UserCreated) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return c.ch.PublishWithContext(
		ctx,
		ExchangeName,
		PublishRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

// ConsumeUserCreated читает очередь и вызывает handler на каждое сообщение.
// autoAck=false: подтверждаем (Ack) только после успешной обработки.
func (c *Client) ConsumeUserCreated(ctx context.Context, handler func(events.UserCreated) error) error {
	deliveries, err := c.ch.Consume(QueueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", QueueName, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-deliveries:
			if !ok {
				return nil
			}
			if err := c.handleDelivery(msg, handler); err != nil {
				slog.Error("ошибка обработки сообщения", "error", err)
				_ = msg.Nack(false, true) // вернуть в очередь
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func (c *Client) handleDelivery(msg amqp.Delivery, handler func(events.UserCreated) error) error {
	var event events.UserCreated
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}
	return handler(event)
}

func (c *Client) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
