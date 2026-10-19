package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"golang/cache"
	"golang/config"
	"golang/pkg/logging"
	queuepkg "golang/pkg/queue"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Bridge transport aliases. Define them once in config where the env var lives,
// and reuse them here so there is a single source of truth.
const (
	BridgeRedis  = config.BridgeRedis
	BridgeRabbit = config.BridgeRabbit
	BridgeBoth   = config.BridgeBoth
)

var (
	errNilRedis  = errors.New("realtime: nil redis cache")
	errNilBroker = errors.New("realtime: nil rabbitmq broker")
)

// Bridge transports messages between hub instances so any server in a fleet
// can broadcast to every connected client. It is transport-specific; the hub
// simply calls Publish for outbound traffic and drains Incoming for traffic
// originating on other instances.
type Bridge interface {
	// Name is the transport identifier used in logging and the health report.
	Name() string
	// Publish forwards a message to other instances.
	Publish(ctx context.Context, message *Message) error
	// Incoming yields messages received from other instances.
	Incoming() <-chan *Message
	// Close stops the transport goroutines.
	Close() error
}

// redisChannelName maps a realtime channel to the Redis pub/sub topic used to
// fan it out across instances.
func redisChannelName(channel string) string {
	if channel == "" {
		return "realtime:broadcast"
	}
	return "realtime:channel:" + channel
}

// RedisBridge fans out messages through Redis pub/sub using the same
// "realtime:*" topics the pre-bridge hub used, so it drops in as a drop-in
// replacement.
type RedisBridge struct {
	redis  *cache.RedisCache
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan *Message
	logger *slog.Logger
}

// NewRedisBridge builds a Redis-backed bridge. A non-nil Redis client is
// mandatory.
func NewRedisBridge(redis *cache.RedisCache, logger *slog.Logger) (*RedisBridge, error) {
	if redis == nil {
		return nil, errNilRedis
	}
	ctx, cancel := context.WithCancel(context.Background())
	b := &RedisBridge{
		redis:  redis,
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan *Message, 256),
		logger: logger,
	}
	go b.listen()
	return b, nil
}

func (b *RedisBridge) Name() string { return BridgeRedis }

func (b *RedisBridge) Publish(ctx context.Context, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return b.redis.Client.Publish(ctx, redisChannelName(message.Channel), payload).Err()
}

func (b *RedisBridge) Incoming() <-chan *Message { return b.ch }

func (b *RedisBridge) listen() {
	pubsub := b.redis.Client.PSubscribe(b.ctx, "realtime:*")
	defer func() { _ = pubsub.Close() }()

	deliveries := pubsub.Channel()
	for {
		select {
		case <-b.ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			var msg Message
			if err := json.Unmarshal([]byte(delivery.Payload), &msg); err != nil {
				b.logger.Warn("realtime redis bridge: unmarshal failed", logging.Err(err))
				continue
			}
			select {
			case b.ch <- &msg:
			default:
				b.logger.Warn("realtime redis bridge: incoming channel full, dropping message")
			}
		}
	}
}

func (b *RedisBridge) Close() error {
	b.cancel()
	return nil
}

// rabbitExchange is a single fanout exchange carrying every realtime message;
// each instance binds its own exclusive queue, so every instance receives every
// published message without competing consumers.
const rabbitExchange = "realtime.events"

// RabbitMQBridge fans out messages through a RabbitMQ fanout exchange. It is
// the preferred bridge for large fleets where Redis is not desirable.
type RabbitMQBridge struct {
	broker queuepkg.Broker
	queue  string
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan *Message
	logger *slog.Logger
}

// NewRabbitMQBridge declares the fanout exchange plus one exclusive,
// auto-delete queue per instance and starts consuming.
func NewRabbitMQBridge(broker queuepkg.Broker, logger *slog.Logger) (*RabbitMQBridge, error) {
	if broker == nil {
		return nil, errNilBroker
	}

	if err := broker.DeclareExchange(rabbitExchange, "fanout", true, false); err != nil {
		return nil, err
	}

	q, err := broker.DeclareQueue("realtime."+uuid.NewString(), true, true, true)
	if err != nil {
		return nil, err
	}
	if err := broker.BindQueue(q.Name, "", rabbitExchange); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	b := &RabbitMQBridge{
		broker: broker,
		queue:  q.Name,
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan *Message, 256),
		logger: logger,
	}

	deliveries, err := broker.Consume(ctx, q.Name, "realtime-bridge", false)
	if err != nil {
		cancel()
		return nil, err
	}
	go b.listen(deliveries)
	return b, nil
}

func (b *RabbitMQBridge) Name() string { return BridgeRabbit }

func (b *RabbitMQBridge) Publish(ctx context.Context, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return b.broker.Publish(ctx, rabbitExchange, "", payload)
}

func (b *RabbitMQBridge) Incoming() <-chan *Message { return b.ch }

func (b *RabbitMQBridge) listen(deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-b.ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}
			_ = delivery.Ack(false)
			var msg Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				b.logger.Warn("realtime rabbitmq bridge: unmarshal failed", logging.Err(err))
				continue
			}
			select {
			case b.ch <- &msg:
			default:
				b.logger.Warn("realtime rabbitmq bridge: incoming channel full, dropping message")
			}
		}
	}
}

func (b *RabbitMQBridge) Close() error {
	b.cancel()
	return nil
}
