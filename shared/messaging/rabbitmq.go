package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/discord-subscriptions/shared/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DefaultDLX          = "discord.events.dlx"
	DefaultDLQ          = "discord.events.dlq"
	DefaultDLQKey       = "dlq"
	MaxRetries          = 3
	HeaderRetryCount    = "x-retry-count"
	HeaderSchemaVersion = "x-event-version"
)

var (
	ErrNotConnected = errors.New("not connected to rabbitmq broker")
)

type subscriptionInfo struct {
	queueName  string
	exchange   string
	routingKey string
	handler    HandlerFunc
}

// RabbitMQClient implements the resilient Client interface for RabbitMQ.
type RabbitMQClient struct {
	mu            sync.RWMutex
	conn          *amqp.Connection
	pubChannel    *amqp.Channel
	subChannel    *amqp.Channel
	logger        *slog.Logger
	url           string
	isClosed      bool
	subscriptions []subscriptionInfo
	reconnectCh   chan struct{}
}

// NewRabbitMQClient establishes a resilient connection to RabbitMQ with background reconnection.
func NewRabbitMQClient(url string, logger *slog.Logger) (*RabbitMQClient, error) {
	if logger == nil {
		logger = slog.Default()
	}

	client := &RabbitMQClient{
		logger:      logger,
		url:         url,
		reconnectCh: make(chan struct{}, 1),
	}

	if err := client.connect(); err != nil {
		// Log warning and begin background reconnect attempts so service startup is non-blocking
		logger.Warn("initial rabbitmq connection failed, entering reconnection loop", slog.String("error", err.Error()))
		go client.reconnectLoop()
		return client, nil
	}

	go client.reconnectLoop()
	return client, nil
}

func (r *RabbitMQClient) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isClosed {
		return errors.New("client is closed")
	}

	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to dial rabbitmq: %w", err)
	}

	// 1. Separate dedicated publishing channel
	pubCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open publish channel: %w", err)
	}

	// 2. Separate dedicated subscribing channel
	subCh, err := conn.Channel()
	if err != nil {
		_ = pubCh.Close()
		_ = conn.Close()
		return fmt.Errorf("failed to open subscription channel: %w", err)
	}

	// 3. Declare Dead-Letter Exchange (DLX) and Dead-Letter Queue (DLQ)
	if err := r.setupDeadLetterInfrastructure(pubCh); err != nil {
		r.logger.Warn("failed to initialize dead-letter exchange/queue", slog.String("error", err.Error()))
	}

	r.conn = conn
	r.pubChannel = pubCh
	r.subChannel = subCh

	// Listen for unexpected connection loss
	closeNotify := conn.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		err, ok := <-closeNotify
		if ok && err != nil {
			r.logger.Warn("rabbitmq connection lost", slog.String("reason", err.Reason), slog.Int("code", err.Code))
			r.triggerReconnect()
		}
	}()

	r.logger.Info("connected to rabbitmq with separated pub/sub channels and DLX configured")
	return nil
}

func (r *RabbitMQClient) setupDeadLetterInfrastructure(ch *amqp.Channel) error {
	// Declare DLX topic exchange
	err := ch.ExchangeDeclare(
		DefaultDLX,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// Declare DLQ queue
	dlq, err := ch.QueueDeclare(
		DefaultDLQ,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind DLQ to DLX catching all dead-lettered messages
	err = ch.QueueBind(
		dlq.Name,
		"#",
		DefaultDLX,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ to DLX: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) triggerReconnect() {
	select {
	case r.reconnectCh <- struct{}{}:
	default:
	}
}

func (r *RabbitMQClient) reconnectLoop() {
	for range r.reconnectCh {
		r.mu.RLock()
		closed := r.isClosed
		r.mu.RUnlock()
		if closed {
			return
		}

		backoff := 1 * time.Second
		maxBackoff := 30 * time.Second

		for {
			r.mu.RLock()
			closed = r.isClosed
			r.mu.RUnlock()
			if closed {
				return
			}

			// Add jitter
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			sleepDuration := backoff + jitter
			r.logger.Info("attempting rabbitmq reconnection in...", slog.Duration("delay", sleepDuration))
			time.Sleep(sleepDuration)

			if err := r.connect(); err != nil {
				r.logger.Warn("rabbitmq reconnection failed", slog.String("error", err.Error()))
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}

			// Re-subscribe all registered consumers
			r.resubscribeAll()
			break
		}
	}
}

func (r *RabbitMQClient) resubscribeAll() {
	r.mu.RLock()
	subs := make([]subscriptionInfo, len(r.subscriptions))
	copy(subs, r.subscriptions)
	r.mu.RUnlock()

	for _, sub := range subs {
		r.logger.Info("re-establishing consumer subscription", slog.String("queue", sub.queueName), slog.String("routing_key", sub.routingKey))
		if err := r.bindAndConsume(context.Background(), sub); err != nil {
			r.logger.Error("failed to re-establish subscription", slog.String("queue", sub.queueName), slog.String("error", err.Error()))
		}
	}
}

// IsConnected reports whether the client has active publishing and subscribing channels.
func (r *RabbitMQClient) IsConnected() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn != nil && !r.conn.IsClosed() &&
		r.pubChannel != nil && !r.pubChannel.IsClosed() &&
		r.subChannel != nil && !r.subChannel.IsClosed()
}

// Publish serializes a domain event to JSON, attaches schema version headers, and publishes via pubChannel.
func (r *RabbitMQClient) Publish(ctx context.Context, exchange, routingKey string, event events.Event) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.pubChannel == nil || r.pubChannel.IsClosed() {
		return ErrNotConnected
	}

	// Ensure exchange exists
	if exchange != "" {
		err := r.pubChannel.ExchangeDeclare(
			exchange,
			"topic",
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event %s: %w", event.GetID(), err)
	}

	schemaVersion := event.GetSchemaVersion()
	if schemaVersion == "" {
		schemaVersion = events.CurrentSchemaVersion
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		MessageId:    event.GetID(),
		Type:         string(event.GetType()),
		Headers: amqp.Table{
			HeaderSchemaVersion: schemaVersion,
			HeaderRetryCount:    0,
		},
		Body: body,
	}

	err = r.pubChannel.PublishWithContext(
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
		slog.String("version", schemaVersion),
		slog.String("routing_key", routingKey),
	)

	return nil
}

// Subscribe binds a queue to an exchange with DLX configured and launches consumer goroutines.
func (r *RabbitMQClient) Subscribe(ctx context.Context, queueName, exchange, routingKey string, handler HandlerFunc) error {
	r.mu.Lock()
	sub := subscriptionInfo{
		queueName:  queueName,
		exchange:   exchange,
		routingKey: routingKey,
		handler:    handler,
	}
	r.subscriptions = append(r.subscriptions, sub)
	r.mu.Unlock()

	return r.bindAndConsume(ctx, sub)
}

func (r *RabbitMQClient) bindAndConsume(ctx context.Context, sub subscriptionInfo) error {
	r.mu.RLock()
	ch := r.subChannel
	r.mu.RUnlock()

	if ch == nil || ch.IsClosed() {
		return ErrNotConnected
	}

	// Declare exchange if specified
	if sub.exchange != "" {
		err := ch.ExchangeDeclare(
			sub.exchange,
			"topic",
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	// Queue configuration with Dead-Letter Exchange
	queueArgs := amqp.Table{
		"x-dead-letter-exchange":    DefaultDLX,
		"x-dead-letter-routing-key": DefaultDLQKey,
	}

	q, err := ch.QueueDeclare(
		sub.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		queueArgs,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s with DLX: %w", sub.queueName, err)
	}

	// Bind queue to exchange
	if sub.exchange != "" {
		err = ch.QueueBind(
			q.Name,
			sub.routingKey,
			sub.exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
	}

	// Set QoS prefetch count
	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("failed to set channel QoS: %w", err)
	}

	deliveries, err := ch.Consume(
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

	go r.processDeliveries(ctx, deliveries, sub.handler, q.Name, sub.exchange, sub.routingKey)

	r.logger.Info("subscribed to queue with DLX support",
		slog.String("queue", q.Name),
		slog.String("exchange", sub.exchange),
		slog.String("routing_key", sub.routingKey),
	)

	return nil
}

func (r *RabbitMQClient) processDeliveries(ctx context.Context, deliveries <-chan amqp.Delivery, handler HandlerFunc, queueName, exchange, routingKey string) {
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
				retryCount := getRetryCount(d.Headers)
				r.logger.Error("handler error processing message",
					slog.String("queue", queueName),
					slog.String("msg_id", d.MessageId),
					slog.Int("retry_count", retryCount),
					slog.String("error", err.Error()),
				)

				if retryCount < MaxRetries {
					// Transient failure: republish with incremented retry count after backoff
					go r.retryDelivery(ctx, d, retryCount+1, exchange, routingKey)
					_ = d.Ack(false)
				} else {
					// Max retries exceeded: Nack(requeue=false) routes directly to DLX
					r.logger.Error("poison message: max retries reached, dead-lettering to DLQ",
						slog.String("queue", queueName),
						slog.String("msg_id", d.MessageId),
						slog.String("dlx", DefaultDLX),
					)
					_ = d.Nack(false, false)
				}
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

func (r *RabbitMQClient) retryDelivery(ctx context.Context, d amqp.Delivery, newRetryCount int, exchange, routingKey string) {
	delay := time.Duration(newRetryCount) * 1 * time.Second
	time.Sleep(delay)

	r.mu.RLock()
	pubCh := r.pubChannel
	r.mu.RUnlock()

	if pubCh == nil || pubCh.IsClosed() {
		return
	}

	headers := d.Headers
	if headers == nil {
		headers = make(amqp.Table)
	}
	headers[HeaderRetryCount] = newRetryCount

	msg := amqp.Publishing{
		ContentType:  d.ContentType,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		MessageId:    d.MessageId,
		Type:         d.Type,
		Headers:      headers,
		Body:         d.Body,
	}

	_ = pubCh.PublishWithContext(ctx, exchange, routingKey, false, false, msg)
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	val, ok := headers[HeaderRetryCount]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

// Close gracefully closes channels and the connection.
func (r *RabbitMQClient) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.isClosed = true
	select {
	case r.reconnectCh <- struct{}{}:
	default:
	}

	var errs []error
	if r.pubChannel != nil && !r.pubChannel.IsClosed() {
		if err := r.pubChannel.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.subChannel != nil && !r.subChannel.IsClosed() {
		if err := r.subChannel.Close(); err != nil {
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
