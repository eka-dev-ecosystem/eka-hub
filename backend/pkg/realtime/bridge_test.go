package realtime

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"golang/pkg/logging"
)

// fakeBridge implements Bridge in-memory so hub behavior around fan-out and
// echo suppression can be tested without any external service.
type fakeBridge struct {
	mu        sync.RWMutex
	name      string
	in        chan *Message
	published []*Message
}

func newFakeBridge(name string) *fakeBridge {
	return &fakeBridge{name: name, in: make(chan *Message, 16)}
}

func (b *fakeBridge) Name() string              { return b.name }
func (b *fakeBridge) Incoming() <-chan *Message { return b.in }
func (b *fakeBridge) Close() error              { return nil }
func (b *fakeBridge) Publish(_ context.Context, m *Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.published = append(b.published, m)
	return nil
}

func (b *fakeBridge) Published() []*Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]*Message, len(b.published))
	copy(out, b.published)
	return out
}

func (b *fakeBridge) PublishedCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.published)
}

func startHubWithBridge(t *testing.T) (*Hub, *fakeBridge, *fakeBridge) {
	t.Helper()
	redis := newFakeBridge("redis")
	rabbit := newFakeBridge("rabbitmq")

	hub := NewHub(logging.Logger(), redis, rabbit)
	go hub.Run()
	t.Cleanup(hub.Shutdown)

	return hub, redis, rabbit
}

func TestHubPublishesToEveryBridge(t *testing.T) {
	hub, redis, rabbit := startHubWithBridge(t)

	hub.Broadcast(&Message{Type: "ping", Data: "hello"})
	waitFor(t, func() bool { return redis.PublishedCount() > 0 && rabbit.PublishedCount() > 0 },
		"bridges to receive the broadcast")

	if got := redis.Published()[0].Type; got != "ping" {
		t.Errorf("redis bridge got type %q, want ping", got)
	}
	if got := rabbit.Published()[0].Type; got != "ping" {
		t.Errorf("rabbitmq bridge got type %q, want ping", got)
	}
}

func TestHubSuppressesOwnEcho(t *testing.T) {
	hub, redis, _ := startHubWithBridge(t)

	client := newTestClient("c1")
	hub.Register(client)
	defer hub.Unregister(client)
	waitFor(t, func() bool { return hub.GetStats()["total_clients"] == 1 }, "client registered")

	// Simulate the message traveling out and coming straight back on the
	// bridge, stamped with this hub's own ID.
	hub.Broadcast(&Message{Type: "tick"})
	waitFor(t, func() bool { return redis.PublishedCount() > 0 }, "outbound publish")

	echo := *redis.Published()[0]
	redis.in <- &echo

	// The broadcast + echo should deliver exactly one tick (echo suppressed).
	select {
	case raw := <-client.Channel:
		var m Message
		_ = json.Unmarshal(raw, &m)
		if m.Type != "tick" {
			t.Fatalf("unexpected message: %+v", m)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for local broadcast")
	}
	select {
	case <-client.Channel:
		t.Fatal("own echo must be suppressed, got a duplicate tick")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestHubForwardsIncomingFromOtherInstance(t *testing.T) {
	hub, redis, _ := startHubWithBridge(t)

	client := newTestClient("c2")
	hub.Register(client)
	defer hub.Unregister(client)
	waitFor(t, func() bool { return hub.GetStats()["total_clients"] == 1 }, "client registered")

	// A message originating on another instance (different ID) must be
	// delivered to local clients.
	other := &Message{Type: "tick", Data: "remote", ID: "some-other-hub-id"}
	redis.in <- other

	select {
	case raw := <-client.Channel:
		var m Message
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("marshal message: %v", err)
		}
		if m.Type != "tick" || m.ID != "some-other-hub-id" {
			t.Fatalf("unexpected forwarded message: %+v", m)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for forwarded message")
	}
}
