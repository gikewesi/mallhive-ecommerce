package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database connection
var db *sql.DB

// Product struct for JSON responses
type Product struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Price        float64 `json:"price"`
	Availability bool    `json:"availability"`
}

// Category struct for JSON responses
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Initialize database connection
func initDB() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("❌ Database unreachable:", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")
}

// Fetch all products
func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, category, price, availability FROM products")
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.Availability); err != nil {
			http.Error(w, "Error scanning product data", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// Fetch all categories
func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			http.Error(w, "Error scanning category data", http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// Update product inventory availability
func updateAvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Availability bool `json:"availability"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE products SET availability=$1 WHERE id=$2", req.Availability, id)
	if err != nil || result == nil {
		http.Error(w, "Failed to update availability", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "✅ Product availability updated")
}

// Send order request to Order Service
func sendOrderRequest(productID int) error {
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL") // e.g., "http://order-service/orders"
	orderData := map[string]int{"product_id": productID}

	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return fmt.Errorf("failed to marshal order data: %v", err)
	}

	resp, err := http.Post(orderServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send order request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("order service returned error: %d", resp.StatusCode)
	}

	fmt.Println("✅ Order request sent successfully!")
	return nil
}

// Handle order requests for products
func orderProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = sendOrderRequest(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send order request: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Order request sent for product %d", id)
}

// Main function to start the server
func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/api/products", getProductsHandler).Methods("GET")                          // Fetch all products
	r.HandleFunc("/api/categories", getCategoriesHandler).Methods("GET")                      // Fetch all categories
	r.HandleFunc("/api/products/{id}/availability", updateAvailabilityHandler).Methods("PUT") // Update availability
	r.HandleFunc("/api/products/{id}/order", orderProductHandler).Methods("POST")             // Send order request

	fmt.Println("🚀 Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
