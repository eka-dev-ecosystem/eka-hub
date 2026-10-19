# Real-Time Communication Quick Start

This guide will help you quickly get started with WebSocket and SSE (Server-Sent Events) in your Go application.

## 🚀 Quick Start

### 1. Start the Server

```bash
# Make sure Redis is running
docker-compose up -d redis

# Run the server
go run cmd/server/main.go
```

The server will start on `http://localhost:3000` with the following real-time endpoints:

- **WebSocket**: `ws://localhost:3000/api/v1/realtime/ws/connect`
- **SSE**: `http://localhost:3000/api/v1/realtime/sse/events`

### 2. Test WebSocket (Option 1: HTML Client)

Open `examples/websocket_client.html` in your browser:

```bash
# On Windows
start examples/websocket_client.html

# On Mac/Linux
open examples/websocket_client.html
```

Features:

- ✅ Real-time connection status
- ✅ Subscribe/unsubscribe from channels
- ✅ Send and receive messages
- ✅ Live statistics (messages sent/received)
- ✅ Connection uptime tracker

### 3. Test SSE (Option 2: HTML Client)

Open `examples/sse_client.html` in your browser:

```bash
# On Windows
start examples/sse_client.html

# On Mac/Linux
open examples/sse_client.html
```

Features:

- ✅ Real-time event streaming
- ✅ Automatic reconnection
- ✅ Broadcasting to channels
- ✅ User-specific messaging
- ✅ Auto-scroll with toggle

### 4. Test with cURL

#### Connect to SSE Stream

```bash
curl -N http://localhost:3000/api/v1/realtime/sse/events
```

#### Broadcast a Message (requires JWT token)

```bash
curl -X POST http://localhost:3000/api/v1/realtime/sse/broadcast \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "channel": "notifications",
    "type": "alert",
    "data": {
      "message": "System maintenance in 10 minutes",
      "priority": "high"
    }
  }'
```

#### Send Message to Specific User

```bash
curl -X POST http://localhost:3000/api/v1/realtime/sse/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "user_id": "user-123",
    "type": "notification",
    "data": {
      "title": "Welcome!",
      "body": "Thanks for joining"
    }
  }'
```

#### Get Statistics

```bash
curl http://localhost:3000/api/v1/realtime/ws/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📝 Code Examples

### Broadcasting from Your Backend

```go
// In your handler
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "golang/pkg/handlers"
)

type NotificationHandler struct {
    wsHandler *handlers.WebSocketHandler
}

// Send notification to all users in a channel
func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
    h.wsHandler.BroadcastToChannel("notifications", "notification", fiber.Map{
        "title": "New Feature Released!",
        "body":  "Check out our new dark mode",
        "icon":  "🎉",
    })
    
    return c.JSON(fiber.Map{"status": "sent"})
}

// Send private message to specific user
func (h *NotificationHandler) SendPrivateMessage(c *fiber.Ctx) error {
    userID := c.Params("userId")
    
    h.wsHandler.SendToUser(userID, "private_message", fiber.Map{
        "from":    "admin",
        "message": "Your account has been verified",
    })
    
    return c.JSON(fiber.Map{"status": "sent"})
}
```

### WebSocket Client (JavaScript)

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:3000/api/v1/realtime/ws/connect');

// Connection opened
ws.addEventListener('open', () => {
    console.log('✅ Connected to WebSocket');
    
    // Subscribe to notifications channel
    ws.send(JSON.stringify({
        action: 'subscribe',
        channel: 'notifications'
    }));
});

// Listen for messages
ws.addEventListener('message', (event) => {
    const message = JSON.parse(event.data);
    console.log('📨 Received:', message);
    
    // Handle different message types
    if (message.type === 'notification') {
        showNotification(message.data.title, message.data.body);
    }
});

// Send a message
function sendMessage(text) {
    ws.send(JSON.stringify({
        action: 'message',
        channel: 'chat',
        data: { text, timestamp: Date.now() }
    }));
}
```

### SSE Client (JavaScript)

```javascript
// Connect to SSE
const eventSource = new EventSource('http://localhost:3000/api/v1/realtime/sse/events');

// Listen for connected event
eventSource.addEventListener('connected', (e) => {
    const data = JSON.parse(e.data);
    console.log('✅ Connected with client ID:', data.client_id);
});

// Listen for notifications
eventSource.addEventListener('notification', (e) => {
    const data = JSON.parse(e.data);
    showNotification(data.title, data.body);
});

// Listen for any message
eventSource.addEventListener('message', (e) => {
    const data = JSON.parse(e.data);
    console.log('📨 Message:', data);
});

// Error handling
eventSource.addEventListener('error', (error) => {
    console.error('❌ SSE error:', error);
});
```

### React Hook Example

```javascript
import { useEffect, useState } from 'react';

function useWebSocket(url) {
    const [ws, setWs] = useState(null);
    const [messages, setMessages] = useState([]);
    const [connected, setConnected] = useState(false);

    useEffect(() => {
        const websocket = new WebSocket(url);

        websocket.addEventListener('open', () => {
            setConnected(true);
            console.log('Connected to WebSocket');
        });

        websocket.addEventListener('message', (event) => {
            const message = JSON.parse(event.data);
            setMessages(prev => [...prev, message]);
        });

        websocket.addEventListener('close', () => {
            setConnected(false);
            console.log('Disconnected from WebSocket');
        });

        setWs(websocket);

        return () => {
            websocket.close();
        };
    }, [url]);

    const sendMessage = (action, channel, data) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ action, channel, data }));
        }
    };

    return { ws, messages, connected, sendMessage };
}

// Usage
function ChatComponent() {
    const { messages, connected, sendMessage } = useWebSocket(
        'ws://localhost:3000/api/v1/realtime/ws/connect'
    );

    const handleSend = () => {
        sendMessage('message', 'chat', { text: 'Hello!' });
    };

    return (
        <div>
            <p>Status: {connected ? '🟢 Connected' : '🔴 Disconnected'}</p>
            <button onClick={handleSend}>Send Message</button>
            <ul>
                {messages.map((msg, i) => (
                    <li key={i}>{JSON.stringify(msg)}</li>
                ))}
            </ul>
        </div>
    );
}
```

## 🎯 Common Use Cases

### 1. Real-time Chat Application

```go
// Subscribe to chat room
ws.send(JSON.stringify({
    action: 'subscribe',
    channel: 'chat-room-' + roomId
}));

// Send chat message
ws.send(JSON.stringify({
    action: 'message',
    channel: 'chat-room-' + roomId,
    data: {
        user: 'John',
        text: 'Hello everyone!',
        timestamp: Date.now()
    }
}));
```

### 2. Live Notifications

```go
// Backend: Send notification
wsHandler.BroadcastToChannel("notifications", "notification", fiber.Map{
    "title": "New comment on your post",
    "body":  "John commented: Nice work!",
    "link":  "/posts/123",
})

// Frontend: Listen for notifications
eventSource.addEventListener('notification', (e) => {
    const notification = JSON.parse(e.data);
    new Notification(notification.title, { body: notification.body });
});
```

### 3. Live Dashboard Updates

```go
// Backend: Broadcast metrics every second
ticker := time.NewTicker(1 * time.Second)
go func() {
    for range ticker.C {
        wsHandler.BroadcastToChannel("dashboard", "metrics", fiber.Map{
            "cpu":    getCPUUsage(),
            "memory": getMemoryUsage(),
            "users":  getActiveUsers(),
        })
    }
}()

// Frontend: Update dashboard
ws.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    if (data.channel === 'dashboard') {
        updateDashboard(data.data);
    }
});
```

### 4. Presence System (Who's Online)

```go
// Backend: Track user presence
func (h *PresenceHandler) UserConnected(userID string) {
    h.wsHandler.BroadcastToChannel("presence", "user_online", fiber.Map{
        "user_id": userID,
        "status":  "online",
    })
}

func (h *PresenceHandler) UserDisconnected(userID string) {
    h.wsHandler.BroadcastToChannel("presence", "user_offline", fiber.Map{
        "user_id": userID,
        "status":  "offline",
    })
}

// Frontend: Track online users
const onlineUsers = new Set();

ws.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'user_online') {
        onlineUsers.add(data.data.user_id);
    } else if (data.type === 'user_offline') {
        onlineUsers.delete(data.data.user_id);
    }
    updateOnlineUsersList(onlineUsers);
});
```

## 🔧 Configuration

### Environment Variables

```bash
# Redis (required for scalability)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT (for authenticated endpoints)
JWT_ACCESS_SECRET=your-access-secret
JWT_REFRESH_SECRET=your-refresh-secret
```

### Docker Compose

```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - redis

volumes:
  redis_data:
```

## 🚀 Production Deployment

### Nginx Configuration

```nginx
# WebSocket support
location /api/v1/realtime/ws/ {
    proxy_pass http://backend:3000;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_read_timeout 86400;
}

# SSE support
location /api/v1/realtime/sse/ {
    proxy_pass http://backend:3000;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_buffering off;
    proxy_cache off;
    chunked_transfer_encoding off;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

### Load Balancing

The system supports horizontal scaling through Redis Pub/Sub. You can run multiple instances:

```bash
# Instance 1
PORT=3000 go run cmd/server/main.go

# Instance 2
PORT=3001 go run cmd/server/main.go

# Instance 3
PORT=3002 go run cmd/server/main.go
```

All instances will share messages through Redis, allowing clients to connect to any instance.

## 📊 Monitoring

### Get Real-time Statistics

```bash
curl http://localhost:3000/api/v1/realtime/ws/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:

```json
{
  "message": "Statistics retrieved",
  "success": true,
  "data": {
    "total_clients": 150,
    "channels": {
      "notifications": 85,
      "chat-room-1": 45,
      "dashboard": 20
    }
  },
  "code": "SUCCESS",
  "timestamp": "2026-01-18T10:30:00Z"
}
```

## 🐛 Troubleshooting

### WebSocket Connection Fails

**Problem**: Cannot connect to WebSocket
**Solution**:

- Check if server is running: `curl http://localhost:3000/health`
- Verify WebSocket URL: `ws://localhost:3000/api/v1/realtime/ws/connect`
- Check browser console for errors

### SSE Events Not Received

**Problem**: Not receiving SSE events
**Solution**:

- Check SSE connection: `curl -N http://localhost:3000/api/v1/realtime/sse/events`
- Verify Content-Type is `text/event-stream`
- Check if nginx/proxy is buffering responses (add `X-Accel-Buffering: no`)

### Messages Not Broadcasting

**Problem**: Messages sent but not received
**Solution**:

- Verify Redis is running: `redis-cli ping`
- Check if client is subscribed to the correct channel
- Review server logs for errors

### High Memory Usage

**Problem**: Server memory growing over time
**Solution**:

- Check number of connected clients: GET `/api/v1/realtime/ws/stats`
- Verify clients are properly disconnecting
- Consider implementing connection limits

## 📚 Next Steps

1. ✅ Add authentication to WebSocket connections
2. ✅ Implement message persistence with database
3. ✅ Add rate limiting to prevent abuse
4. ✅ Implement typing indicators for chat
5. ✅ Add read receipts for messages
6. ✅ Create admin dashboard for monitoring
7. ✅ Add message history API
8. ✅ Implement reconnection with message replay

## 🔗 Related Documentation

- [Full Documentation](REALTIME_SETUP.md)
- [JWT Authentication](JWT_COMPLETE.md)
- [Repository Pattern](REPOSITORY_COMPLETE.md)
- [Error Handling](ERROR_HANDLING_COMPLETE.md)

## 💡 Tips

- Use SSE for **server-to-client** streaming (simpler, auto-reconnect)
- Use WebSocket for **bidirectional** communication (chat, gaming)
- Always subscribe to specific channels instead of broadcasting to all
- Implement heartbeat/ping-pong to detect dead connections
- Use Redis for production to enable horizontal scaling
- Set appropriate timeout values based on your use case
