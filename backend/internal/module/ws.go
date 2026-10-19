package module

import (
	"log/slog"
	"time"

	"golang/pkg/logging"
	"golang/pkg/realtime"
	"golang/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// WSHandler exposes the WebSocket endpoints backed by the realtime hub.
type WSHandler struct {
	hub    *realtime.Hub
	logger *slog.Logger
}

// NewWSHandler builds the WebSocket handler.
func NewWSHandler(hub *realtime.Hub) *WSHandler {
	return &WSHandler{hub: hub, logger: logging.Logger().With("module", "websocket")}
}

// upgradeWS guards the WebSocket endpoint with an upgrade check.
func upgradeWS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// Handle accepts a WebSocket connection and pumps messages.
func (h *WSHandler) Handle(conn *websocket.Conn) {
	clientID := uuid.New().String()
	userID, _ := conn.Locals("user_id").(string)
	if userID == "" {
		if uid := conn.Locals("userID"); uid != nil {
			userID, _ = uid.(string)
		}
	}

	client := &realtime.Client{
		ID:      clientID,
		UserID:  userID,
		Channel: make(chan []byte, 256),
		Type:    "websocket",
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	go h.writePump(conn, client)
	h.readPump(conn, client)
}

func (h *WSHandler) readPump(conn *websocket.Conn, client *realtime.Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Warn("websocket read error", "client_id", client.ID, logging.Err(err))
			}
			return
		}
		h.handleClientMessage(client, raw)
	}
}

func (h *WSHandler) writePump(conn *websocket.Conn, client *realtime.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Channel:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Warn("websocket write error", "client_id", client.ID, logging.Err(err))
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHandler) handleClientMessage(client *realtime.Client, raw []byte) {
	var message map[string]interface{}
	if err := jsonUnmarshal(raw, &message); err != nil {
		h.logger.Warn("websocket unmarshal failed", "client_id", client.ID, logging.Err(err))
		return
	}

	action, ok := message["action"].(string)
	if !ok {
		return
	}

	switch action {
	case "subscribe":
		if channel, ok := message["channel"].(string); ok {
			h.hub.SubscribeToChannel(client.ID, channel)
		}
	case "unsubscribe":
		if channel, ok := message["channel"].(string); ok {
			h.hub.UnsubscribeFromChannel(client.ID, channel)
		}
	case "message":
		channel, _ := message["channel"].(string)
		data := message["data"]
		h.hub.Broadcast(&realtime.Message{
			Type:    "message",
			Channel: channel,
			Data:    data,
			UserID:  client.UserID,
		})
	default:
		h.logger.Warn("websocket unknown action", "client_id", client.ID, "action", action)
	}
}

// GetStats returns hub statistics.
func (h *WSHandler) GetStats(c *fiber.Ctx) error {
	return response.SuccessResponse(c, h.hub.GetStats(), "Statistics retrieved")
}
