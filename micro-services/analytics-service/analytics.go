package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Event struct {
	EventID       string    `json:"event_id"`
	UserID        string    `json:"user_id" validate:"required"`
	SessionID     string    `json:"session_id"`
	EventType     string    `json:"event_type" validate:"required,eventtype"`
	EventSubtype  string    `json:"event_subtype"`
	SourceService string    `json:"source_service"`
	DeviceType    string    `json:"device_type"`
	Location      string    `json:"location"`
	ProductID     string    `json:"product_id,omitempty"`
	Category      string    `json:"category,omitempty"`
	Value         float64   `json:"value,omitempty"`
	Metadata      Metadata  `json:"metadata,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

type Metadata struct {
	Browser     string `json:"browser,omitempty"`
	OS          string `json:"os,omitempty"`
	ScreenRes   string `json:"screen_res,omitempty"`
	IPAddress   string `json:"ip_address,omitempty" validate:"omitempty,ip"`
	UserAgent   string `json:"user_agent,omitempty"`
	Referrer    string `json:"referrer,omitempty"`
	UTMSource   string `json:"utm_source,omitempty"`
	UTMMedium   string `json:"utm_medium,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`
}

var (
	kinesisClient *kinesis.Kinesis
	streamName    string
	rateLimiter   *rate.Limiter
	apiKey        string
	logger        *zap.Logger
	batch         []*kinesis.PutRecordsRequestEntry
	batchMutex    = &sync.Mutex{}
	startTime     = time.Now()
)

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// AWS configuration
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(getEnv("AWS_REGION", "us-west-2")),
	}))
	kinesisClient = kinesis.New(sess)
	streamName = getEnv("KINESIS_STREAM_NAME", "analytics-events")

	// Rate limiting
	rateLimit := rate.Limit(getEnvAsFloat("RATE_LIMIT", 10.0))
	burstSize := getEnvAsInt("RATE_BURST", 30)
	rateLimiter = rate.NewLimiter(rateLimit, burstSize)

	// Security
	apiKey = getAPIKey()

	// Start batch processor
	go processBatch()
}

func main() {
	router := gin.Default()

	// Middleware
	router.Use(rateLimitMiddleware())
	router.Use(authMiddleware())
	router.MaxMultipartMemory = int64(getEnvAsInt("MAX_BODY_SIZE", 10)) << 20 // MB to bytes

	// Routes
	router.GET("/health", healthCheck)
	router.POST("/events", handleEvent)

	// Start server
	port := getEnv("PORT", "8080")
	logger.Info("Starting analytics service", zap.String("port", port))

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"version":  "1.0.0",
		"uptime":   time.Since(startTime).String(),
		"stream":   streamName,
		"requests": rateLimiter.Limit(),
	})
}

func handleEvent(c *gin.Context) {
	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Warn("Invalid request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse("invalid request payload"))
		return
	}

	// Validate event
	if !isValidEvent(event) {
		logger.Warn("Validation failed", zap.Any("event", event))
		c.JSON(http.StatusBadRequest, errorResponse("invalid event data"))
		return
	}

	// Enrich event
	event.EventID = xid.New().String()
	event.Timestamp = time.Now().UTC()
	if event.SourceService == "" {
		event.SourceService = getEnv("DEFAULT_SOURCE", "unknown")
	}

	// Add to batch
	if err := addToBatch(event); err != nil {
		logger.Error("Failed to queue event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse("failed to process event"))
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"event_id":  event.EventID,
		"status":    "queued",
		"timestamp": event.Timestamp,
	})
}

func isValidEvent(event Event) bool {
	validTypes := map[string]bool{
		"user_behavior": true,
		"sales_metric":  true,
		"system_metric": true,
	}

	return event.UserID != "" &&
		validTypes[event.EventType] &&
		!strings.ContainsAny(event.UserID, "<>{}'\"") &&
		len(event.UserID) < 256
}

func addToBatch(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	batchMutex.Lock()
	defer batchMutex.Unlock()

	batch = append(batch, &kinesis.PutRecordsRequestEntry{
		Data:         data,
		PartitionKey: aws.String(event.EventID),
	})

	return nil
}

func processBatch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		batchMutex.Lock()
		if len(batch) == 0 {
			batchMutex.Unlock()
			continue
		}

		currentBatch := batch
		batch = make([]*kinesis.PutRecordsRequestEntry, 0)
		batchMutex.Unlock()

		_, err := kinesisClient.PutRecords(&kinesis.PutRecordsInput{
			Records:    currentBatch,
			StreamName: aws.String(streamName),
		})

		if err != nil {
			logger.Error("Failed to send batch",
				zap.Int("count", len(currentBatch)),
				zap.Error(err))
		} else {
			logger.Info("Batch processed",
				zap.Int("events", len(currentBatch)))
		}
	}
}

// Helper functions
func getAPIKey() string {
	if key := os.Getenv("API_KEY"); key != "" {
		return key
	}
	return fetchFromAWSSecretsManager()
}

func fetchFromAWSSecretsManager() string {
	// Implement AWS Secrets Manager integration
	return "default-secret-key"
}

func errorResponse(message string) gin.H {
	return gin.H{"error": message, "ts": time.Now().UTC()}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvAsFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}

// Middleware
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, errorResponse("rate limit exceeded"))
			c.Abort()
		}
		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token != apiKey {
			logger.Warn("Unauthorized access attempt",
				zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
			c.Abort()
		}
		c.Next()
	}
}
