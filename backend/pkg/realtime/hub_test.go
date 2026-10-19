package realtime

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"golang/pkg/logging"
)

func startHub(t *testing.T) *Hub {
	t.Helper()
	hub := NewHub(logging.Logger())
	go hub.Run()
	t.Cleanup(hub.Shutdown)
	return hub
}

func newTestClient(id string) *Client {
	return &Client{
		ID:      id,
		UserID:  "user-" + id,
		Channel: make(chan []byte, 10),
		Type:    "sse",
	}
}

func waitFor(t *testing.T, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

// waitClients blocks until the hub reports n registered clients.
func waitClients(t *testing.T, hub *Hub, n int) {
	t.Helper()
	waitFor(t, func() bool {
		total, _ := hub.GetStats()["total_clients"].(int)
		return total == n
	}, fmt.Sprintf("%d registered clients", n))
}

func TestHubBroadcastToAll(t *testing.T) {
	hub := startHub(t)
	c := newTestClient("c1")
	hub.Register(c)
	waitClients(t, hub, 1)

	hub.Broadcast(&Message{Type: "hello", Data: "world"})

	select {
	case raw := <-c.Channel:
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("bad message: %v", err)
		}
		if msg.Type != "hello" || msg.Data != "world" {
			t.Errorf("unexpected message: %#v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("client did not receive broadcast")
	}
}

func TestHubBroadcastToChannelOnly(t *testing.T) {
	hub := startHub(t)
	c1 := newTestClient("c1")
	c2 := newTestClient("c2")
	hub.Register(c1)
	hub.Register(c2)
	waitClients(t, hub, 2)

	hub.SubscribeToChannel("c1", "sports")
	hub.SubscribeToChannel("c2", "finance")

	hub.Broadcast(&Message{Type: "score", Channel: "sports", Data: "1-0"})

	select {
	case raw := <-c1.Channel:
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("bad message: %v", err)
		}
		if msg.Channel != "sports" {
			t.Errorf("expected sports channel, got %#v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("subscribed client did not receive")
	}

	select {
	case <-c2.Channel:
		t.Error("c2 must not receive messages from another channel")
	default:
	}
}

func TestHubUserScopedDelivery(t *testing.T) {
	hub := startHub(t)
	client := newTestClient("c1")
	client.UserID = "alice"
	hub.Register(client)
	waitClients(t, hub, 1)

	hub.Broadcast(&Message{Type: "dm", UserID: "bob", Data: "secret-bob"})
	hub.Broadcast(&Message{Type: "dm", UserID: "alice", Data: "secret-alice"})

	select {
	case raw := <-client.Channel:
		var msg Message
		_ = json.Unmarshal(raw, &msg)
		if msg.Data != "secret-alice" {
			t.Errorf("expected alice message, got %#v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("alice did not receive her message")
	}
}

func TestHubUnregisterClosesChannel(t *testing.T) {
	hub := startHub(t)
	c := newTestClient("c1")
	hub.Register(c)
	waitClients(t, hub, 1)
	hub.Unregister(c)
	select {
	case _, ok := <-c.Channel:
		if ok {
			t.Error("channel should be closed after unregister")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected channel to close")
	}
}

func TestHubStats(t *testing.T) {
	hub := startHub(t)
	hub.Register(newTestClient("c1"))
	hub.Register(newTestClient("c2"))
	waitClients(t, hub, 2)

	stats := hub.GetStats()
	if stats["total_clients"] != 2 {
		t.Errorf("expected 2 clients, got %#v", stats)
	}
}
