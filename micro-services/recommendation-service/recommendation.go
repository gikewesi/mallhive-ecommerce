package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc"
)

var (
	ctx              = context.Background()
	redisClient      *redis.Client
	opensearchClient *elastic.Client
)

// Initialize Redis client
func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

// Initialize OpenSearch client (using elastic client for simplicity)
func initOpenSearch() {
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.Fatalf("Error creating OpenSearch client: %v", err)
	}
	opensearchClient = client
}

// Fetch user profile from User Profile Service (mocked)
func getUserProfile(userID string) map[string]string {
	// Mocked profile data
	return map[string]string{"age": "30", "gender": "female", "location": "NY"}
}

// Search products using OpenSearch
func searchProducts(query string) []string {
	if cached, err := redisClient.Get(ctx, "search:"+query).Result(); err == nil {
		var result []string
		json.Unmarshal([]byte(cached), &result)
		return result
	}

	searchResult, err := opensearchClient.Search().
		Index("products").
		Query(elastic.NewMultiMatchQuery(query, "name", "description")).
		Do(ctx)
	if err != nil {
		log.Println("Search error:", err)
		return []string{}
	}

	var results []string
	for _, hit := range searchResult.Hits.Hits {
		results = append(results, string(hit.Source))
	}

	encoded, _ := json.Marshal(results)
	redisClient.Set(ctx, "search:"+query, encoded, 10*time.Minute)

	return results
}

// Call AWS SageMaker for AI recommendations (mocked)
func callSageMaker(userID string, profile map[string]string) []string {
	key := "recommend:" + userID
	if cached, err := redisClient.Get(ctx, key).Result(); err == nil {
		var result []string
		json.Unmarshal([]byte(cached), &result)
		return result
	}

	// Call SageMaker using AWS SDK (mocked with static return)
	recommended := []string{"recommended1", "recommended2", "recommended3"}

	encoded, _ := json.Marshal(recommended)
	redisClient.Set(ctx, key, encoded, 15*time.Minute)
	return recommended
}

// Sync product catalog from Product Catalog Service to OpenSearch (mocked)
func syncProductCatalog() {
	// In real case, consume messages from a queue or event stream
	products := []map[string]interface{}{
		{"id": "1", "name": "Smartphone", "description": "Latest phone"},
		{"id": "2", "name": "Headphones", "description": "Noise-cancelling"},
	}
	for _, product := range products {
		_, err := opensearchClient.Index().
			Index("products").
			Id(product["id"].(string)).
			BodyJson(product).
			Do(ctx)
		if err != nil {
			log.Println("Failed to sync product:", err)
		}
	}
}

// REST handler for search
func searchHandler(c *gin.Context) {
	query := c.Query("q")
	results := searchProducts(query)
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// REST handler for recommendations
func recommendHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}
	profile := getUserProfile(userID)
	recommendations := callSageMaker(userID, profile)
	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

func startRESTServer() {
	r := gin.Default()
	r.GET("/search", searchHandler)
	r.GET("/recommend", recommendHandler)
	log.Fatal(r.Run(":8080"))
}

// gRPC server implementation placeholder
func startGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	// pb.RegisterRecommendationServiceServer(s, &RecommendationServiceServer{})
	fmt.Println("gRPC server listening on port 50051")
	s.Serve(lis)
}

func main() {
	initRedis()
	initOpenSearch()
	syncProductCatalog()
	go startGRPCServer()
	startRESTServer()
}
