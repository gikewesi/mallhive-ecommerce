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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/sqs"
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
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

type CartItem struct {
	ProductID int64   `json:"product_id"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type PaymentCallback struct {
	OrderID int    `json:"order_id"`
	Status  string `json:"status"` // "success", "failed", etc.
	Message string `json:"message,omitempty"`
}

// Order status constants
const (
	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusFailed   = "payment_failed"
	StatusComplete = "completed"
)

var (
	db                *sql.DB
	sqsClient         *sqs.SQS
	eventBridgeClient *eventbridge.EventBridge
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}

	connStr := os.Getenv("POSTGRES_CONN")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Unable to reach the database:", err)
	}

	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	})
	if err != nil {
		log.Fatal("Failed to create AWS session:", err)
	}
	sqsClient = sqs.New(sess)
	eventBridgeClient = eventbridge.New(sess)
}

func main() {
	http.HandleFunc("/orders/", ordersHandler)
	http.HandleFunc("/orders/callback", paymentCallbackHandler)
	log.Println("Order service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/orders/")

	switch {
	case r.Method == http.MethodPost:
		handleCreateOrder(w, r)
	case r.Method == http.MethodGet && path == "":
		handleListOrders(w, r)
	case r.Method == http.MethodGet && path != "":
		handleGetOrder(w, r, path)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if order.UserID == 0 {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	// Fetch and validate cart
	cartItems, err := fetchCartItems(order.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate products
	if err := validateProducts(cartItems); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate order details
	order.ProductIDs, order.Total = calculateOrderDetails(cartItems)
	order.Status = StatusPending

	// Save to database
	if err := saveOrderToDB(&order); err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Async integrations
	go processPostOrderActions(order)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func handleGetOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	id, err := strconv.Atoi(orderID)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order Order
	var productIDs []int64

	query := `SELECT id, user_id, product_ids, total, status, created_at, updated_at 
			  FROM orders WHERE id = $1`

	row := db.QueryRow(query, id)
	err = row.Scan(&order.ID, &order.UserID, pq.Array(&productIDs),
		&order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	order.ProductIDs = productIDs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func paymentCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var callback PaymentCallback
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if callback.OrderID == 0 {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	// Update order status based on payment result
	newStatus := StatusPending
	if callback.Status == "success" {
		newStatus = StatusPaid
	} else if callback.Status == "failed" {
		newStatus = StatusFailed
	}

	// Update the order in the database
	err := updateOrderStatus(callback.OrderID, newStatus)
	if err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		log.Printf("Failed to update order status: %v", err)
		return
	}

	// Send notification about order status update
	go sendOrderStatusUpdate(callback.OrderID, newStatus)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func updateOrderStatus(orderID int, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := db.Exec(query, status, time.Now(), orderID)
	return err
}

func sendOrderStatusUpdate(orderID int, status string) {
	var order Order
	var productIDs []int64

	query := `SELECT id, user_id, product_ids, total, status FROM orders WHERE id = $1`
	err := db.QueryRow(query, orderID).Scan(&order.ID, &order.UserID,
		pq.Array(&productIDs), &order.Total, &order.Status)
	if err != nil {
		log.Printf("Error fetching order for notification: %v", err)
		return
	}
	order.ProductIDs = productIDs

	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":  order.UserID,
		"order_id": order.ID,
		"status":   status,
		"message":  fmt.Sprintf("Your order #%d is now %s", order.ID, status),
	})

	resp, err := http.Post(notificationURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Notification service error: %v", err)
		return
	}
	defer resp.Body.Close()

	sendStatusUpdateToEventBridge(order, status)
}

func sendStatusUpdateToEventBridge(order Order, status string) {
	eventBus := os.Getenv("EVENT_BUS_NAME")
	detail, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID,
		"status":   status,
		"user_id":  order.UserID,
	})

	_, err := eventBridgeClient.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{{
			Source:       aws.String("order-service"),
			Detail:       aws.String(string(detail)),
			DetailType:   aws.String("OrderStatusChanged"),
			EventBusName: aws.String(eventBus),
		}},
	})
	if err != nil {
		log.Printf("EventBridge error: %v", err)
	}
}

func fetchCartItems(userID int) ([]CartItem, error) {
	cartServiceURL := os.Getenv("CART_SERVICE_URL")
	resp, err := http.Get(fmt.Sprintf("%s/%d", cartServiceURL, userID))
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch cart data")
	}
	defer resp.Body.Close()

	var cartItems []CartItem
	if err := json.NewDecoder(resp.Body).Decode(&cartItems); err != nil {
		return nil, fmt.Errorf("failed to parse cart data")
	}

	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	return cartItems, nil
}

func validateProducts(cartItems []CartItem) error {
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")

	for _, item := range cartItems {
		resp, err := http.Get(fmt.Sprintf("%s/%d", productServiceURL, item.ProductID))
		if err != nil || resp.StatusCode != http.StatusOK {
			return fmt.Errorf("invalid product ID: %d", item.ProductID)
		}
		resp.Body.Close()
	}
	return nil
}

func calculateOrderDetails(cartItems []CartItem) ([]int64, float64) {
	var total float64
	var productIDs []int64

	for _, item := range cartItems {
		total += item.Price * float64(item.Quantity)
		productIDs = append(productIDs, item.ProductID)
	}
	return productIDs, total
}

func saveOrderToDB(order *Order) error {
	query := `INSERT INTO orders (user_id, product_ids, total, status, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $5) RETURNING id, created_at`
	return db.QueryRow(
		query,
		order.UserID,
		pq.Array(order.ProductIDs),
		order.Total,
		order.Status,
		time.Now(),
	).Scan(&order.ID, &order.CreatedAt)
}

func processPostOrderActions(order Order) {
	sendToPaymentService(order)
	sendToNotificationService(order)
	sendToSQS(order)
	sendToEventBridge(order)
}

func sendToPaymentService(order Order) {
	paymentURL := os.Getenv("PAYMENT_SERVICE_URL")
	host := os.Getenv("ORDER_SERVICE_HOST")
	if host == "" {
		host = "http://localhost:8080"
	}
	callbackURL := fmt.Sprintf("%s/orders/callback", host)

	body, _ := json.Marshal(map[string]interface{}{
		"order_id":     order.ID,
		"amount":       order.Total,
		"user_id":      order.UserID,
		"callback_url": callbackURL,
	})

	resp, err := http.Post(paymentURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Payment service error: %v", err)
		return
	}
	defer resp.Body.Close()
}

func sendToNotificationService(order Order) {
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":  order.UserID,
		"order_id": order.ID,
		"total":    order.Total,
		"status":   order.Status,
		"message":  fmt.Sprintf("New order #%d created", order.ID),
	})

	resp, err := http.Post(notificationURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Notification service error: %v", err)
		return
	}
	defer resp.Body.Close()
}

func sendToSQS(order Order) {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	body, _ := json.Marshal(order)

	_, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(body)),
		QueueUrl:    aws.String(queueURL),
	})
	if err != nil {
		log.Printf("SQS send error: %v", err)
	}
}

func sendToEventBridge(order Order) {
	eventBus := os.Getenv("EVENT_BUS_NAME")
	detail, _ := json.Marshal(order)
	_, err := eventBridgeClient.PutEvents(&eventbridge.PutEventsInput{
		Entries: []*eventbridge.PutEventsRequestEntry{{
			Source:       aws.String("order-service"),
			Detail:       aws.String(string(detail)),
			DetailType:   aws.String("OrderCreated"),
			EventBusName: aws.String(eventBus),
		}},
	})
	if err != nil {
		log.Printf("EventBridge error: %v", err)
	}
}

func handleListOrders(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query("SELECT id, user_id, product_ids, total, status, created_at, updated_at FROM orders ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		var productIDs []int64
		if err := rows.Scan(&order.ID, &order.UserID, pq.Array(&productIDs), &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		order.ProductIDs = productIDs
		orders = append(orders, order)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
