# Real-Time Communication Setup (WebSocket & SSE)

This document explains the scalable real-time communication setup using WebSocket and Server-Sent Events (SSE) with
Redis pub/sub for horizontal scaling.

## Architecture Overview

### Components

1. **Hub (`pkg/realtime/hub.go`)**: Central message broker that manages all client connections and message distribution
2. **WebSocket Handler (`pkg/handlers/websocket_handler.go`)**: Handles bidirectional WebSocket connections
3. **SSE Handler (`pkg/handlers/sse_handler.go`)**: Handles Server-Sent Events for server-to-client streaming
4. **Redis Pub/Sub**: Enables message broadcasting across multiple server instances for horizontal scaling

### Scalability Features

- **Redis Pub/Sub Integration**: Messages are published to Redis, allowing multiple server instances to broadcast to
  their connected clients
- **Channel-based Broadcasting**: Clients can subscribe to specific channels for targeted messaging
- **User-specific Messaging**: Send messages to specific users across all their connections
- **Connection Management**: Automatic client registration, unregistration, and cleanup
- **Keep-alive Mechanisms**: Automatic ping/pong for WebSocket and periodic keep-alive for SSE

## API Endpoints

### WebSocket Endpoints

#### 1. Connect to WebSocket

```
GET /api/v1/realtime/ws/connect
```

**Query Parameters:**

- `user_id` (optional): User identifier for authenticated connections

**Response:** WebSocket upgrade

**Client Messages:**

```javascript
// Subscribe to a channel
{
    "action"
:
    "subscribe",
        "channel"
:
    "chat-room-1"
}

// Unsubscribe from a channel
{
    "action"
:
    "unsubscribe",
        "channel"
:
    "chat-room-1"
}

// Send a message
{
    "action"
:
    "message",
        "channel"
:
    "chat-room-1",
        "data"
:
    {
        "text"
    :
        "Hello, World!"
    }
}
```

**Server Messages:**

```javascript
{
    "type"
:
    "message",
        "channel"
:
    "chat-room-1",
        "data"
:
    {
        "text"
    :
        "Hello, World!"
    }
,
    "user_id"
:
    "user-123"
}
```

#### 2. Get WebSocket Statistics (Protected)

```
GET /api/v1/realtime/ws/stats
Authorization: Bearer <access_token>
```

**Response:**

```json
{
  "message": "Statistics retrieved",
  "success": true,
  "data": {
    "total_clients": 42,
    "channels": {
      "chat-room-1": 15,
      "notifications": 27
    }
  },
  "code": "SUCCESS",
  "timestamp": "2026-01-18T10:30:00Z"
}
```

### Server-Sent Events (SSE) Endpoints

#### 1. Connect to SSE Stream

```
GET /api/v1/realtime/sse/events
```

**Response:** SSE stream

**Server Events:**

```
event: connected
data: {"client_id":"abc-123","timestamp":"2026-01-18T10:30:00Z"}

event: message
data: {"type":"notification","data":{"title":"New Message","body":"You have a new message"}}

event: user_update
data: {"user_id":"user-123","status":"online"}
```

#### 2. Subscribe to Specific Channels

```
GET /api/v1/realtime/sse/subscribe?channels=chat-room-1,notifications
```

**Query Parameters:**

- `channels`: Comma-separated list of channel names
- `client_id` (optional): Client identifier for reconnection

#### 3. Broadcast to Channel (Protected)

```
POST /api/v1/realtime/sse/broadcast
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "channel": "chat-room-1",
  "type": "message",
  "data": {
    "text": "Hello everyone!",
    "sender": "John Doe"
  },
  "user_id": "" // Optional: target specific user
}
```

**Response:**

```json
{
  "message": "Message broadcasted",
  "success": true,
  "data": {
    "channel": "chat-room-1",
    "type": "message"
  },
  "code": "SUCCESS",
  "timestamp": "2026-01-18T10:30:00Z"
}
```

#### 4. Send to Specific User (Protected)

```
POST /api/v1/realtime/sse/send
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "user_id": "user-123",
  "type": "notification",
  "data": {
    "title": "Private Message",
    "body": "You have been mentioned"
  }
}
```

#### 5. Get SSE Statistics (Protected)

```
GET /api/v1/realtime/sse/stats
Authorization: Bearer <access_token>
```

## Client Examples

### WebSocket Client (JavaScript)

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:3000/api/v1/realtime/ws/connect?user_id=user-123');

// Connection opened
ws.addEventListener('open', (event) => {
    console.log('Connected to WebSocket');

    // Subscribe to a channel
    ws.send(JSON.stringify({
        action: 'subscribe',
        channel: 'chat-room-1'
    }));
});

// Listen for messages
ws.addEventListener('message', (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);

    // Handle different message types
    switch (message.type) {
        case 'message':
            console.log('Chat message:', message.data);
            break;
        case 'notification':
            console.log('Notification:', message.data);
            break;
        default:
            console.log('Unknown message type:', message.type);
    }
});

// Send a message
function sendMessage(text) {
    ws.send(JSON.stringify({
        action: 'message',
        channel: 'chat-room-1',
        data: {text}
    }));
}

// Close connection
ws.addEventListener('close', (event) => {
    console.log('WebSocket closed');
});

// Error handling
ws.addEventListener('error', (error) => {
    console.error('WebSocket error:', error);
});
```

### SSE Client (JavaScript)

```javascript
// Connect to SSE
const eventSource = new EventSource('http://localhost:3000/api/v1/realtime/sse/events');

// Listen for connection event
eventSource.addEventListener('connected', (e) => {
    const data = JSON.parse(e.data);
    console.log('Connected with client ID:', data.client_id);
});

// Listen for custom events
eventSource.addEventListener('message', (e) => {
    const data = JSON.parse(e.data);
    console.log('Message received:', data);
});

eventSource.addEventListener('notification', (e) => {
    const data = JSON.parse(e.data);
    console.log('Notification:', data);
});

// Error handling
eventSource.addEventListener('error', (error) => {
    console.error('SSE error:', error);
    if (eventSource.readyState === EventSource.CLOSED) {
        console.log('SSE connection closed');
    }
});

// Close connection
function closeSSE() {
    eventSource.close();
}
```

### Broadcasting from Backend

```go
// In your handler or service
func (h *SomeHandler) NotifyUsers(c *fiber.Ctx) error {
// Broadcast to all users in a channel
h.wsHandler.BroadcastToChannel("notifications", "notification", fiber.Map{
"title": "System Update",
"body":  "System will be updated at 2 AM",
"priority": "high",
})

// Send to specific user
h.wsHandler.SendToUser("user-123", "private_message", fiber.Map{
"from": "admin",
"message": "Your account has been verified",
})

return c.JSON(fiber.Map{"status": "sent"})
}
```

## Use Cases

### 1. Real-time Chat Application

```javascript
// Client subscribes to chat room
ws.send(JSON.stringify({
    action: 'subscribe',
    channel: 'chat-room-1'
}));

// User sends message
ws.send(JSON.stringify({
    action: 'message',
    channel: 'chat-room-1',
    data: {
        text: 'Hello everyone!',
        timestamp: Date.now()
    }
}));
```

### 2. Live Notifications

```javascript
// Client connects to SSE for notifications
const eventSource = new EventSource('/api/v1/realtime/sse/events');

eventSource.addEventListener('notification', (e) => {
    const notification = JSON.parse(e.data);
    showNotification(notification.title, notification.body);
});
```

### 3. Live Dashboard Updates

```javascript
// Subscribe to dashboard data channel
ws.send(JSON.stringify({
    action: 'subscribe',
    channel: 'dashboard-metrics'
}));

// Receive real-time metrics
ws.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    if (data.channel === 'dashboard-metrics') {
        updateDashboard(data.data);
    }
});
```

### 4. Presence System

```javascript
// Track online users
ws.send(JSON.stringify({
    action: 'subscribe',
    channel: 'presence'
}));

ws.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'user_online') {
        addOnlineUser(data.data.user_id);
    } else if (data.type === 'user_offline') {
        removeOnlineUser(data.data.user_id);
    }
});
```

## Scaling Considerations

### Horizontal Scaling

- **Redis Pub/Sub**: All server instances subscribe to Redis channels
- When a message is published to Redis, all instances receive it and broadcast to their local clients
- No need for sticky sessions - users can connect to any server instance

### Load Balancing

```nginx
upstream websocket_backend {
    least_conn;
    server server1:3000;
    server server2:3000;
    server server3:3000;
}

server {
    location /api/v1/realtime/ws/ {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }
    
    location /api/v1/realtime/sse/ {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;
    }
}
```

### Performance Tips

1. **Connection Limits**: Set appropriate limits on client connections per server
2. **Message Buffering**: Client channels have buffer of 256 messages
3. **Redis Connection Pool**: Redis client handles connection pooling automatically
4. **Keep-alive Tuning**: Adjust ping/pong intervals based on your needs
5. **Graceful Shutdown**: The hub properly closes all connections on shutdown

## Monitoring

### Get Real-time Statistics

```bash
# Get current connection stats
curl -H "Authorization: Bearer <token>" \
  http://localhost:3000/api/v1/realtime/ws/stats

# Response
{
  "total_clients": 150,
  "channels": {
    "chat-room-1": 45,
    "notifications": 105
  }
}
```

### Redis Monitoring

```bash
# Monitor Redis pub/sub
redis-cli
> PUBSUB CHANNELS realtime:*
> PUBSUB NUMSUB realtime:channel:chat-room-1
```

## Security

### Authentication

- Use JWT middleware to protect broadcast endpoints
- Pass user_id through WebSocket query params after authentication
- Validate user permissions before allowing channel subscriptions

### Rate Limiting

```go
// Add rate limiting middleware to broadcast endpoints
import "github.com/gofiber/fiber/v2/middleware/limiter"

sse.Post("/broadcast",
limiter.New(limiter.Config{
Max: 10,
Expiration: 1 * time.Minute,
}),
authMiddleware.AuthMiddleware(jwtService),
sseHandler.BroadcastToChannel,
)
```

## Troubleshooting

### WebSocket Connection Issues

- Check if reverse proxy properly handles WebSocket upgrades
- Verify `Upgrade` and `Connection` headers are set
- Ensure firewall allows WebSocket connections

### SSE Connection Drops

- Disable buffering in reverse proxy (nginx, apache)
- Set appropriate timeouts
- Implement reconnection logic on client side

### Message Loss

- Messages are not persisted - implement message queue for critical messages
- Use acknowledgment system for important messages
- Consider using RabbitMQ for guaranteed delivery

## Next Steps

1. Implement authentication for WebSocket connections
2. Add message persistence using database or message queue
3. Implement room/channel permissions
4. Add typing indicators and read receipts for chat
5. Implement presence system (online/offline status)
6. Add message history API
7. Implement reconnection logic with message replay
