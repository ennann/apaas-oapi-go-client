package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ennann/apaas-oapi-go-client/apaas"
)

func main() {
	// Create client with custom retry configuration
	client, err := apaas.NewClient(apaas.ClientOptions{
		ClientID:     "your_client_id",
		ClientSecret: "your_client_secret",
		Namespace:    "app_xxx",
		RetryConfig: &apaas.RetryConfig{
			MaxRetries:   3,
			InitialDelay: 500 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
		},
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Set log level
	client.SetLoggerLevel(apaas.LoggerLevelInfo)

	// Initialize client (fetch token)
	ctx := context.Background()
	if err := client.Init(ctx); err != nil {
		log.Fatalf("failed to initialise client: %v", err)
	}

	log.Printf("✓ Client initialized successfully")
	log.Printf("  Access Token: %s", client.Token()[:20]+"...")
	if ttl, ok := client.TokenExpiresIn(); ok {
		log.Printf("  Token expires in: %s", ttl.Truncate(time.Second))
	}
	log.Printf("  Namespace: %s", client.Namespace())

	// Example 1: List objects with error handling
	log.Println("\n--- Example 1: List Objects ---")
	listObjects(client)

	// Example 2: Query records with context timeout
	log.Println("\n--- Example 2: Query Records with Timeout ---")
	queryRecordsWithTimeout(client)

	// Example 3: Create record with retry
	log.Println("\n--- Example 3: Create Record ---")
	createRecord(client)

	// Example 4: Batch operations
	log.Println("\n--- Example 4: Batch Operations ---")
	batchOperations(client)
}

func listObjects(client *apaas.Client) {
	ctx := context.Background()

	res, err := client.Object.List(ctx, apaas.ObjectListParams{
		Offset: 0,
		Limit:  20,
	})

	if err != nil {
		handleError("list objects", err)
		return
	}

	log.Printf("✓ Objects listed successfully: code=%s, msg=%s", res.Code, res.Msg)

	var data struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	if err := res.DecodeData(&data); err != nil {
		log.Printf("✗ Failed to decode response: %v", err)
		return
	}

	log.Printf("  Total objects: %d", data.Total)
	log.Printf("  Fetched: %d objects", len(data.Items))
}

func queryRecordsWithTimeout(client *apaas.Client) {
	// Set a 5-second timeout for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Object.Search.Records(ctx, apaas.ObjectSearchRecordsParams{
		ObjectName: "your_object_name",
		Data: map[string]any{
			"limit":  10,
			"offset": 0,
		},
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("✗ Operation timed out after 5 seconds")
			return
		}
		handleError("query records", err)
		return
	}

	log.Printf("✓ Records queried successfully: code=%s", res.Code)
}

func createRecord(client *apaas.Client) {
	ctx := context.Background()

	record := map[string]any{
		"name":        "Test Record",
		"description": "Created by Go SDK example",
		"status":      "active",
	}

	res, err := client.Object.Create.Record(ctx, apaas.ObjectCreateRecordParams{
		ObjectName: "your_object_name",
		Record:     record,
	})

	if err != nil {
		handleError("create record", err)
		return
	}

	log.Printf("✓ Record created successfully: code=%s", res.Code)

	var data struct {
		ID string `json:"_id"`
	}
	if err := res.DecodeData(&data); err == nil {
		log.Printf("  Record ID: %s", data.ID)
	}
}

func batchOperations(client *apaas.Client) {
	ctx := context.Background()

	// Create multiple records with automatic batching
	records := make([]map[string]any, 0, 150)
	for i := 0; i < 150; i++ {
		records = append(records, map[string]any{
			"name":   "Batch Record " + string(rune(i)),
			"status": "active",
		})
	}

	log.Printf("Creating %d records with automatic batching...", len(records))

	result, err := client.Object.Create.RecordsWithIterator(ctx, apaas.ObjectCreateRecordsIteratorParams{
		ObjectName: "your_object_name",
		Records:    records,
	})

	if err != nil {
		handleError("batch create", err)
		return
	}

	log.Printf("✓ Batch creation completed")
	log.Printf("  Total records: %d", result.Total)
	log.Printf("  Success: %d", result.SuccessCount)
	log.Printf("  Failed: %d", result.FailedCount)

	if result.FailedCount > 0 {
		log.Printf("  Failed records:")
		for _, failed := range result.Failed {
			log.Printf("    ID: %s, Error: %s", failed.ID, failed.Error)
		}
	}
}

func handleError(operation string, err error) {
	log.Printf("✗ Failed to %s: %v", operation, err)

	// Check if it's an API error
	var apiErr *apaas.APIError
	if errors.As(err, &apiErr) {
		log.Printf("  API Error Details:")
		log.Printf("    Status Code: %d", apiErr.StatusCode)
		log.Printf("    Error Code: %s", apiErr.Code)
		log.Printf("    Message: %s", apiErr.Message)
		if apiErr.RequestID != "" {
			log.Printf("    Request ID: %s", apiErr.RequestID)
		}
		log.Printf("    Endpoint: %s %s", apiErr.Method, apiErr.Endpoint)

		if apiErr.IsRetryable() {
			log.Printf("  ℹ This error is retryable (SDK will retry automatically)")
		}
	}

	// Check error code
	if code := apaas.ErrorCode(err); code != "" {
		log.Printf("  Error Code: %s", code)
	}

	// Check status code
	if status := apaas.StatusCode(err); status > 0 {
		log.Printf("  HTTP Status: %d", status)
	}

	// Check if it's a network error
	var netErr *apaas.NetworkError
	if errors.As(err, &netErr) {
		log.Printf("  Network Error: %v", netErr.Err)
	}
}
