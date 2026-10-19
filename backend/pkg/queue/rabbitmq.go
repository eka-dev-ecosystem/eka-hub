package queue

import (
	"context"
	"fmt"
	"log/slog"

	"golang/config"
	"golang/pkg/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ is a concrete Broker implementation backed by RabbitMQ.
type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	logger     *slog.Logger
}

// NewRabbitMQ creates a new RabbitMQ connection and channel.
// The logger is optional; when nil the package-wide logger is used.
func NewRabbitMQ(cfg *config.RabbitMQConfig, logger *slog.Logger) (*RabbitMQ, error) {
	if logger == nil {
		logger = logging.Logger()
	}

	conn, err := amqp.Dial(cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	logger.Info("connected to RabbitMQ",
		slog.String("host", cfg.Host),
		slog.String("vhost", cfg.Vhost),
	)

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
		logger:     logger,
	}, nil
}

// DeclareQueue declares a queue and returns its metadata.
func (r *RabbitMQ) DeclareQueue(name string, durable, autoDelete, exclusive bool) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		name,       // name
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		false,      // no-wait
		nil,        // arguments
	)
}

// DeclareExchange declares an exchange.
func (r *RabbitMQ) DeclareExchange(name, kind string, durable, autoDelete bool) error {
	return r.Channel.ExchangeDeclare(
		name,       // name
		kind,       // type (direct, fanout, topic, headers)
		durable,    // durable
		autoDelete, // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
}

// BindQueue binds a queue to an exchange with the given routing key.
func (r *RabbitMQ) BindQueue(queueName, routingKey, exchangeName string) error {
	return r.Channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
}

// Publish publishes a message to an exchange with the given routing key.
func (r *RabbitMQ) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return r.Channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// PublishToQueue publishes a message directly to a queue.
func (r *RabbitMQ) PublishToQueue(ctx context.Context, queueName string, body []byte) error {
	return r.Publish(ctx, "", queueName, body)
}

// Consume starts consuming messages from a queue.
func (r *RabbitMQ) Consume(ctx context.Context, queueName, consumerName string, autoAck bool) (<-chan amqp.Delivery, error) {
	return r.Channel.ConsumeWithContext(
		ctx,
		queueName,    // queue
		consumerName, // consumer
		autoAck,      // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
}

// HealthCheck reports whether the RabbitMQ connection is alive.
func (r *RabbitMQ) HealthCheck() error {
	if r.Connection.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	return nil
}

// Close closes the RabbitMQ connection and channel.
func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return err
	}
	return r.Connection.Close()
}
