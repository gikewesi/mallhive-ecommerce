package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type CartItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

type Product struct {
	ID    string  `json:"id"`
	Price float64 `json:"price"`
}

var (
	rdb              *redis.Client
	ctx              = context.Background()
	snsClient        *sns.SNS
	snsTopicARN      = ""
	productSvcURL    = ""
	orderSvcEndpoint = ""
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	snsTopicARN = os.Getenv("SNS_TOPIC_ARN")
	productSvcURL = os.Getenv("PRODUCT_SERVICE_URL")
	orderSvcEndpoint = os.Getenv("ORDER_SERVICE_URL")

	// Initialize Redis and SNS client
	initRedis()
	initSNS()
}

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func initSNS() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	snsClient = sns.New(sess)
}

func getAllowedOrigins() []string {
	origins := os.Getenv("CORS_ORIGINS")
	return strings.Split(origins, ",")
}

func getProductDetails(productID string) (*Product, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", productSvcURL, productID))
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch product: %s", productID)
	}
	defer resp.Body.Close()

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

func addToCart(w http.ResponseWriter, r *http.Request) {
	var item CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	product, err := getProductDetails(item.ProductID)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	item.Price = product.Price
	userID := mux.Vars(r)["user_id"]
	key := fmt.Sprintf("cart:%s", userID)

	itemID := uuid.New().String()
	itemBytes, _ := json.Marshal(item)
	rdb.HSet(ctx, key, itemID, itemBytes)
	w.WriteHeader(http.StatusCreated)
}

func getCart(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]
	key := fmt.Sprintf("cart:%s", userID)

	entries, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}

	var items []CartItem
	for _, val := range entries {
		var item CartItem
		json.Unmarshal([]byte(val), &item)
		items = append(items, item)
	}
	json.NewEncoder(w).Encode(Cart{UserID: userID, Items: items})
}

func checkout(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]
	key := fmt.Sprintf("cart:%s", userID)

	entries, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		http.Error(w, "Could not checkout", http.StatusInternalServerError)
		return
	}

	var items []CartItem
	for _, val := range entries {
		var item CartItem
		json.Unmarshal([]byte(val), &item)
		items = append(items, item)
	}

	cart := Cart{UserID: userID, Items: items}
	orderPayload, _ := json.Marshal(cart)

	// Send order to Order Microservice
	resp, err := http.Post(orderSvcEndpoint, "application/json", bytes.NewBuffer(orderPayload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		http.Error(w, "Failed to place order", http.StatusInternalServerError)
		return
	}

	// Optional: publish to SNS topic
	snsClient.Publish(&sns.PublishInput{
		Message:  aws.String(string(orderPayload)),
		TopicArn: aws.String(snsTopicARN),
	})

	rdb.Del(ctx, key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order placed successfully."))
}

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/cart/{user_id}", addToCart).Methods("POST")
	api.HandleFunc("/cart/{user_id}", getCart).Methods("GET")
	api.HandleFunc("/cart/{user_id}/checkout", checkout).Methods("POST")

	// Integrate CORS
	handler := cors.New(cors.Options{
		AllowedOrigins: getAllowedOrigins(),
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}).Handler(r)

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Shopping Cart Service running on port 8080")
	log.Fatal(srv.ListenAndServe())
}
