# Real-Time Examples

This directory contains examples and test clients for WebSocket and SSE functionality.

## Files

### HTML Test Clients

#### `websocket_client.html`

Interactive WebSocket client for testing real-time bidirectional communication.

**Features:**

- Connect/disconnect to WebSocket server
- Subscribe/unsubscribe from channels
- Send and receive messages
- Real-time statistics (messages sent/received, connection time)
- Message history with timestamps

**How to use:**

1. Start the server: `go run cmd/server/main.go`
2. Open the file in your browser: `examples/websocket_client.html`
3. Click "Connect" to establish WebSocket connection
4. Subscribe to channels and send messages

#### `sse_client.html`

Interactive SSE (Server-Sent Events) client for testing server-to-client streaming.

**Features:**

- Connect to SSE stream
- Receive real-time events
- Broadcast messages to channels (with authentication)
- Send messages to specific users
- Event history with auto-scroll
- Connection statistics

**How to use:**

1. Start the server: `go run cmd/server/main.go`
2. Open the file in your browser: `examples/sse_client.html`
3. Click "Connect to SSE" to start receiving events
4. (Optional) Add JWT token to send broadcasts

### Go Examples

#### `realtime_example.go`

Standalone example demonstrating WebSocket and SSE implementation.

**Features:**

- Complete WebSocket server setup
- Complete SSE server setup
- Automatic message broadcasting every 5 seconds
- HTTP endpoint to trigger broadcasts

**How to run:**

```bash
# Make sure Redis is running
docker-compose up -d redis

# Run the example
go run examples/realtime_example.go
```

Then open:

- WebSocket test: `ws://localhost:3000/ws`
- SSE test: `http://localhost:3000/sse`
- Trigger broadcast: `POST http://localhost:3000/broadcast`

#### `basic_usage.go`

Basic example showing database and cache usage.

#### `repository_pattern_example.go`

Example demonstrating the repository pattern implementation.

## Quick Test Scenarios

### Scenario 1: Real-time Chat

1. Open `websocket_client.html` in two browser tabs
2. In both tabs, connect to WebSocket
3. Subscribe both to channel "chat"
4. Send a message from one tab
5. See it appear in both tabs instantly

### Scenario 2: Live Notifications

1. Open `sse_client.html` in your browser
2. Connect to SSE
3. From another terminal, send a broadcast:

```bash
curl -X POST http://localhost:3000/api/v1/realtime/sse/broadcast \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "channel": "notifications",
    "type": "notification",
    "data": {"title": "Test", "body": "Hello!"}
  }'
```

4. See the notification appear in the browser

### Scenario 3: Multi-Channel Broadcasting

1. Open multiple tabs of `websocket_client.html`
2. Subscribe each tab to different channels:
    - Tab 1: "general"
    - Tab 2: "notifications"
    - Tab 3: "alerts"
3. Send messages to specific channels
4. Verify only subscribed tabs receive the messages

### Scenario 4: User-Specific Messages

1. Open `sse_client.html`
2. Connect with a specific user ID
3. Send a user-specific message:

```bash
curl -X POST http://localhost:3000/api/v1/realtime/sse/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "user_id": "user-123",
    "type": "private",
    "data": {"message": "This is for you only"}
  }'
```

4. Only the targeted user receives the message

## Testing with Browser DevTools

### WebSocket Testing

```javascript
// In browser console
const ws = new WebSocket('ws://localhost:3000/api/v1/realtime/ws/connect');

ws.onopen = () => {
    console.log('Connected');
    ws.send(JSON.stringify({
        action: 'subscribe',
        channel: 'test'
    }));
};

ws.onmessage = (e) => {
    console.log('Received:', JSON.parse(e.data));
};

// Send a message
ws.send(JSON.stringify({
    action: 'message',
    channel: 'test',
    data: {text: 'Hello from console!'}
}));
```

### SSE Testing

```javascript
// In browser console
const es = new EventSource('http://localhost:3000/api/v1/realtime/sse/events');

es.addEventListener('connected', (e) => {
    console.log('Connected:', JSON.parse(e.data));
});

es.addEventListener('message', (e) => {
    console.log('Message:', JSON.parse(e.data));
});

es.addEventListener('notification', (e) => {
    console.log('Notification:', JSON.parse(e.data));
});
```

## Performance Testing

### Load Testing WebSocket

```bash
# Install wscat
npm install -g wscat

# Connect and send messages
wscat -c ws://localhost:3000/api/v1/realtime/ws/connect
```

### Load Testing SSE

```bash
# Multiple concurrent connections
for i in {1..100}; do
  curl -N http://localhost:3000/api/v1/realtime/sse/events &
done

# Check statistics
curl http://localhost:3000/api/v1/realtime/ws/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Troubleshooting

### WebSocket Issues

**Problem**: Connection immediately closes

- Check server logs for errors
- Verify WebSocket upgrade headers
- Test with a simple wscat connection

**Problem**: Messages not received

- Verify subscription to correct channel
- Check server logs for broadcast events
- Test with browser DevTools

### SSE Issues

**Problem**: Connection times out

- Check nginx/proxy buffering settings
- Verify keep-alive is being sent
- Test with curl -N

**Problem**: Events not received

- Check event type matches listener
- Verify server is broadcasting to correct channel
- Check browser DevTools Network tab

## Additional Resources

- [Real-Time Quick Start Guide](../REALTIME_QUICKSTART.md)
- [Full Real-Time Documentation](../REALTIME_SETUP.md)
- [WebSocket MDN Docs](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
- [SSE MDN Docs](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
