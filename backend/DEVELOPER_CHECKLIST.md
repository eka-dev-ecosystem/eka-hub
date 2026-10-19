# Repository Pattern - Developer Checklist

## ✅ Implementation Checklist

Use this checklist when adding new features to your application.

---

## 📋 Adding a New Entity (e.g., "Products")

### Step 1: Create Migration

```bash
cd migrations
# Create migration files
echo. > 000004_create_products_table.up.sql
echo. > 000004_create_products_table.down.sql
```

**000004_create_products_table.up.sql:**

```sql
CREATE TABLE products
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255)   NOT NULL,
    description TEXT,
    price       DECIMAL(10, 2) NOT NULL,
    stock       INT              DEFAULT 0,
    is_active   BOOLEAN          DEFAULT true,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_name ON products (name);
CREATE INDEX idx_products_is_active ON products (is_active);
```

**000004_create_products_table.down.sql:**

```sql
DROP TABLE IF EXISTS products;
```

### Step 2: Run Migration

```bash
make migrate-up
# or
./migrate.bat up
```

---

### Step 3: Create Model

**File:** `pkg/models/product.go`

```go
package models

import "time"

// Product represents a product in the system
type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProductRequest validation
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Description string  `json:"description" validate:"omitempty,max=1000"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"omitempty,gte=0"`
}

// UpdateProductRequest validation
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=3,max=255"`
	Description string  `json:"description" validate:"omitempty,max=1000"`
	Price       float64 `json:"price" validate:"omitempty,gt=0"`
	Stock       int     `json:"stock" validate:"omitempty,gte=0"`
}

// ProductResponse is the public response for product data
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
```

**Checklist:**

- [ ] Model struct with db tags
- [ ] CreateRequest with validation tags
- [ ] UpdateRequest with validation tags
- [ ] Response struct (without sensitive fields)
- [ ] ToResponse() method

---

### Step 4: Create Repository

**File:** `pkg/repository/product_repository.go`

```go
package repository

import (
	"database/sql"
	"fmt"
	"golang/pkg/models"

	"github.com/jmoiron/sqlx"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	GetAll() ([]models.Product, error)
	GetByID(id string) (*models.Product, error)
	Create(product *models.Product) error
	Update(id string, updates map[string]interface{}) (*models.Product, error)
	Delete(id string) error
	ExistsByID(id string) (bool, error)
}

// productRepository implements ProductRepository
type productRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db: db}
}

// GetAll retrieves all active products
func (r *productRepository) GetAll() ([]models.Product, error) {
	var products []models.Product
	query := "SELECT * FROM products WHERE is_active = true ORDER BY name"
	err := r.db.Select(&products, query)
	return products, err
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(id string) (*models.Product, error) {
	var product models.Product
	query := "SELECT * FROM products WHERE id = $1"
	err := r.db.Get(&product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// Create creates a new product
func (r *productRepository) Create(product *models.Product) error {
	query := `
        INSERT INTO products (id, name, description, price, stock, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.Exec(query, product.ID, product.Name, product.Description,
		product.Price, product.Stock, product.IsActive,
		product.CreatedAt, product.UpdatedAt)
	return err
}

// Update updates an existing product
func (r *productRepository) Update(id string, updates map[string]interface{}) (*models.Product, error) {
	query := "UPDATE products SET updated_at = NOW()"
	args := []interface{}{id}
	argPosition := 2

	if name, ok := updates["name"].(string); ok && name != "" {
		query += fmt.Sprintf(", name = $%d", argPosition)
		args = append(args, name)
		argPosition++
	}

	if description, ok := updates["description"].(string); ok {
		query += fmt.Sprintf(", description = $%d", argPosition)
		args = append(args, description)
		argPosition++
	}

	if price, ok := updates["price"].(float64); ok && price > 0 {
		query += fmt.Sprintf(", price = $%d", argPosition)
		args = append(args, price)
		argPosition++
	}

	if stock, ok := updates["stock"].(int); ok {
		query += fmt.Sprintf(", stock = $%d", argPosition)
		args = append(args, stock)
		argPosition++
	}

	query += " WHERE id = $1 RETURNING *"

	var product models.Product
	err := r.db.QueryRowx(query, args...).StructScan(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// Delete soft deletes a product
func (r *productRepository) Delete(id string) error {
	query := "UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ExistsByID checks if a product exists by ID
func (r *productRepository) ExistsByID(id string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)"
	err := r.db.Get(&exists, query, id)
	return exists, err
}
```

**Checklist:**

- [ ] Interface definition with all CRUD methods
- [ ] Struct implementation
- [ ] Constructor function (NewXxxRepository)
- [ ] GetAll() method
- [ ] GetByID() method
- [ ] Create() method
- [ ] Update() method with dynamic fields
- [ ] Delete() method (soft delete)
- [ ] Helper methods (ExistsByID, etc.)

---

### Step 5: Create Service

**File:** `pkg/services/product_service.go`

```go
package services

import (
	"database/sql"
	"time"

	"golang/pkg/models"
	"golang/pkg/repository"
	customErrors "golang/pkg/errors"

	"github.com/google/uuid"
)

// ProductService defines the business logic for product operations
type ProductService interface {
	GetAllProducts() ([]models.ProductResponse, error)
	GetProductByID(id string) (*models.ProductResponse, error)
	CreateProduct(req *models.CreateProductRequest) (*models.ProductResponse, error)
	UpdateProduct(id string, req *models.UpdateProductRequest) (*models.ProductResponse, error)
	DeleteProduct(id string) error
}

// productService implements ProductService
type productService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

// GetAllProducts retrieves all active products
func (s *productService) GetAllProducts() ([]models.ProductResponse, error) {
	products, err := s.repo.GetAll()
	if err != nil {
		return nil, customErrors.DatabaseError("Failed to fetch products", err)
	}

	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

// GetProductByID retrieves a product by ID
func (s *productService) GetProductByID(id string) (*models.ProductResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, customErrors.BadRequest("Invalid product ID format")
	}

	product, err := s.repo.GetByID(id)
	if err != nil {
		return nil, customErrors.DatabaseError("Failed to fetch product", err)
	}

	if product == nil {
		return nil, customErrors.NotFound("Product not found")
	}

	response := product.ToResponse()
	return &response, nil
}

// CreateProduct creates a new product
func (s *productService) CreateProduct(req *models.CreateProductRequest) (*models.ProductResponse, error) {
	now := time.Now()
	product := &models.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.repo.Create(product)
	if err != nil {
		return nil, customErrors.ParseDatabaseError(err)
	}

	response := product.ToResponse()
	return &response, nil
}

// UpdateProduct updates an existing product
func (s *productService) UpdateProduct(id string, req *models.UpdateProductRequest) (*models.ProductResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, customErrors.BadRequest("Invalid product ID format")
	}

	exists, err := s.repo.ExistsByID(id)
	if err != nil {
		return nil, customErrors.DatabaseError("Failed to check product existence", err)
	}
	if !exists {
		return nil, customErrors.NotFound("Product not found")
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.Stock >= 0 {
		updates["stock"] = req.Stock
	}

	product, err := s.repo.Update(id, updates)
	if err != nil {
		return nil, customErrors.ParseDatabaseError(err)
	}

	response := product.ToResponse()
	return &response, nil
}

// DeleteProduct soft deletes a product
func (s *productService) DeleteProduct(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return customErrors.BadRequest("Invalid product ID format")
	}

	err := s.repo.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return customErrors.NotFound("Product not found")
		}
		return customErrors.ParseDatabaseError(err)
	}

	return nil
}
```

**Checklist:**

- [ ] Interface definition
- [ ] Struct implementation
- [ ] Constructor function (NewXxxService)
- [ ] GetAll with transformation to Response
- [ ] GetByID with validation
- [ ] Create with business logic (UUID, defaults)
- [ ] Update with partial updates
- [ ] Delete with validation
- [ ] Proper error handling

---

### Step 6: Create Handler

**File:** `pkg/handlers/product_handler.go`

```go
package handlers

import (
	"golang/pkg/models"
	"golang/pkg/response"
	"golang/pkg/services"
	customErrors "golang/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service services.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// GetProducts returns all products
func (h *ProductHandler) GetProducts() fiber.Handler {
	return func(c *fiber.Ctx) error {
		products, err := h.service.GetAllProducts()
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, products, "Products retrieved successfully")
	}
}

// GetProduct returns a single product by ID
func (h *ProductHandler) GetProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		product, err := h.service.GetProductByID(id)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, product, "Product retrieved successfully")
	}
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CreateProductRequest

		if err := c.BodyParser(&req); err != nil {
			return customErrors.BadRequest("Invalid request body")
		}

		if err := validate.Struct(req); err != nil {
			return err
		}

		product, err := h.service.CreateProduct(&req)
		if err != nil {
			return err
		}

		return response.CreatedResponse(c, product, "Product created successfully")
	}
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req models.UpdateProductRequest
		if err := c.BodyParser(&req); err != nil {
			return customErrors.BadRequest("Invalid request body")
		}

		if err := validate.Struct(req); err != nil {
			return err
		}

		product, err := h.service.UpdateProduct(id, &req)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, product, "Product updated successfully")
	}
}

// DeleteProduct soft deletes a product
func (h *ProductHandler) DeleteProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		err := h.service.DeleteProduct(id)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, fiber.Map{}, "Product deleted successfully")
	}
}
```

**Checklist:**

- [ ] Struct with service dependency
- [ ] Constructor function (NewXxxHandler)
- [ ] GetAll handler
- [ ] GetByID handler
- [ ] Create handler with parsing & validation
- [ ] Update handler with parsing & validation
- [ ] Delete handler
- [ ] Proper error handling
- [ ] Standard response format

---

### Step 7: Wire Up Routes

**File:** `cmd/server/main.go`

```go
// Add import
import (
// ... existing imports ...
"golang/pkg/handlers"
"golang/pkg/repository"
"golang/pkg/services"
)

// Add setup function
func setupProductRoutes(api fiber.Router, db *database.PostgresDB) {
// Layer 1: Repository
productRepo := repository.NewProductRepository(db.GetDB())

// Layer 2: Service
productService := services.NewProductService(productRepo)

// Layer 3: Handler
productHandler := handlers.NewProductHandler(productService)

// Layer 4: Routes
products := api.Group("/products")
{
products.Get("/", productHandler.GetProducts())
products.Get("/:id", productHandler.GetProduct())
products.Post("/", productHandler.CreateProduct())
products.Put("/:id", productHandler.UpdateProduct())
products.Delete("/:id", productHandler.DeleteProduct())
}
}

// In main() function, add:
func main() {
// ... existing code ...

setupUserRoutes(api, db)
setupProductRoutes(api, db) // Add this line

// ... existing code ...
}
```

**Checklist:**

- [ ] Create setupXxxRoutes function
- [ ] Wire up all layers (Repository → Service → Handler)
- [ ] Define all routes
- [ ] Add to main() function
- [ ] Optional: Add authentication middleware

---

## 🧪 Step 8: Test the Implementation

```bash
# Build
go build -o bin/server.exe ./cmd/server

# Run
./bin/server.exe

# Test endpoints
curl -X POST http://localhost:3000/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "stock": 50
  }'

curl http://localhost:3000/api/v1/products
```

**Checklist:**

- [ ] Build succeeds
- [ ] Server starts without errors
- [ ] Can create entity
- [ ] Can retrieve all entities
- [ ] Can retrieve by ID
- [ ] Can update entity
- [ ] Can delete entity
- [ ] Validation works
- [ ] Error handling works
- [ ] Response format is correct

---

## 📝 Summary Checklist

When adding a new feature:

- [ ] Create migration files (up & down)
- [ ] Run migration
- [ ] Create model with validation
- [ ] Create repository interface & implementation
- [ ] Create service interface & implementation
- [ ] Create handler
- [ ] Wire up in main.go
- [ ] Test all endpoints
- [ ] Check error handling
- [ ] Verify response format

---

## 🎯 Best Practices

- ✅ Always use interfaces for repositories and services
- ✅ Keep business logic in service layer
- ✅ Keep database logic in repository layer
- ✅ Use validation tags on request structs
- ✅ Return Response types from services (hide sensitive data)
- ✅ Handle errors at appropriate layer
- ✅ Use transactions when needed
- ✅ Add indexes for commonly queried fields
- ✅ Use soft deletes (is_active flag)
- ✅ Always validate UUIDs before database queries

---

## 🚫 Common Mistakes to Avoid

- ❌ Don't skip layers (Handler → Repository directly)
- ❌ Don't put business logic in handlers
- ❌ Don't put SQL queries in services
- ❌ Don't expose database models directly in API responses
- ❌ Don't forget to validate input
- ❌ Don't ignore errors
- ❌ Don't hardcode values
- ❌ Don't forget to add migrations
- ❌ Don't forget timestamps (created_at, updated_at)
- ❌ Don't return passwords or sensitive fields

---

## 🎉 Done!

Your feature is now implemented following the repository pattern!
