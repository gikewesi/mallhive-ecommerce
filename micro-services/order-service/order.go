package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ProductIDs []int64   `json:"product_ids"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

var db *sql.DB

func init() {
	_ = godotenv.Load()

	connStr := os.Getenv("POSTGRES_CONN")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=postgres password=postgres dbname=orders_db sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Unable to reach the database:", err)
	}
}

func main() {
	http.HandleFunc("/orders/", ordersHandler)
	log.Println("Order service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateOrder(w, r)
	case http.MethodGet:
		handleListOrders(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if order.UserID == 0 || len(order.ProductIDs) == 0 || order.Total <= 0 {
		http.Error(w, "Missing or invalid order data", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO orders (user_id, product_ids, total, status, created_at)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err = db.QueryRow(
		query,
		order.UserID,
		pq.Array(order.ProductIDs),
		order.Total,
		"pending",
		time.Now(),
	).Scan(&order.ID, &order.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		log.Println("Insert error:", err)
		return
	}

	order.Status = "pending"
	go sendToPaymentService(order)
	go sendToNotificationService(order)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func handleListOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, user_id, product_ids, total, status, created_at FROM orders ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		var productIDs []int64
		err := rows.Scan(&o.ID, &o.UserID, pq.Array(&productIDs), &o.Total, &o.Status, &o.CreatedAt)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}
		o.ProductIDs = productIDs
		orders = append(orders, o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func sendToPaymentService(order Order) {
	paymentURL := "http://payment-service/process"

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID,
		"user_id":  order.UserID,
		"amount":   order.Total,
	})

	resp, err := http.Post(paymentURL, "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Payment service error for order %d: %v", order.ID, err)
	}
}

// Send a notification to the Notification Service
func sendToNotificationService(order Order) {
	notificationURL := "http://notification-service/notifications/send"

	message := fmt.Sprintf("Your order %d has been placed. Total amount: $%.2f", order.ID, order.Total)
	payload, _ := json.Marshal(map[string]interface{}{
		"user_id":  order.UserID,
		"order_id": order.ID,
		"message":  message,
	})

	resp, err := http.Post(notificationURL, "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Notification service error for order %d: %v", order.ID, err)
	}
}
