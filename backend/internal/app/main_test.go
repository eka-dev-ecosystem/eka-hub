//go:build integration

package app_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Disable file logging for tests
	os.Setenv("LOG_TO_FILE", "false")
	os.Setenv("LOG_LEVEL", "error")

	// We only spin up Postgres for the integration tests since the existing tests
	// mostly rely on PostgreSQL being reachable, and mock Redis/RabbitMQ.
	fmt.Println("Starting PostgreSQL Testcontainer...")
	pgC, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("golang_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("Failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	pgHost, _ := pgC.Host(ctx)
	pgPort, _ := pgC.MappedPort(ctx, "5432/tcp")

	os.Setenv("TEST_DB_HOST", pgHost)
	os.Setenv("TEST_DB_PORT", pgPort.Port())
	os.Setenv("TEST_DB_NAME", "golang_test")
	os.Setenv("TEST_DB_USER", "postgres")
	os.Setenv("TEST_DB_PASSWORD", "postgres")

	// The existing integration_test.go uses DB_HOST, DB_PORT etc in migrate.RunMigrations
	os.Setenv("DB_HOST", pgHost)
	os.Setenv("DB_PORT", pgPort.Port())
	os.Setenv("DB_NAME", "golang_test")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_SSLMODE", "disable")

	// Disable Redis and RabbitMQ to test core flows without them
	os.Setenv("TEST_REDIS_ENABLED", "false")

	// Run tests
	code := m.Run()

	// Cleanup
	_ = pgC.Terminate(ctx)

	os.Exit(code)
}
