package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/discord-subscriptions/shared/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrNotConnected = errors.New("not connected to rabbitmq broker")
)

// RabbitMQClient implements the Client interface for RabbitMQ.
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex
	logger  *slog.Logger
	url     string
}

// NewRabbitMQClient establishes a connection to RabbitMQ and returns a managed client.
func NewRabbitMQClient(url string, logger *slog.Logger) (*RabbitMQClient, error) {
	if logger == nil {
		logger = slog.Default()
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	client := &RabbitMQClient{
		conn:    conn,
		channel: ch,
		logger:  logger,
		url:     url,
	}

	return client, nil
}

// IsConnected reports whether the client has an active connection and channel.
func (r *RabbitMQClient) IsConnected() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn != nil && !r.conn.IsClosed() && r.channel != nil && !r.channel.IsClosed()
}

// Publish serializes a domain event to JSON and publishes it to the specified exchange and routing key.
func (r *RabbitMQClient) Publish(ctx context.Context, exchange, routingKey string, event events.Event) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.channel == nil || r.channel.IsClosed() {
		return ErrNotConnected
	}

	// Ensure exchange exists
	if exchange != "" {
		err := r.channel.ExchangeDeclare(
			exchange,
			"topic",
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event %s: %w", event.GetID(), err)
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		MessageId:    event.GetID(),
		Type:         string(event.GetType()),
		Body:         body,
	}

	err = r.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	r.logger.Debug("published event",
		slog.String("id", event.GetID()),
		slog.String("type", string(event.GetType())),
		slog.String("routing_key", routingKey),
	)

	return nil
}

// Subscribe binds a queue to an exchange and launches a message processing goroutine.
func (r *RabbitMQClient) Subscribe(ctx context.Context, queueName, exchange, routingKey string, handler HandlerFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel == nil || r.channel.IsClosed() {
		return ErrNotConnected
	}

	// Declare exchange if specified
	if exchange != "" {
		err := r.channel.ExchangeDeclare(
			exchange,
			"topic",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	// Declare durable queue
	q, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	// Bind queue to exchange
	if exchange != "" {
		err = r.channel.QueueBind(
			q.Name,
			routingKey,
			exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
	}

	// Set QoS prefetch count
	err = r.channel.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set channel QoS: %w", err)
	}

	deliveries, err := r.channel.Consume(
		q.Name,
		"",    // consumer tag
		false, // autoAck (manual ack for resilience)
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consumer on queue %s: %w", q.Name, err)
	}

	go r.processDeliveries(ctx, deliveries, handler, q.Name)

	r.logger.Info("subscribed to queue",
		slog.String("queue", q.Name),
		slog.String("exchange", exchange),
		slog.String("routing_key", routingKey),
	)

	return nil
}

func (r *RabbitMQClient) processDeliveries(ctx context.Context, deliveries <-chan amqp.Delivery, handler HandlerFunc, queueName string) {
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("stopping consumer due to context cancellation", slog.String("queue", queueName))
			return
		case d, ok := <-deliveries:
			if !ok {
				r.logger.Warn("delivery channel closed", slog.String("queue", queueName))
				return
			}

			if err := handler(ctx, d.Body); err != nil {
				r.logger.Error("handler error processing message",
					slog.String("queue", queueName),
					slog.String("msg_id", d.MessageId),
					slog.String("error", err.Error()),
				)
				// Requeue on transient errors or nack
				_ = d.Nack(false, true)
				continue
			}

			if err := d.Ack(false); err != nil {
				r.logger.Error("failed to ack message",
					slog.String("queue", queueName),
					slog.String("msg_id", d.MessageId),
					slog.String("error", err.Error()),
				)
			}
		}
	}
}

// Close gracefully closes the channel and connection.
func (r *RabbitMQClient) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	if r.channel != nil && !r.channel.IsClosed() {
		if err := r.channel.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.conn != nil && !r.conn.IsClosed() {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing rabbitmq client: %v", errs)
	}

	return nil
}
