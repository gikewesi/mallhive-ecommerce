package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/gin-gonic/gin"
)

// EventType values can include: "page_view", "product_click", "purchase", "trend_metric"
type Event struct {
	UserID      string  `json:"user_id"`
	SessionID   string  `json:"session_id"`
	EventType   string  `json:"event_type"`  // user_behavior | sales_metric | trend
	Description string  `json:"description"` // e.g. "clicked product", "completed checkout"
	ProductID   string  `json:"product_id"`  // if applicable
	Category    string  `json:"category"`    // e.g. electronics, fashion, etc.
	Value       float64 `json:"value"`       // e.g. cart total, time on page, etc.
	Timestamp   string  `json:"timestamp"`   // set by backend
}

// Set default env vars if none are set
func LoadEnvDefaults() {
	_ = os.Setenv("KINESIS_STREAM_NAME", "analytics-stream")
	_ = os.Setenv("AWS_REGION", "us-west-2")
}

func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func PublishEventToKinesis(event Event) error {
	streamName := os.Getenv("KINESIS_STREAM_NAME")
	region := os.Getenv("AWS_REGION")

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := kinesis.New(sess)

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = svc.PutRecord(&kinesis.PutRecordInput{
		Data:         data,
		StreamName:   aws.String(streamName),
		PartitionKey: aws.String(event.UserID),
	})

	return err
}

func LogEventHandler(c *gin.Context) {
	var event Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamp server-side
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)

	err := PublishEventToKinesis(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "event received"})
}

func main() {
	LoadEnvDefaults()

	r := gin.Default()

	r.POST("/event", LogEventHandler)

	port := GetEnv("PORT", "8080")
	log.Printf("Starting Analytics Service on port %s...", port)
	err := r.Run(":" + port)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
