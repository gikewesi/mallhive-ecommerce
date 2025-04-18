package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

var (
	rdb         *redis.Client
	ctx         = context.Background()
	snsClient   *sns.SNS
	snsTopicARN = os.Getenv("SNS_TOPIC_ARN")
)

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func initSNS() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	snsClient = sns.New(sess)
}

func addToCart(w http.ResponseWriter, r *http.Request) {
	var item CartItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
		http.Error(w, "Could not fetch cart", http.StatusInternalServerError)
		return
	}
	items := []CartItem{}
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

	items := []CartItem{}
	for _, val := range entries {
		var item CartItem
		json.Unmarshal([]byte(val), &item)
		items = append(items, item)
	}

	orderData, _ := json.Marshal(Cart{UserID: userID, Items: items})
	_, err = snsClient.Publish(&sns.PublishInput{
		Message:  aws.String(string(orderData)),
		TopicArn: aws.String(snsTopicARN),
	})
	if err != nil {
		http.Error(w, "Failed to publish checkout event", http.StatusInternalServerError)
		return
	}

	rdb.Del(ctx, key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Checkout complete and event sent."))
}

func main() {
	initRedis()
	initSNS()

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/cart/{user_id}", getCart).Methods("GET")
	api.HandleFunc("/cart/{user_id}", addToCart).Methods("POST")
	api.HandleFunc("/cart/{user_id}/checkout", checkout).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Shopping Cart service running on port 8080")
	log.Fatal(srv.ListenAndServe())
}
