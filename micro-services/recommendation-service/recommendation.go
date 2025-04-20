package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/olivere/elastic/v7"
)

var (
	ctx              = context.Background()
	redisClient      *redis.Client
	opensearchClient *elastic.Client
)

func initEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Initialize Redis client
func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

// Initialize OpenSearch client
func initOpenSearch() {
	client, err := elastic.NewClient(
		elastic.SetURL(os.Getenv("OPENSEARCH_URL")),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.Fatalf("Error creating OpenSearch client: %v", err)
	}
	opensearchClient = client
}

// Mocked user profile service
func getUserProfile(userID string) map[string]string {
	return map[string]string{
		"age":      "30",
		"gender":   "female",
		"location": "NY",
	}
}

// Search products using OpenSearch
func searchProducts(query string) []string {
	cacheKey := "search:" + query
	if cached, err := redisClient.Get(ctx, cacheKey).Result(); err == nil {
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
	redisClient.Set(ctx, cacheKey, encoded, 10*time.Minute)
	return results
}

// Call AWS SageMaker for recommendations
func callSageMaker(userID string, profile map[string]string) []string {
	cacheKey := "recommend:" + userID
	if cached, err := redisClient.Get(ctx, cacheKey).Result(); err == nil {
		var result []string
		json.Unmarshal([]byte(cached), &result)
		return result
	}

	endpoint := os.Getenv("SAGEMAKER_ENDPOINT_NAME")
	region := os.Getenv("AWS_REGION")

	payload := map[string]interface{}{
		"user_id": userID,
		"profile": profile,
	}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://runtime.sagemaker.%s.amazonaws.com/endpoints/%s/invocations", region, endpoint),
		bytes.NewReader(data))
	if err != nil {
		log.Println("Error creating request:", err)
		return []string{}
	}

	req.Header.Set("Content-Type", "application/json")

	// NOTE: You’ll need to use AWS SDK to sign this request (SigV4).
	// Here it’s mocked, assuming it’s publicly accessible or proxied.

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("SageMaker request failed:", err)
		return []string{}
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var recommendations []string
	if err := json.Unmarshal(body, &recommendations); err != nil {
		log.Println("Failed to unmarshal recommendations:", err)
		return []string{}
	}

	encoded, _ := json.Marshal(recommendations)
	redisClient.Set(ctx, cacheKey, encoded, 15*time.Minute)
	return recommendations
}

// Sync mock product catalog
func syncProductCatalog() {
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

// REST: Search Endpoint
func searchHandler(c *gin.Context) {
	query := c.Query("q")
	results := searchProducts(query)
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// Main function
func main() {
	initEnv()
	initRedis()
	initOpenSearch()
	syncProductCatalog()

	router := gin.Default()
	router.GET("/search", searchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s", port)
	router.Run(":" + port)
}
