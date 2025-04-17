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

var db *gorm.DB

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=mallhive password=yourpassword dbname=mallhive_products port=5432 sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	// Migrate schema
	db.AutoMigrate(&Product{}, &Category{}, &Inventory{})

	r := gin.Default()
	r.Use(corsMiddleware())

	// Product routes
	r.GET("/products", listProducts)
	r.GET("/products/:slug", getProductBySlug)
	r.POST("/products", createProduct)

	// Category routes
	r.GET("/categories", listCategories)
	r.POST("/categories", createCategory)

	// Order forwarding
	r.POST("/orders", forwardToOrderService)

	// Run the server
	r.Run(":8080")
}

// -------- Product Handlers --------

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

	slug := slugify(input.Name)

	product := Product{
		Name:        input.Name,
		Slug:        slug,
		CategoryID:  input.CategoryID,
		Price:       input.Price,
		Available:   input.Available,
		Description: input.Description,
		ImageURL:    input.ImageURL,
	}

	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create product", "details": err.Error()})
		return
	}

	// Create inventory record
	inventory := Inventory{
		ProductID: product.ID,
		Quantity:  input.StockQty,
	}
	db.Create(&inventory)

	// Fetch full product with relations to return
	db.Preload("Category").Preload("Inventory").First(&product, product.ID)

	c.JSON(http.StatusCreated, product)
}

// -------- Category Handlers --------

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

	input.Slug = slugify(input.Name)

	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create category", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// -------- Order Forwarding --------

func forwardToOrderService(c *gin.Context) {
	resp, err := http.Post("http://mallhive.com/orders", "application/json", c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Order service unreachable"})
		return
	}
	defer resp.Body.Close()

	c.JSON(resp.StatusCode, gin.H{"message": "Order forwarded successfully"})
}

// -------- Helpers --------

func slugify(input string) string {
	return strings.ToLower(strings.ReplaceAll(input, " ", "-"))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // You can change "*" to your frontend domain
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
