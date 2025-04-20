package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"log"
	"net/http"
	"os"
	"strings"
)

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Slug        string    `gorm:"unique;not null" json:"slug"`
	CategoryID  uint      `json:"categoryId"`
	Price       float64   `json:"price"`
	Available   bool      `json:"available"`
	Description string    `json:"description"`
	ImageURL    string    `json:"imageURL"`
	Category    Category  `json:"category"`
	Inventory   Inventory `json:"inventory"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
	Slug string `gorm:"unique;not null" json:"slug"`
}

type Inventory struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	ProductID uint `json:"productId"`
	Quantity  int  `json:"stockQuantity"`
}

var (
	db              *gorm.DB
	orderServiceURL string
)

func main() {
	_ = godotenv.Load()

	// Load environment variables
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not found in .env file")
	}

	orderServiceURL = os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		log.Fatal("ORDER_SERVICE_URL not found in .env file")
	}

	// Initialize the database connection
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	// Automatically migrate the database schema
	db.AutoMigrate(&Product{}, &Category{}, &Inventory{})

	// Set up Gin router
	r := gin.Default()
	r.Use(corsMiddleware())

	// Versioned API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/products", listProducts)           // GET all products
		v1.GET("/products/:slug", getProductBySlug) // GET product by slug
		v1.POST("/products", createProduct)         // POST create new product

		v1.GET("/categories", listCategories)  // GET all categories
		v1.POST("/categories", createCategory) // POST create new category

		v1.POST("/orders", forwardToOrderService) // POST order request to Order service
	}

	// Run the server on port 8080
	r.Run(":8080")
}

// -------- Handlers --------

func listProducts(c *gin.Context) {
	var products []Product
	db.Preload("Category").Preload("Inventory").Find(&products)
	c.JSON(http.StatusOK, products)
}

func getProductBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var product Product
	result := db.Preload("Category").Preload("Inventory").Where("slug = ?", slug).First(&product)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func createProduct(c *gin.Context) {
	var input struct {
		Name        string  `json:"name" binding:"required"`
		CategoryID  uint    `json:"categoryId" binding:"required"`
		Price       float64 `json:"price" binding:"required"`
		Available   bool    `json:"available"`
		Description string  `json:"description"`
		ImageURL    string  `json:"imageURL"`
		StockQty    int     `json:"stockQuantity"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate slug from product name
	slug := slugify(input.Name)

	// Create the new product record
	product := Product{
		Name:        input.Name,
		Slug:        slug,
		CategoryID:  input.CategoryID,
		Price:       input.Price,
		Available:   input.Available,
		Description: input.Description,
		ImageURL:    input.ImageURL,
	}

	// Insert the product into the database
	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create product", "details": err.Error()})
		return
	}

	// Create the corresponding inventory record
	inventory := Inventory{
		ProductID: product.ID,
		Quantity:  input.StockQty,
	}
	db.Create(&inventory)

	// Fetch the full product with category and inventory details
	db.Preload("Category").Preload("Inventory").First(&product, product.ID)

	// Return the created product
	c.JSON(http.StatusCreated, product)
}

func listCategories(c *gin.Context) {
	var categories []Category
	db.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

func createCategory(c *gin.Context) {
	var input Category
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate slug from category name
	input.Slug = slugify(input.Name)

	// Create the category in the database
	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create category", "details": err.Error()})
		return
	}

	// Return the created category
	c.JSON(http.StatusCreated, input)
}

func forwardToOrderService(c *gin.Context) {
	resp, err := http.Post(orderServiceURL, "application/json", c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Order service unreachable"})
		return
	}
	defer resp.Body.Close()
	c.JSON(resp.StatusCode, gin.H{"message": "Order forwarded successfully"})
}

// -------- Helpers --------

// Slugify a string (convert to lowercase, replace spaces with dashes)
func slugify(input string) string {
	return strings.ToLower(strings.ReplaceAll(input, " ", "-"))
}

// CORS middleware to handle cross-origin requests
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigins := os.Getenv("CORS_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:3000,https://your-microfrontend-domain.com" // default if not set
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
