package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"golang/pkg/logging"

	"github.com/google/uuid"
)

// Message represents a real-time message
type Message struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
	UserID  string      `json:"user_id,omitempty"`
	// ID is the originating instance's publication ID used to trip the local
	// echo of a message this instance itself published.
	ID string `json:"id,omitempty"`
}

// Client represents a connected client
type Client struct {
	ID        string
	UserID    string
	Channel   chan []byte
	Type      string // "websocket" or "sse"
	closeOnce sync.Once
}

// Close closes the client channel safely at most once
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.Channel)
	})
}

// Hub manages all client connections and message broadcasting
type Hub struct {
	id         string
	clients    map[string]*Client
	channels   map[string]map[string]*Client // channel -> clientID -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	bridges    []Bridge // distributed fan-out transports (may be empty)
	logger     *slog.Logger
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewHub creates a new Hub instance. The bridges (Redis and/or RabbitMQ,
// picked through REALTIME_BRIDGE) fan messages out across a fleet of
// instances; when no bridge is given the hub runs in single-instance
// (local-only) mode.
func NewHub(logger *slog.Logger, bridges ...Bridge) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = logging.Logger()
	}
	return &Hub{
		id:         uuid.NewString(),
		clients:    make(map[string]*Client),
		channels:   make(map[string]map[string]*Client),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		broadcast:  make(chan *Message, 1000),
		bridges:    bridges,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Listen for messages originating on other instances.
	for _, bridge := range h.bridges {
		go h.listenBridge(bridge)
	}

	for {
		select {
		case <-h.ctx.Done():
			return

		case client := <-h.register:
			h.Register(client)

		case client := <-h.unregister:
			h.Unregister(client)

		case message := <-h.broadcast:
			h.publishToBridges(message)
			h.broadcastLocal(message)
		}
	}
}

// publishToBridges forwards a message to every configured bridge so other
// instances can fan it out to their own clients. The outbound copy is stamped
// with the origin instance ID so listenBridge can drop it when it comes back.
func (h *Hub) publishToBridges(message *Message) {
	for _, bridge := range h.bridges {
		msgCopy := &Message{
			Type:    message.Type,
			Channel: message.Channel,
			Data:    message.Data,
			UserID:  message.UserID,
			ID:      h.id,
		}
		go func(b Bridge, m *Message) {
			if err := b.Publish(h.ctx, m); err != nil {
				h.logger.Error("hub bridge publish failed",
					slog.String("bridge", b.Name()),
					logging.Err(err),
				)
			}
		}(bridge, msgCopy)
	}
}

// listenBridge drains the bridge's incoming channel and rebroadcasts messages
// that arrived from other instances locally. Echoes of the messages this
// instance itself published (detected via ID) are dropped.
func (h *Hub) listenBridge(bridge Bridge) {
	h.logger.Info("hub listening to bridge", slog.String("bridge", bridge.Name()))
	for {
		select {
		case <-h.ctx.Done():
			_ = bridge.Close()
			return
		case message, ok := <-bridge.Incoming():
			if !ok {
				return
			}
			if message.ID == h.id {
				continue
			}
			h.broadcastLocal(message)
		}
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
	h.logger.Debug("hub client registered",
		slog.String("client_id", client.ID),
		slog.String("type", client.Type),
		slog.String("user_id", client.UserID),
	)
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		client.Close()
		h.logger.Debug("hub client unregistered", slog.String("client_id", client.ID))
	}
	// Remove from all channels
	for channelName, clients := range h.channels {
		if _, ok := clients[client.ID]; ok {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(h.channels, channelName)
			}
		}
	}
}

// Broadcast sends a message to all clients or specific channel
func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

// SubscribeToChannel subscribes a client to a specific channel
func (h *Hub) SubscribeToChannel(clientID, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[clientID]
	if !ok {
		return
	}

	if h.channels[channel] == nil {
		h.channels[channel] = make(map[string]*Client)
	}
	h.channels[channel][clientID] = client
	h.logger.Debug("hub client subscribed to channel", slog.String("client_id", clientID), slog.String("channel", channel))
}

// UnsubscribeFromChannel unsubscribes a client from a specific channel
func (h *Hub) UnsubscribeFromChannel(clientID, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.channels[channel]; ok {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(h.channels, channel)
		}
		h.logger.Debug("hub client unsubscribed from channel", slog.String("client_id", clientID), slog.String("channel", channel))
	}
}

// broadcastLocal broadcasts a message to local clients
func (h *Hub) broadcastLocal(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("hub marshal failed", logging.Err(err))
		return
	}

	// Broadcast to specific channel
	if message.Channel != "" {
		if clients, ok := h.channels[message.Channel]; ok {
			for _, client := range clients {
				// If message has a specific UserID, only send to that user
				if message.UserID != "" && client.UserID != message.UserID {
					continue
				}
				select {
				case client.Channel <- data:
				default:
					// Client channel is full or closed, unregister it
					go h.Unregister(client)
				}
			}
		}
		return
	}

	// Broadcast to all clients if no channel specified
	for _, client := range h.clients {
		// If message has a specific UserID, only send to that user
		if message.UserID != "" && client.UserID != message.UserID {
			continue
		}
		select {
		case client.Channel <- data:
		default:
			// Client channel is full or closed, unregister it
			go h.Unregister(client)
		}
	}
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	channelStats := make(map[string]int)
	for channel, clients := range h.channels {
		channelStats[channel] = len(clients)
	}

	return map[string]interface{}{
		"total_clients": len(h.clients),
		"channels":      channelStats,
	}
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	h.logger.Info("hub shutting down")
	h.cancel()

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, client := range h.clients {
		client.Close()
	}
	h.clients = make(map[string]*Client)
	h.channels = make(map[string]map[string]*Client)
}
