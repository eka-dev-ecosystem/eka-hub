package examples

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"time"

	"golang/cache"
	"golang/config"
	"golang/pkg/logging"
	"golang/pkg/realtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/valyala/fasthttp"
)

// This example demonstrates how to use the realtime WebSocket and SSE features.
// It is not a runnable program; copy it into cmd/ or an app package to execute.
func RunRealtimeExample() {
	// Load config and initialize Redis
	cfg := config.Load()
	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	defer redisCache.Close()

	// Create and start the hub. The Redis bridge fans messages out across
	// multiple server instances.
	redisBridge, err := realtime.NewRedisBridge(redisCache, logging.Logger())
	if err != nil {
		log.Fatal(err)
	}
	defer redisBridge.Close()
	hub := realtime.NewHub(logging.Logger(), redisBridge)
	go hub.Run()
	defer hub.Shutdown()

	app := fiber.New()

	// WebSocket endpoint
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		handleWebSocket(c, hub)
	}))

	// SSE endpoint
	app.Get("/sse", func(c *fiber.Ctx) error {
		return handleSSE(c, hub)
	})

	// HTTP endpoint to broadcast messages
	app.Post("/broadcast", func(c *fiber.Ctx) error {
		var req struct {
			Channel string      `json:"channel"`
			Type    string      `json:"type"`
			Data    interface{} `json:"data"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		hub.Broadcast(&realtime.Message{
			Type:    req.Type,
			Channel: req.Channel,
			Data:    req.Data,
		})

		return c.JSON(fiber.Map{"status": "sent"})
	})

	// Simulate periodic broadcasts
	go simulateBroadcasts(hub)

	log.Println("Server starting on :3000")
	log.Fatal(app.Listen(":3000"))
}

func handleWebSocket(c *websocket.Conn, hub *realtime.Hub) {
	client := &realtime.Client{
		ID:      "ws-" + time.Now().Format("20060102150405"),
		UserID:  c.Query("user_id", "anonymous"),
		Channel: make(chan []byte, 256),
		Type:    "websocket",
	}

	hub.Register(client)
	defer hub.Unregister(client)

	// Subscribe to a default channel
	hub.SubscribeToChannel(client.ID, "general")

	// Read messages from client
	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			log.Printf("Received from client: %s", msg)
		}
	}()

	// Send messages to client
	for message := range client.Channel {
		if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

func handleSSE(c *fiber.Ctx, hub *realtime.Hub) error {
	client := &realtime.Client{
		ID:      "sse-" + time.Now().Format("20060102150405"),
		UserID:  c.Query("user_id", "anonymous"),
		Channel: make(chan []byte, 256),
		Type:    "sse",
	}

	hub.Register(client)
	defer hub.Unregister(client)

	// Subscribe to a default channel
	hub.SubscribeToChannel(client.ID, "notifications")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// Send initial connection event
	c.WriteString("event: connected\ndata: {\"status\":\"connected\"}\n\n")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ctx, cancel := context.WithCancel(c.Context())
		defer cancel()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case message, ok := <-client.Channel:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: message\n")
				fmt.Fprintf(w, "data: %s\n\n", message)
				_ = w.Flush()

			case <-ticker.C:
				fmt.Fprintf(w, ": keep-alive\n\n")
				_ = w.Flush()

			case <-ctx.Done():
				return
			}
		}
	}))

	return nil
}

func simulateBroadcasts(hub *realtime.Hub) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	counter := 0
	for range ticker.C {
		counter++

		// Broadcast to general channel
		hub.Broadcast(&realtime.Message{
			Type:    "update",
			Channel: "general",
			Data: fiber.Map{
				"message": "Periodic update",
				"counter": counter,
				"time":    time.Now().Format(time.RFC3339),
			},
		})

		// Broadcast to notifications channel
		if counter%3 == 0 {
			hub.Broadcast(&realtime.Message{
				Type:    "notification",
				Channel: "notifications",
				Data: fiber.Map{
					"title": "New Notification",
					"body":  "This is a test notification",
					"count": counter / 3,
				},
			})
		}

		log.Printf("Broadcasted update #%d", counter)
	}
}
