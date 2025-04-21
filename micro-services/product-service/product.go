package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/opensearch-project/opensearch-go"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
}

var db *sql.DB
var searchClient *opensearch.Client

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Error loading .env file")
	}

	// Connect to PostgreSQL
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	))
	if err != nil {
		log.Fatal("❌ Database connection error:", err)
	}

	// Connect to OpenSearch
	searchClient, err = opensearch.NewClient(opensearch.Config{
		Addresses: []string{os.Getenv("OPENSEARCH_URL")},
	})
	if err != nil {
		log.Fatal("❌ OpenSearch connection error:", err)
	}

	// REST routes
	http.HandleFunc("/products", productHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, name, description, category, price, available FROM products")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.Price, &p.Available)
			products = append(products, p)
		}
		json.NewEncoder(w).Encode(products)

	case http.MethodPost:
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		err := db.QueryRow(`
            INSERT INTO products (name, description, category, price, available)
            VALUES ($1, $2, $3, $4, $5) RETURNING id
        `, p.Name, p.Description, p.Category, p.Price, p.Available).Scan(&p.ID)

		if err != nil {
			http.Error(w, "Database insert failed", http.StatusInternalServerError)
			return
		}

		// Send to OpenSearch
		IndexToOpenSearch(p)

		// Notify other services
		NotifyExternalServices(p)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func IndexToOpenSearch(p Product) {
	data, _ := json.Marshal(p)
	res, err := searchClient.Index(
		"products",
		bytes.NewReader(data),
		searchClient.Index.WithDocumentID(fmt.Sprint(p.ID)),
		searchClient.Index.WithContext(context.Background()),
	)
	if err != nil {
		log.Println("❌ OpenSearch index error:", err)
		return
	}
	defer res.Body.Close()
	log.Println("✅ Product indexed in OpenSearch")
}

func NotifyExternalServices(p Product) {
	cartURL := os.Getenv("CART_SERVICE_URL")
	orderURL := os.Getenv("ORDER_SERVICE_URL")

	payload, _ := json.Marshal(p)
	// Notify Shopping Cart
	http.Post(cartURL+"/products/sync", "application/json", bytes.NewReader(payload))
	// Notify Order Service
	http.Post(orderURL+"/products/sync", "application/json", bytes.NewReader(payload))

	log.Println("📡 Product sync notifications sent")
}
