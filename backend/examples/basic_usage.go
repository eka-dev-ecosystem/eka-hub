package examples

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"golang/cache"
	"golang/config"
	"golang/database"
	"golang/pkg/queue"
)

// Example model for database
type User struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ExampleMessage for RabbitMQ
type ExampleMessage struct {
	ID      string    `json:"id"`
	Message string    `json:"message"`
	SentAt  time.Time `json:"sent_at"`
}

// PostgresExample demonstrates PostgreSQL usage with sqlx
func PostgresExample(db *database.PostgresDB) {
	log.Println("=== PostgreSQL Example ===")

	// Create table (simple migration)
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := db.DB.Exec(createTableSQL)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}
	log.Println("Table 'users' ready")

	// Create a user
	user := User{
		ID:        uuid.New().String(),
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	insertSQL := `INSERT INTO users (id, name, email, created_at, updated_at) 
				  VALUES ($1, $2, $3, $4, $5)`
	err = db.WithTransaction(func(tx *sqlx.Tx) error {
		_, err := tx.Exec(insertSQL, user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
		return err
	})
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		log.Printf("Created user: %+v", user)
	}

	// Find user by email
	var foundUser User
	err = db.WithReadOnlyTransaction(func(tx *sqlx.Tx) error {
		return tx.Get(&foundUser, "SELECT * FROM users WHERE email = $1", "john@example.com")
	})
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to find user: %v", err)
	} else if err == nil {
		log.Printf("Found user: %+v", foundUser)
	}

	// Update user
	updateSQL := `UPDATE users SET name = $1, updated_at = $2 WHERE email = $3`
	err = db.WithTransaction(func(tx *sqlx.Tx) error {
		_, err := tx.Exec(updateSQL, "John Smith", time.Now(), "john@example.com")
		return err
	})
	if err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		log.Println("Updated user")
	}

	// List all users
	var users []User
	err = db.WithReadOnlyTransaction(func(tx *sqlx.Tx) error {
		return tx.Select(&users, "SELECT * FROM users ORDER BY created_at DESC")
	})
	if err != nil {
		log.Printf("Failed to list users: %v", err)
	} else {
		log.Printf("All users: %+v", users)
	}

	// Delete user (optional - uncomment if needed)
	// err = db.WithTransaction(func(tx *sqlx.Tx) error {
	// 	_, err := tx.Exec("DELETE FROM users WHERE email = $1", "john@example.com")
	// 	return err
	// })
	// if err != nil {
	// 	log.Printf("Failed to delete user: %v", err)
	// } else {
	// 	log.Println("Deleted user")
	// }

	// Transaction helper example
	newUser := User{
		ID:        uuid.New().String(),
		Name:      "Jane Doe",
		Email:     "jane@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.WithTransaction(func(tx *sqlx.Tx) error {
		_, err = tx.Exec(insertSQL, newUser.ID, newUser.Name, newUser.Email, newUser.CreatedAt, newUser.UpdatedAt)
		if err != nil {
			return err
		}
		log.Printf("Transaction: Created user %+v", newUser)
		return nil
	})
	if err != nil {
		log.Printf("Transaction failed: %v", err)
	}
}

// RedisExample demonstrates Redis usage
func RedisExample(redisCache *cache.RedisCache) {
	log.Println("=== Redis Example ===")
	ctx := context.Background()

	// Set a simple key-value
	err := redisCache.Set(ctx, "greeting", "Hello, Redis!", 10*time.Minute)
	if err != nil {
		log.Printf("Failed to set key: %v", err)
		return
	}
	log.Println("Set key 'greeting'")

	// Get the value
	val, err := redisCache.Get(ctx, "greeting")
	if err != nil {
		log.Printf("Failed to get key: %v", err)
	} else {
		log.Printf("Got value: %s", val)
	}

	// Set with object (JSON)
	userData := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}
	userJSON, _ := json.Marshal(userData)
	err = redisCache.Set(ctx, "user:alice", userJSON, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to set user data: %v", err)
	} else {
		log.Println("Set user data")
	}

	// Get and parse JSON
	userStr, err := redisCache.Get(ctx, "user:alice")
	if err != nil {
		log.Printf("Failed to get user data: %v", err)
	} else {
		var retrievedUser map[string]interface{}
		json.Unmarshal([]byte(userStr), &retrievedUser)
		log.Printf("Got user data: %+v", retrievedUser)
	}

	// Check if key exists
	count, err := redisCache.Exists(ctx, "greeting", "user:alice")
	if err != nil {
		log.Printf("Failed to check existence: %v", err)
	} else {
		log.Printf("Keys exist: %d", count)
	}

	// Increment counter
	val1, err := redisCache.Increment(ctx, "page:views")
	if err != nil {
		log.Printf("Failed to increment: %v", err)
	} else {
		log.Printf("Page views: %d", val1)
	}

	// Set with expiration using SetNX (only if not exists)
	ok, err := redisCache.SetNX(ctx, "lock:resource", "locked", 30*time.Second)
	if err != nil {
		log.Printf("Failed to set lock: %v", err)
	} else {
		log.Printf("Lock acquired: %v", ok)
	}

	// Delete keys
	err = redisCache.Delete(ctx, "greeting")
	if err != nil {
		log.Printf("Failed to delete key: %v", err)
	} else {
		log.Println("Deleted key 'greeting'")
	}
}

// RabbitMQExample demonstrates RabbitMQ usage
func RabbitMQExample(rmq *queue.RabbitMQ) {
	log.Println("=== RabbitMQ Example ===")

	queueName := "example_queue"

	// Declare a queue
	_, err := rmq.DeclareQueue(queueName, true, false, false)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return
	}
	log.Printf("Declared queue: %s", queueName)

	// Publish messages
	for i := 1; i <= 3; i++ {
		msg := ExampleMessage{
			ID:      fmt.Sprintf("msg-%d", i),
			Message: fmt.Sprintf("Hello from message #%d", i),
			SentAt:  time.Now(),
		}
		msgJSON, _ := json.Marshal(msg)

		err = rmq.PublishToQueue(context.Background(), queueName, msgJSON)
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		} else {
			log.Printf("Published message: %s", msg.Message)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Consume messages
	log.Println("Consuming messages...")
	msgs, err := rmq.Consume(context.Background(), queueName, "example-consumer", false)
	if err != nil {
		log.Printf("Failed to consume: %v", err)
		return
	}

	// Process messages for 5 seconds
	timeout := time.After(5 * time.Second)
	count := 0
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Channel closed")
				return
			}
			var exMsg ExampleMessage
			json.Unmarshal(msg.Body, &exMsg)
			log.Printf("Received message: %+v", exMsg)
			msg.Ack(false)
			count++
			if count >= 3 {
				log.Println("Processed all messages")
				return
			}
		case <-timeout:
			log.Println("Timeout waiting for messages")
			return
		}
	}
}

// RunAllExamples runs all examples
func RunAllExamples() {
	log.Println("=== Running All Examples ===")

	// Load configuration
	cfg := config.Load()

	// PostgreSQL Example
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Printf("Skipping PostgreSQL example: %v", err)
	} else {
		defer db.Close()
		PostgresExample(db)
		fmt.Println()
	}

	// Redis Example
	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.Printf("Skipping Redis example: %v", err)
	} else {
		defer redisCache.Close()
		RedisExample(redisCache)
		fmt.Println()
	}

	// RabbitMQ Example
	rmq, err := queue.NewRabbitMQ(&cfg.RabbitMQ, nil)
	if err != nil {
		log.Printf("Skipping RabbitMQ example: %v", err)
	} else {
		defer rmq.Close()
		RabbitMQExample(rmq)
		fmt.Println()
	}

	log.Println("=== All Examples Completed ===")
}
