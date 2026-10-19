package module

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang/pkg/constants"
	"golang/pkg/logging"
	"golang/pkg/realtime"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// jsonUnmarshal decodes JSON without discarding the error message handling.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// SSEHandler exposes Server-Sent Events endpoints.
type SSEHandler struct {
	hub    *realtime.Hub
	logger *slog.Logger
}

// NewSSEHandler builds the SSE handler.
func NewSSEHandler(hub *realtime.Hub) *SSEHandler {
	return &SSEHandler{hub: hub, logger: logging.Logger().With("module", "sse")}
}

// Events streams SSE events to a connected client.
func (h *SSEHandler) Events(c *fiber.Ctx) error {
	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = uuid.New().String()
	}
	userID, _ := c.Locals("user_id").(string)

	client := &realtime.Client{
		ID:      clientID,
		UserID:  userID,
		Channel: make(chan []byte, 256),
		Type:    "sse",
	}
	h.hub.Register(client)

	// Subscribe to any requested channels
	if channelsParam := c.Query("channels"); channelsParam != "" {
		for _, ch := range strings.Split(channelsParam, ",") {
			if trimmed := strings.TrimSpace(ch); trimmed != "" {
				h.hub.SubscribeToChannel(clientID, trimmed)
			}
		}
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Unregister MUST happen when the stream writer exits, not when Events returns!
		defer h.hub.Unregister(client)

		if err := h.sendEvent(w, "connected", map[string]string{
			"client_id": clientID,
			"timestamp": time.Now().Format(time.RFC3339),
		}); err != nil {
			h.logger.Warn("failed to send connect event", logging.Err(err))
			return
		}

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case message, ok := <-client.Channel:
				if !ok {
					return
				}
				var msg realtime.Message
				if err := json.Unmarshal(message, &msg); err != nil {
					h.logger.Warn("sse unmarshal failed", "client_id", clientID, logging.Err(err))
					continue
				}
				if err := h.sendEvent(w, msg.Type, msg.Data); err != nil {
					return
				}
			case <-ticker.C:
				fmt.Fprintf(w, ": keep-alive\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			case <-c.Context().Done():
				return
			}
		}
	}))

	return nil
}

func (h *SSEHandler) sendEvent(w *bufio.Writer, event string, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", raw)
	return w.Flush()
}

// Subscribe subscribes the client via query params.
func (h *SSEHandler) Subscribe(c *fiber.Ctx) error {
	channels := c.Query("channels")
	if channels == "" {
		return response.ErrorResponse(c, fiber.StatusBadRequest, "Channels parameter required", constants.CodeBadRequest)
	}
	return h.Events(c)
}

// BroadcastToChannel broadcasts a message via HTTP POST.
func (h *SSEHandler) BroadcastToChannel(c *fiber.Ctx) error {
	var req struct {
		Channel string      `json:"channel" validate:"required"`
		Type    string      `json:"type" validate:"required"`
		Data    interface{} `json:"data"`
		UserID  string      `json:"user_id,omitempty"`
	}
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	h.hub.Broadcast(&realtime.Message{
		Type:    req.Type,
		Channel: req.Channel,
		Data:    req.Data,
		UserID:  req.UserID,
	})

	return response.SuccessResponse(c, fiber.Map{"channel": req.Channel, "type": req.Type}, "Message broadcasted")
}

// SendToUser sends a message to a specific user via HTTP POST.
func (h *SSEHandler) SendToUser(c *fiber.Ctx) error {
	var req struct {
		UserID string      `json:"user_id" validate:"required"`
		Type   string      `json:"type" validate:"required"`
		Data   interface{} `json:"data"`
	}
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	h.hub.Broadcast(&realtime.Message{
		Type:   req.Type,
		Data:   req.Data,
		UserID: req.UserID,
	})

	return response.SuccessResponse(c, fiber.Map{"user_id": req.UserID, "type": req.Type}, "Message sent to user")
}

// GetStats returns SSE statistics.
func (h *SSEHandler) GetStats(c *fiber.Ctx) error {
	return response.SuccessResponse(c, h.hub.GetStats(), "Statistics retrieved")
}
