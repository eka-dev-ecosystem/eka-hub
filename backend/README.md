# Go Project with RabbitMQ, PostgreSQL, and Redis

Project template lengkap dengan integrasi RabbitMQ, PostgreSQL, dan Redis untuk aplikasi Go.

## 📋 Daftar Isi

- [Fitur](#fitur)
- [Prerequisites](#prerequisites)
- [Instalasi](#instalasi)
- [Konfigurasi](#konfigurasi)
- [Menjalankan Aplikasi](#menjalankan-aplikasi)
- [Dokumentasi Penggunaan](#dokumentasi-penggunaan)
    - [PostgreSQL](#postgresql)
    - [Redis](#redis)
    - [RabbitMQ](#rabbitmq)
- [Contoh Penggunaan](#contoh-penggunaan)
- [Docker Setup](#docker-setup)
- [Troubleshooting](#troubleshooting)

## ✨ Fitur

- ✅ **Repository Pattern** - Clean architecture dengan separation of concerns (Handler → Service → Repository)
- ✅ **PostgreSQL** dengan sqlx untuk raw SQL queries
- ✅ **Database Migrations** dengan golang-migrate/migrate
- ✅ **Database Transactions** - TypeScript-like withTransaction helper
- ✅ **JWT Authentication** - Access token & refresh token flow dengan Redis-based revocation
- ✅ **Global Error Handling** - Automatic PostgreSQL error parsing (duplicate, foreign key, etc.)
- ✅ **Standard Base Response** - TypeScript-compatible response structure
- ✅ **Redis** untuk caching dan token storage
- ✅ **RabbitMQ** untuk message queuing
- ✅ **Fiber** web framework (Express.js-like)
- ✅ **Validator** untuk validasi data dengan auto-formatting errors
- ✅ **Viper** untuk configuration management
- ✅ **Docker Compose** untuk development environment
- ✅ **Health checks** untuk semua services
- ✅ **Connection pooling** dan retry logic
- ✅ **Contoh lengkap** untuk setiap service

## 🔧 Prerequisites

- Go 1.25 atau lebih tinggi
- Docker dan Docker Compose (opsional, untuk development)
- PostgreSQL 14+ (jika tidak menggunakan Docker)
- Redis 7+ (jika tidak menggunakan Docker)
- RabbitMQ 3+ (jika tidak menggunakan Docker)

## 📥 Instalasi

### 1. Clone atau setup project

```bash
cd D:\template\golang
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Copy environment file

```bash
copy .env.example .env
```

### 4. Start services menggunakan Docker (Recommended)

```bash
docker-compose up -d
```

Atau install dan jalankan services secara manual.

## ⚙️ Konfigurasi

Edit file `.env` sesuai dengan environment Anda:

```env
# Application & Environment
APP_NAME=golang-api
ENVIRONMENT=development   # development | production (JSON logs)

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=3000

# Logging
LOG_LEVEL=info              # debug | info | warn | error

# Feature Flags (on/off fitur opsional)
REDIS_ENABLED=true
RABBITMQ_ENABLED=true
REALTIME_ENABLED=true        # WebSocket / SSE

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mydb
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/
```

### Feature Flags

Setiap integrasi opsional bisa **dimatikan lewat environment variable**. Saat
fitur dimatikan, dependency-nya tidak pernah diinisialisasi dan aplikasi
tetap berjalan normal tanpa fitur tersebut.

```bash
# Jalankan API server tanpa RabbitMQ (misal di development)
RABBITMQ_ENABLED=false go run ./cmd/server

# Atau lewat .env
# RABBITMQ_ENABLED=false
```

Daftar flag:

| Flag                  | Fungsi                                                     |
| --------------------- | ---------------------------------------------------------- |
| `REDIS_ENABLED`       | Aktifkan cache, token revocation & realtime bridge Redis   |
| `RABBITMQ_ENABLED`    | Aktifkan message broker RabbitMQ (dan realtime bridge)     |
| `REALTIME_ENABLED`    | Master switch WebSocket & SSE (matikan untuk menonaktifkan semua) |
| `REALTIME_SSE_ENABLED`| Aktifkan Server-Sent Events saja                            |
| `REALTIME_WS_ENABLED` | Aktifkan WebSocket saja                                     |
| `REALTIME_BRIDGE`     | Transport fan-out realtime: `redis` (default), `rabbitmq`, `both` |

> **Scaling realtime:** pada deployment multi-instance, hub tetap melakukan
> broadcast lokal dan **wajib menerbitkan (publish) ke bridge** agar setiap
> instance menerima pesan yang sama. Pilih transport lewat `REALTIME_BRIDGE`:
> Redis pub/sub (default), RabbitMQ fanout exchange, atau keduanya. Jika salah
> satu service pendukung mati/dimatikan, hub otomatis turun ke mode
> single-instance (broadcast lokal saja) — tidak fail.

### Request Body Schema: JSON & Multipart

Semua endpoint POST/PUT memakai binder `pkg/validation` yang menerima satu DTO
untuk **tiga bentuk body sekaligus**:

- `application/json`
- `application/x-www-form-urlencoded`
- `multipart/form-data` (termasuk upload file, field bertipe
  `*multipart.FileHeader` / `[]*multipart.FileHeader`)

Tipe data form (string/int/float/bool) dikonversi otomatis ke tipe field-nya.
Contoh DTO:

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"omitempty,gte=0,lte=130"`
    Avatar *multipart.FileHeader `json:"avatar"` // field upload
}

// di handler:
var req CreateUserRequest
if err := validation.Default.BindAndValidate(c, &req); err != nil {
    return err   // error handler otomatis membentuk respon VALIDATION_ERROR
}
```

Client bisa mengirim field yang sama lewat JSON **atau** form/multipart tanpa
mengubah kode server. Dokumen error tetap memakai format standar
(`VALIDATION_ERROR` 422/400). Lihat [pkg/validation](pkg/validation).

### Logging

Logging selalu **structured JSON** (`log/slog`) di seluruh modul — setiap
baris adalah satu objek JSON (startup, access log, error), tanpa banner
tambahan, sehingga langsung siap di-consume oleh log pipeline:

```json
{"time":"...","level":"INFO","msg":"http request","request_id":"...","method":"POST","path":"/api/v1/users","status":201,"latency":123456,"ip":"127.0.0.1","user_agent":"curl/8.18.0","error":""}
```

- `LOG_LEVEL` mengontrol level: `debug`, `info`, `warn`, `error`
- `LOG_FILE` (opsional): semua log juga ditulis ke file JSON tersebut
- Klik [pkg/logging/logger.go](pkg/logging/logger.go) untuk objek logger global

## 🚀 Menjalankan Aplikasi

### Start Docker services

```bash
docker-compose up -d
```

### Run API server

```bash
# Lewat Makefile
make run            # -> go run ./cmd/server

# atau langsung
go run ./cmd/server

# matikan RabbitMQ tanpa mengubah kode
RABBITMQ_ENABLED=false go run ./cmd/server

# untuk demo bootstrap semua services (tanpa HTTP API)
make run-demo       # -> go run main.go
```

### Check status services

```bash
docker-compose ps
```

### View logs

```bash
# Semua services
docker-compose logs -f

# Specific service
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f rabbitmq
```

### Run database migrations

```bash
# Run all pending migrations
make migrate-up

# Check migration version
make migrate-version

# Create new migration
make migrate-create name=create_products_table

# Rollback last migration
make migrate-down
```

Lihat [migrations/README.md](migrations/README.md) untuk dokumentasi lengkap tentang migrations.

### 🧪 Testing

Dua level test, terpisah lewat build tag (`integration`) sehingga unit test selalu cepat dan tidak butuh database:

```bash
# 1. Unit test (tanpa dependency eksternal, jalan di mana saja)
go test ./...

# 2. Integration test (REAL, tanpa mock, butuh PostgreSQL hidup)
#    Jalankan migration dulu:
#    go run ./cmd -action up   # dengan DB_* pointing ke test DB
#    lalu:
go test -tags=integration ./internal/app/...

# Koneksi test DB bisa di-override:
# TEST_DB_HOST, TEST_DB_PORT, TEST_DB_NAME, TEST_DB_USER, TEST_DB_PASSWORD
```

Integration test mem-boot seluruh container aplikasi (`app.Boot` + `app.NewHTTP`)
tanpa mock, lalu memverifikasi end-to-end: health check + feature flags,
auth lifecycle (register/login/refresh/profile/change-password/logout),
user CRUD dan kode error (`NOT_FOUND`, `UNAUTHORIZED`, `CONFLICT`,
`VALIDATION_ERROR`, dst.), proteksi route, serta header `X-Request-ID`.
Jika database tidak bisa diakses, integration test akan `skip` (tidak fail).

### 🎨 Code Quality (Formatting & Lint)

Setup "ESLint"-nya Go — formatting gofmt/goimports, lint extended lewat
golangci-lint (`govet`, `staticcheck`, `errcheck`, `revive`, `gocritic`,
`gocyclo`, `bodyclose`, dst.). Konfigurasi: `.golangci.yml` dan
`.editorconfig` (konsistensi indentasi/line-ending lintas editor).

```bash
# 1. Install tools sekali (butuh Go)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

# 2. Format otomatis (gofmt -s + goimports, import lokal dikelompokkan)
make fmt

# 3. Quality gate: format + vet + lint sekaligus (fail kalau ada temuan)
make check

# Komponen individual:
make fmt-check        # fail kalau belum diformat
make vet              # go vet ./...
make lint             # golangci-lint run ./...
make test-unit        # go test ./...
make test-integration # go test -tags=integration ./internal/app/...
```

Linter menargetkan **nol temuan** (`max-issues-per-linter: 0`), jadi
`make check` sudah layak jadi gate sebelum commit tanpa butuh CI.

### Stop services

```bash
docker-compose down
```

### Stop dan hapus volumes (data akan hilang)

```bash
docker-compose down -v
```

## 📚 Dokumentasi Penggunaan

### PostgreSQL

#### Inisialisasi Database Connection

```go
import (
"golang/config"
"golang/database"
)

// Load config
cfg := config.Load()

// Connect to database
db, err := database.NewPostgresDB(&cfg.Database)
if err != nil {
log.Fatal(err)
}
defer db.Close()
```

#### Membuat Model

```go
type User struct {
ID        uint      `gorm:"primarykey" json:"id"`
Name      string    `json:"name"`
Email     string    `gorm:"uniqueIndex" json:"email"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}
```

#### Auto Migration

```go
err := db.AutoMigrate(&User{})
if err != nil {
log.Fatal(err)
}
```

#### Create

```go
user := User{
Name:  "John Doe",
Email: "john@example.com",
}
result := db.DB.Create(&user)
if result.Error != nil {
log.Fatal(result.Error)
}
```

#### Read

```go
// Find by ID
var user User
db.DB.First(&user, 1)

// Find by condition
db.DB.Where("email = ?", "john@example.com").First(&user)

// Find all
var users []User
db.DB.Find(&users)

// With pagination
db.DB.Limit(10).Offset(0).Find(&users)
```

#### Update

```go
// Update single field
db.DB.Model(&user).Update("name", "John Smith")

// Update multiple fields
db.DB.Model(&user).Updates(User{Name: "John Smith", Email: "johnsmith@example.com"})

// Update dengan map
db.DB.Model(&user).Updates(map[string]interface{}{"name": "John", "age": 18})
```

#### Delete

```go
// Soft delete
db.DB.Delete(&user)

// Permanent delete
db.DB.Unscoped().Delete(&user)
```

#### Transaction

```go
err := db.DB.Transaction(func (tx *gorm.DB) error {
// Create user
if err := tx.Create(&user).Error; err != nil {
return err
}

// Create related record
profile := Profile{UserID: user.ID, Bio: "Hello"}
if err := tx.Create(&profile).Error; err != nil {
return err
}

return nil
})
```

#### Health Check

```go
err := db.HealthCheck()
if err != nil {
log.Printf("Database is down: %v", err)
}
```

### Redis

#### Inisialisasi Redis Connection

```go
import (
"context"
"golang/cache"
"golang/config"
)

cfg := config.Load()
redisCache, err := cache.NewRedisCache(&cfg.Redis)
if err != nil {
log.Fatal(err)
}
defer redisCache.Close()

ctx := context.Background()
```

#### Set dan Get

```go
// Set dengan expiration
err := redisCache.Set(ctx, "key", "value", 10*time.Minute)

// Get value
value, err := redisCache.Get(ctx, "key")
if err == redis.Nil {
log.Println("Key tidak ditemukan")
} else if err != nil {
log.Fatal(err)
}
```

#### Caching JSON Data

```go
// Cache struct/object
data := map[string]interface{}{
"id":   1,
"name": "John",
}
jsonData, _ := json.Marshal(data)
redisCache.Set(ctx, "user:1", jsonData, 5*time.Minute)

// Retrieve
cachedData, _ := redisCache.Get(ctx, "user:1")
var user map[string]interface{}
json.Unmarshal([]byte(cachedData), &user)
```

#### Check Existence

```go
exists, err := redisCache.Exists(ctx, "key")
if exists > 0 {
log.Println("Key exists")
}
```

#### Delete

```go
err := redisCache.Delete(ctx, "key1", "key2", "key3")
```

#### Increment (Counter)

```go
count, err := redisCache.Increment(ctx, "counter:visits")
log.Printf("Visit count: %d", count)
```

#### Set If Not Exists (Lock Pattern)

```go
// Acquire lock
acquired, err := redisCache.SetNX(ctx, "lock:resource", "locked", 30*time.Second)
if acquired {
// Process critical section
defer redisCache.Delete(ctx, "lock:resource")
}
```

#### Set Expiration

```go
err := redisCache.Expire(ctx, "key", 1*time.Hour)
```

#### Health Check

```go
err := redisCache.HealthCheck(ctx)
if err != nil {
log.Printf("Redis is down: %v", err)
}
```

### RabbitMQ

#### Inisialisasi RabbitMQ Connection

```go
import (
"golang/config"
"golang/pkg/queue"
)

cfg := config.Load()
rmq, err := queue.NewRabbitMQ(&cfg.RabbitMQ, nil) // logger opsional
if err != nil {
log.Fatal(err)
}
defer rmq.Close()
```

#### Declare Queue

```go
// Parameters: name, durable, autoDelete, exclusive
q, err := rmq.DeclareQueue("my_queue", true, false, false)
if err != nil {
log.Fatal(err)
}
```

#### Publish Message ke Queue

```go
type Message struct {
ID      string    `json:"id"`
Content string    `json:"content"`
SentAt  time.Time `json:"sent_at"`
}

msg := Message{
ID:      "msg-1",
Content: "Hello RabbitMQ",
SentAt:  time.Now(),
}

body, _ := json.Marshal(msg)
err := rmq.PublishToQueue(context.Background(), "my_queue", body)
if err != nil {
log.Fatal(err)
}
```

#### Consume Messages dari Queue

```go
// Parameters: queueName, consumerName, autoAck
msgs, err := rmq.Consume(context.Background(), "my_queue", "my_consumer", false)
if err != nil {
log.Fatal(err)
}

// Process messages
for msg := range msgs {
var message Message
json.Unmarshal(msg.Body, &message)

log.Printf("Received: %+v", message)

// Process message...

// Acknowledge message
msg.Ack(false)

// Or reject and requeue
// msg.Nack(false, true)
}
```

#### Declare Exchange

```go
// Parameters: name, type, durable, autoDelete
// Types: "direct", "fanout", "topic", "headers"
err := rmq.DeclareExchange("my_exchange", "topic", true, false)
```

#### Bind Queue to Exchange

```go
err := rmq.BindQueue("queue_name", "routing.key", "exchange_name")
```

#### Publish to Exchange

```go
body, _ := json.Marshal(message)
err := rmq.Publish(context.Background(), "my_exchange", "routing.key", body)
```

> Catatan API: seluruh method bertipe `Publish*`/`Consume*` menerima `context.Context`
> sebagai argumen pertama sejak refactor (lihat `pkg/queue/broker.go`).

#### Pattern: Work Queue (Task Distribution)

```go
// Producer
for i := 0; i < 10; i++ {
task := fmt.Sprintf("Task #%d", i)
rmq.PublishToQueue("work_queue", []byte(task))
}

// Multiple consumers will automatically distribute the work
```

#### Pattern: Pub/Sub (Fanout)

```go
// Setup
rmq.DeclareExchange("logs", "fanout", true, false)
q1, _ := rmq.DeclareQueue("", false, true, true) // Temporary queue
rmq.BindQueue(q1.Name, "", "logs")

// Publish
rmq.Publish("logs", "", []byte("Log message"))
```

#### Pattern: Routing (Direct)

```go
// Setup
rmq.DeclareExchange("direct_logs", "direct", true, false)
rmq.BindQueue("error_queue", "error", "direct_logs")
rmq.BindQueue("info_queue", "info", "direct_logs")

// Publish
rmq.Publish("direct_logs", "error", []byte("Error occurred"))
rmq.Publish("direct_logs", "info", []byte("Info message"))
```

#### Pattern: Topics

```go
// Setup
rmq.DeclareExchange("topic_logs", "topic", true, false)
rmq.BindQueue("all_logs", "#", "topic_logs")
rmq.BindQueue("kernel_logs", "kern.*", "topic_logs")
rmq.BindQueue("critical_logs", "*.critical", "topic_logs")

// Publish
rmq.Publish("topic_logs", "kern.critical", []byte("Kernel critical error"))
```

#### Health Check

```go
err := rmq.HealthCheck()
if err != nil {
log.Printf("RabbitMQ is down: %v", err)
}
```

## 💡 Contoh Penggunaan

Lihat file `examples/basic_usage.go` untuk contoh lengkap penggunaan semua services.

### Menjalankan Contoh

```go
package main

import (
	"golang/examples"
)

func main() {
	examples.RunAllExamples()
}
```

## 🐳 Docker Setup

### Management UI URLs

Setelah menjalankan `docker-compose up -d`, Anda dapat mengakses:

- **RabbitMQ Management**: http://localhost:15672
    - Username: `guest`
    - Password: `guest`

### Useful Docker Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose stop

# Restart a service
docker-compose restart postgres

# View logs
docker-compose logs -f

# Execute command in container
docker-compose exec postgres psql -U postgres -d mydb
docker-compose exec redis redis-cli

# Remove all containers and volumes
docker-compose down -v
```

### Connect ke Database dari Host

```bash
# PostgreSQL
psql -h localhost -p 5432 -U postgres -d mydb

# Redis
redis-cli -h localhost -p 6379

# RabbitMQ
# Management UI: http://localhost:15672
```

## 🔍 Troubleshooting

### Connection Refused Errors

1. Pastikan Docker containers berjalan:
   ```bash
   docker-compose ps
   ```

2. Check health status:
   ```bash
   docker-compose ps
   ```

3. View logs untuk error:
   ```bash
   docker-compose logs postgres
   docker-compose logs redis
   docker-compose logs rabbitmq
   ```

### Port Already in Use

Jika port sudah digunakan, edit `docker-compose.yml` dan ubah port mapping:

```yaml
ports:
  - "15432:5432"  # Ubah port host
```

### Database Connection Issues

1. Pastikan credentials di `.env` sesuai dengan `docker-compose.yml`
2. Check database logs:
   ```bash
   docker-compose logs postgres
   ```

### Redis Connection Issues

1. Test koneksi:
   ```bash
   docker-compose exec redis redis-cli ping
   ```

2. Check password di config jika ada

### RabbitMQ Connection Issues

1. Pastikan RabbitMQ sudah ready:
   ```bash
   docker-compose logs rabbitmq | grep "Server startup complete"
   ```

2. Access management UI: http://localhost:15672

### Clean Restart

Jika ada masalah, lakukan clean restart:

```bash
docker-compose down -v
docker-compose up -d
```

## 📄 License

MIT License

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Support

Jika ada pertanyaan atau issue, silakan buat issue di repository ini.

---

**Happy Coding! 🚀**

