package queue

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Broker is the contract every message broker implementation must satisfy.
// Handlers and application code should depend on this interface instead of
// a concrete RabbitMQ type so the queue can be swapped or disabled entirely.
type Broker interface {
	// DeclareQueue declares a queue and returns its metadata.
	DeclareQueue(name string, durable, autoDelete, exclusive bool) (amqp.Queue, error)
	// DeclareExchange declares an exchange.
	DeclareExchange(name, kind string, durable, autoDelete bool) error
	// BindQueue binds a queue to an exchange with the given routing key.
	BindQueue(queueName, routingKey, exchangeName string) error
	// Publish publishes a message to an exchange with the given routing key.
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	// PublishToQueue publishes a message directly to a queue.
	PublishToQueue(ctx context.Context, queueName string, body []byte) error
	// Consume starts consuming messages from a queue.
	Consume(ctx context.Context, queueName, consumerName string, autoAck bool) (<-chan amqp.Delivery, error)
	// HealthCheck reports whether the broker connection is alive.
	HealthCheck() error
	// Close releases the broker resources.
	Close() error
}
