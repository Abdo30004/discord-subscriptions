package messaging

import (
	"context"

	"github.com/discord-subscriptions/shared/events"
)

// HandlerFunc processes raw message payloads received from a queue.
type HandlerFunc func(ctx context.Context, payload []byte) error

// Publisher defines the contract for emitting domain events to a message exchange.
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, event events.Event) error
}

// Subscriber defines the contract for consuming messages from queues.
type Subscriber interface {
	Subscribe(ctx context.Context, queueName, exchange, routingKey string, handler HandlerFunc) error
}

// Client defines the composite message broker interface including lifecycle management.
type Client interface {
	Publisher
	Subscriber
	Close() error
}
