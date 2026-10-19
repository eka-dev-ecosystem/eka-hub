# 🎉 Migration System - Setup Complete!

## ✅ Yang Sudah Dibuat

### 📁 File Structure

```
golang/
├── migrations/                          # Migration SQL files
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_user_fields.up.sql
│   ├── 000002_add_user_fields.down.sql
│   └── README.md                        # Migration documentation
├── database/
│   ├── postgres.go                      # Database connection (sqlx)
│   └── migrate/
│       └── migrate.go                   # Migration helper functions
├── cmd/
│   └── migrate.go                       # Migration CLI tool
├── migrate.bat                          # Windows migration script
├── migrate.sh                           # Linux/Mac migration script
├── Dockerfile.migrate                   # Docker image for migrations
└── docker-compose.yml                   # Updated with migration service
```

## 🚀 Cara Menggunakan

### 1. Windows (Paling Mudah)

```cmd
REM Jalankan semua migrations
migrate.bat up

REM Buat migration baru
migrate.bat create add_products_table

REM Check version
migrate.bat version

REM Rollback 1 migration
migrate.bat down

REM Rollback 2 migrations
migrate.bat down 2

REM Lihat help
migrate.bat help
```

### 2. Linux/Mac

```bash
# Jalankan semua migrations
./migrate.sh up

# Buat migration baru
./migrate.sh create add_products_table

# Check version
./migrate.sh version

# Rollback
./migrate.sh down 2
```

### 3. Menggunakan Make

```bash
# Jalankan migrations
make migrate-up

# Buat migration baru
make migrate-create name=add_products_table

# Check version
make migrate-version

# Rollback
make migrate-down
```

### 4. Direct Go Command

```bash
# Run migrations
go run cmd/migrate.go -action=up

# Create migration
go run cmd/migrate.go -action=create -name=add_products_table

# Check version
go run cmd/migrate.go -action=version

# Rollback
go run cmd/migrate.go -action=down -steps=1
```

### 5. Menggunakan Binary

```bash
# Build terlebih dahulu
go build -o bin/migrate.exe cmd/migrate.go

# Lalu gunakan
bin/migrate.exe -action=up
bin/migrate.exe -action=create -name=new_migration
```

## 📝 Contoh Migration Files

### File: migrations/000001_create_users_table.up.sql

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Auto update updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### File: migrations/000001_create_users_table.down.sql

```sql
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## 🔧 Integration dengan Code

### Auto-run migrations on app startup

Edit `main.go`:

```go
package main

import (
    "log"
    "golang/config"
    "golang/database"
    "golang/database/migrate"
)

func main() {
    cfg := config.Load()
    
    // Run migrations automatically
    log.Println("Running database migrations...")
    if err := migrate.RunMigrations(&cfg.Database, "./migrations"); err != nil {
        log.Printf("Warning: Failed to run migrations: %v", err)
    } else {
        log.Println("✓ Migrations completed successfully")
    }
    
    // Connect to database
    db, err := database.NewPostgresDB(&cfg.Database)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Your application code...
}
```

## 🐳 Docker Integration

### Auto-run migrations with Docker Compose

Edit `docker-compose.yml` dan uncomment migration service:

```yaml
services:
  # ... postgres, redis, rabbitmq services ...
  
  migrate:
    build:
      context: .
      dockerfile: Dockerfile.migrate
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: mydb
      DB_SSLMODE: disable
    command: ["./migrate", "-action=up"]
    restart: on-failure
```

Lalu jalankan:

```bash
docker-compose up -d
```

## 🎯 Workflow Development

### 1. Start Services

```bash
docker-compose up -d
```

### 2. Run Migrations

```bash
migrate.bat up
# atau
make migrate-up
```

### 3. Develop & Create Migrations

```bash
# Buat migration baru saat butuh perubahan schema
migrate.bat create add_orders_table

# Edit file SQL yang dibuat di folder migrations/
# - Edit .up.sql untuk forward migration
# - Edit .down.sql untuk rollback

# Test migration
migrate.bat up

# Test rollback
migrate.bat down

# Jika ok, commit ke git
git add migrations/
git commit -m "Add orders table migration"
```

### 4. Check Status

```bash
# Lihat versi migration saat ini
migrate.bat version

# Output: Current version: 2
```

## 📦 Dependencies

Migration system menggunakan:

- ✅ `github.com/golang-migrate/migrate/v4 v4.19.1` - Migration tool
- ✅ `github.com/jmoiron/sqlx v1.4.0` - SQL extensions
- ✅ `github.com/lib/pq v1.10.9` - PostgreSQL driver
- ✅ `github.com/google/uuid v1.6.0` - UUID generation

## 🎓 Tips

1. **Migration naming**: Gunakan nama yang deskriptif
   ```bash
   migrate.bat create create_users_table
   migrate.bat create add_email_to_users
   migrate.bat create create_index_on_email
   ```

2. **Always reversible**: Pastikan setiap UP migration punya DOWN yang benar

3. **Test locally first**: Selalu test migration di local sebelum push ke production

4. **Use transactions**: PostgreSQL migrations run in transactions by default

5. **Keep migrations small**: Satu migration = satu logical change

## 🚨 Troubleshooting

### Error: "Dirty database version"

```bash
# Check version
migrate.bat version
# Output: Current version: 1 (dirty - migration incomplete)

# Fix: Force to last known good version
migrate.bat force 1

# Then try again
migrate.bat up
```

### Error: "No change"

Ini normal jika tidak ada migration baru yang perlu dijalankan.

### Reset Database Completely

```bash
# Drop all tables
migrate.bat drop

# Run migrations from scratch
migrate.bat up
```

## 📚 Dokumentasi Lengkap

Lihat file berikut untuk informasi lebih detail:

- `migrations/README.md` - Dokumentasi lengkap migration system
- `README.md` - Dokumentasi project secara keseluruhan
- `QUICKSTART.md` - Quick start guide

## ✨ Ready to Use!

Migration system sudah siap digunakan. Cukup jalankan:

```bash
# Start services
docker-compose up -d

# Run migrations
migrate.bat up

# Start coding!
```

Happy coding! 🚀

