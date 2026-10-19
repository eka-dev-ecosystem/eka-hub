# Quick Start Guide - Setup RabbitMQ, PostgreSQL, dan Redis

## 🚀 Cara Cepat Memulai

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Setup Environment

Copy file `.env.example` ke `.env`:

```bash
copy .env.example .env
```

### 3. Jalankan Docker Services

```bash
docker-compose up -d
```

Tunggu beberapa saat sampai semua services ready.

### 4. Jalankan Database Migrations

```bash
# Windows
migrate.bat up

# Linux/Mac
./migrate.sh up

# Atau menggunakan make
make migrate-up
```

### 5. Jalankan Aplikasi

```bash
# API server (recommended)
go run ./cmd/server

# Atau lewat make
make run
```

Atau untuk demo bootstrap semua services (tanpa HTTP API):

```bash
go run main.go
```

Anda akan melihat output:

```
time=... level=INFO msg="postgres connected and healthy"
time=... level=INFO msg="redis connected and healthy"
...
```

**Matikan fitur optional via env:**

```bash
RABBITMQ_ENABLED=false go run ./cmd/server
REDIS_ENABLED=false REALTIME_ENABLED=false go run ./cmd/server
# Hanya SSE (tanpa WebSocket), atau hanya WebSocket:
REALTIME_WS_ENABLED=false go run ./cmd/server
REALTIME_SSE_ENABLED=false go run ./cmd/server
# Fan-out lintas instance: redis (default) / rabbitmq / both
REALTIME_BRIDGE=rabbitmq go run ./cmd/server
```

## 📚 Contoh Penggunaan Lengkap

Lihat file `examples/basic_usage.go` untuk contoh lengkap semua fitur.

### Menjalankan Contoh Lengkap

Buat file baru `main_example.go`:

```go
package main

import (
	"golang/examples"
)

func main() {
	examples.RunAllExamples()
}
```

Jalankan:

```bash
go run main_example.go
```

## 🔧 Management UIs

Setelah `docker-compose up -d`, akses:

- **RabbitMQ Management**: http://localhost:15672
    - Username: `guest`
    - Password: `guest`

## 📖 Dokumentasi Detail

### PostgreSQL

```go
// Connect
cfg := config.Load()
db, err := database.NewPostgresDB(&cfg.Database)
if err != nil {
log.Fatal(err)
}
defer db.Close()

// Create model
type User struct {
ID    uint   `gorm:"primarykey"`
Name  string
Email string `gorm:"uniqueIndex"`
}

// Auto migrate
db.AutoMigrate(&User{})

// Create
user := User{Name: "John", Email: "john@example.com"}
db.DB.Create(&user)

// Read
var user User
db.DB.First(&user, 1) // By ID
db.DB.Where("email = ?", "john@example.com").First(&user)

// Update
db.DB.Model(&user).Update("name", "John Doe")

// Delete
db.DB.Delete(&user)
```

### Redis

```go
// Connect
cfg := config.Load()
redisCache, err := cache.NewRedisCache(&cfg.Redis)
if err != nil {
log.Fatal(err)
}
defer redisCache.Close()

ctx := context.Background()

// Set
redisCache.Set(ctx, "key", "value", 10*time.Minute)

// Get
value, err := redisCache.Get(ctx, "key")

// Cache JSON
data := map[string]interface{}{"id": 1, "name": "John"}
jsonData, _ := json.Marshal(data)
redisCache.Set(ctx, "user:1", jsonData, 5*time.Minute)

// Retrieve JSON
cachedData, _ := redisCache.Get(ctx, "user:1")
var user map[string]interface{}
json.Unmarshal([]byte(cachedData), &user)

// Counter
count, _ := redisCache.Increment(ctx, "visits")

// Lock pattern
acquired, _ := redisCache.SetNX(ctx, "lock:resource", "locked", 30*time.Second)
if acquired {
// Critical section
defer redisCache.Delete(ctx, "lock:resource")
}
```

### RabbitMQ

```go
// Connect
cfg := config.Load()
rmq, err := queue.NewRabbitMQ(&cfg.RabbitMQ, nil)
if err != nil {
log.Fatal(err)
}
defer rmq.Close()

// Declare queue
q, err := rmq.DeclareQueue("my_queue", true, false, false)

// Publish message
type Message struct {
ID      string    `json:"id"`
Content string    `json:"content"`
SentAt  time.Time `json:"sent_at"`
}

msg := Message{ID: "1", Content: "Hello", SentAt: time.Now()}
body, _ := json.Marshal(msg)
rmq.PublishToQueue(context.Background(), "my_queue", body)

// Consume messages
msgs, err := rmq.Consume(context.Background(), "my_queue", "consumer1", false)
for msg := range msgs {
var message Message
json.Unmarshal(msg.Body, &message)

// Process message
log.Printf("Received: %+v", message)

// Acknowledge
msg.Ack(false)
}
```

### Pattern: Work Queue (Task Distribution)

```go
// Producer
for i := 0; i < 10; i++ {
task := fmt.Sprintf("Task #%d", i)
rmq.PublishToQueue("work_queue", []byte(task))
}

// Multiple consumers akan otomatis distribusi task
```

### Pattern: Pub/Sub dengan Exchange

```go
// Declare exchange
rmq.DeclareExchange("logs", "fanout", true, false)

// Bind queue
q, _ := rmq.DeclareQueue("", false, true, true)
rmq.BindQueue(q.Name, "", "logs")

// Publish
rmq.Publish("logs", "", []byte("Log message"))
```

### Pattern: Topic Routing

```go
// Setup
rmq.DeclareExchange("topic_logs", "topic", true, false)
rmq.BindQueue("all_logs", "#", "topic_logs")
rmq.BindQueue("kernel_logs", "kern.*", "topic_logs")

// Publish
rmq.Publish("topic_logs", "kern.critical", []byte("Critical error"))
```

## 🛠️ Commands Berguna

### Docker

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# View logs for specific service
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f rabbitmq

# Check status
docker-compose ps

# Restart service
docker-compose restart postgres

# Stop dan hapus semua data
docker-compose down -v
```

### Database

```bash
# Connect ke PostgreSQL
docker-compose exec postgres psql -U postgres -d mydb

# Connect ke Redis
docker-compose exec redis redis-cli

# Test Redis
docker-compose exec redis redis-cli ping
```

### Debugging

```bash
# Check health
docker-compose ps

# View detailed logs
docker-compose logs --tail=100 -f

# Execute command in container
docker-compose exec postgres psql -U postgres -c "SELECT version();"
```

## 🔍 Troubleshooting

### Services tidak bisa connect

1. Check apakah Docker services running:
   ```bash
   docker-compose ps
   ```

2. Restart services:
   ```bash
   docker-compose down
   docker-compose up -d
   ```

3. Check logs untuk error:
   ```bash
   docker-compose logs
   ```

### Port sudah digunakan

Edit `docker-compose.yml` dan ubah port mapping:

```yaml
ports:
  - "15432:5432"  # Ubah 5432 ke 15432 di host
```

### Clean restart

```bash
docker-compose down -v
docker-compose up -d
```

Ini akan menghapus semua data dan memulai fresh.

## 📁 Struktur Project

```
golang/
├── cache/              # Redis cache implementation
│   └── redis.go
├── config/             # Configuration management
│   └── config.go
├── database/           # Database implementation
│   └── postgres.go
├── queue/              # Message queue implementation
│   └── rabbitmq.go
├── examples/           # Usage examples
│   └── basic_usage.go
├── .env.example        # Environment template
├── docker-compose.yml  # Docker services setup
├── main.go             # Main application
├── go.mod              # Go dependencies
└── README.md           # Full documentation
```

## 🎯 Best Practices

### Database (PostgreSQL)

1. **Gunakan Transaction untuk operasi multiple**
   ```go
   db.DB.Transaction(func(tx *gorm.DB) error {
       // Multiple operations
       return nil
   })
   ```

2. **Index untuk kolom yang sering di-query**
   ```go
   Email string `gorm:"uniqueIndex"`
   ```

3. **Gunakan Preload untuk relationship**
   ```go
   db.DB.Preload("Orders").Find(&users)
   ```

### Redis

1. **Set expiration untuk semua cache**
   ```go
   redisCache.Set(ctx, key, value, 10*time.Minute)
   ```

2. **Gunakan prefix untuk key namespacing**
   ```go
   "user:1", "session:abc123", "cache:product:123"
   ```

3. **Handle redis.Nil error**
   ```go
   val, err := redisCache.Get(ctx, key)
   if err == redis.Nil {
       // Key tidak ada
   }
   ```

### RabbitMQ

1. **Gunakan durable queues untuk production**
   ```go
   rmq.DeclareQueue("queue", true, false, false)
   ```

2. **Manual acknowledgment untuk reliability**
   ```go
   msgs, _ := rmq.Consume("queue", "consumer", false)
   for msg := range msgs {
       // Process
       msg.Ack(false)
   }
   ```

3. **Gunakan exchange untuk routing yang flexible**
   ```go
   rmq.DeclareExchange("events", "topic", true, false)
   rmq.Publish("events", "user.created", body)
   ```

## 📞 Support

Untuk pertanyaan atau masalah, lihat:

- `README.md` - Dokumentasi lengkap
- `examples/basic_usage.go` - Contoh kode lengkap

---

**Selamat coding! 🎉**

