# Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

## Quick Start

### Run Migrations

```bash
# Run all pending migrations
make migrate-up

# Or using go run directly
go run cmd/migrate.go -action=up
```

### Create New Migration

```bash
# Create new migration files
make migrate-create name=create_products_table

# Or using go run
go run cmd/migrate.go -action=create -name=create_products_table
```

This will create two files:

- `migrations/000003_create_products_table.up.sql` - Forward migration
- `migrations/000003_create_products_table.down.sql` - Rollback migration

### Check Migration Version

```bash
# Show current migration version
make migrate-version

# Or
go run cmd/migrate.go -action=version
```

### Rollback Migrations

```bash
# Rollback last migration
make migrate-down

# Rollback multiple migrations
go run cmd/migrate.go -action=down -steps=2
```

## Migration Commands

| Command                        | Description                                |
|--------------------------------|--------------------------------------------|
| `make migrate-up`              | Run all pending migrations                 |
| `make migrate-down`            | Rollback last migration                    |
| `make migrate-create name=xxx` | Create new migration files                 |
| `make migrate-version`         | Show current migration version             |
| `make migrate-force version=N` | Force migration version (use with caution) |
| `make migrate-drop`            | Drop all tables (dangerous!)               |

## Migration File Format

Migration files follow the naming pattern:

```
{version}_{name}.{up|down}.sql
```

Example:

- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`

### UP Migration (Forward)

The `.up.sql` file contains SQL to apply the migration:

```sql
CREATE TABLE users
(
    id         UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
    name       VARCHAR(255)        NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### DOWN Migration (Rollback)

The `.down.sql` file contains SQL to revert the migration:

```sql
DROP TABLE IF EXISTS users;
```

## Existing Migrations

1. **000001_create_users_table** - Creates users table with basic fields
2. **000002_add_user_fields** - Adds phone, is_active, and last_login fields

## Running Migrations Automatically

### On Application Start

You can run migrations automatically when your app starts:

```go
package main

import (
	"log"
	"golang/config"
	"golang/database/migrate"
)

func main() {
	cfg := config.Load()

	// Run migrations
	if err := migrate.RunMigrations(&cfg.Database, "./migrations"); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	// Continue with your application...
}
```

### With Docker Compose

You can add a migration service to your `docker-compose.yml`:

```yaml
migrate:
  build: .
  command: go run cmd/migrate.go -action=up
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
```

## Environment Variables

Migration tool uses the same database configuration as the main application:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mydb
DB_SSLMODE=disable
```

## Troubleshooting

### Dirty Migration State

If a migration fails halfway, the database will be in a "dirty" state:

```bash
# Check current state
make migrate-version

# If dirty, force to a specific version
make migrate-force version=1

# Then try running migrations again
make migrate-up
```

### Reset Database

To completely reset the database:

```bash
# Drop all tables
make migrate-drop

# Run migrations from scratch
make migrate-up
```

Or with Docker:

```bash
# Remove database volume
docker-compose down -v

# Start fresh
docker-compose up -d
make migrate-up
```

## Best Practices

1. **Always create both UP and DOWN migrations** - Make sure your migrations are reversible
2. **Test migrations locally first** - Run and rollback before committing
3. **Keep migrations small** - One logical change per migration
4. **Never modify existing migrations** - Create new migrations to fix issues
5. **Use transactions** - PostgreSQL migrations run in transactions by default
6. **Backup before migrating in production** - Always backup your database first

## CLI Usage Examples

```bash
# Run all pending migrations
go run cmd/migrate.go -action=up

# Rollback last migration
go run cmd/migrate.go -action=down -steps=1

# Rollback 3 migrations
go run cmd/migrate.go -action=down -steps=3

# Check current version
go run cmd/migrate.go -action=version

# Create new migration
go run cmd/migrate.go -action=create -name=add_products_table

# Force version (dangerous - use only if needed)
go run cmd/migrate.go -action=force -version=2

# Drop all tables (dangerous!)
go run cmd/migrate.go -action=drop

# Specify custom migrations path
go run cmd/migrate.go -action=up -path=./db/migrations
```

## Integration with CI/CD

Add migration step to your CI/CD pipeline:

```yaml
# Example GitHub Actions
- name: Run Database Migrations
  run: |
    make migrate-up
  env:
    DB_HOST: localhost
    DB_PORT: 5432
    DB_USER: postgres
    DB_PASSWORD: postgres
    DB_NAME: testdb
    DB_SSLMODE: disable
```

## Resources

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL migration best practices](https://www.postgresql.org/docs/current/ddl.html)

