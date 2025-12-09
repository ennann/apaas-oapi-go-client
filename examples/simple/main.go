package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ennann/apaas-oapi-go-client/apaas"
)

func main() {
	client, err := apaas.NewClient(apaas.ClientOptions{
		ClientID:     "your_client_id",
		ClientSecret: "your_client_secret",
		Namespace:    "app_xxx",
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	client.SetLoggerLevel(apaas.LoggerLevelInfo)

	ctx := context.Background()
	if err := client.Init(ctx); err != nil {
		log.Fatalf("failed to initialise client: %v", err)
	}

	log.Printf("Access Token: %s", client.Token())
	if ttl, ok := client.TokenExpiresIn(); ok {
		log.Printf("Token expires in: %s", ttl.Truncate(time.Second))
	}
	log.Printf("Namespace: %s", client.Namespace())

	// Example: list objects
	listObjects(ctx, client)

	// Example: search records with pagination
	// searchRecordsWithPagination(ctx, client)

	// Example: batch create with automatic chunking
	// batchCreateExample(ctx, client)

	// Example: batch update with automatic chunking
	// batchUpdateExample(ctx, client)

	// Example: batch delete with automatic chunking
	// batchDeleteExample(ctx, client)
}

func listObjects(ctx context.Context, client *apaas.Client) {
	res, err := client.Object.List(ctx, apaas.ObjectListParams{
		Offset: 0,
		Limit:  20,
	})
	if err != nil {
		log.Printf("failed to list objects: %v", err)
		return
	}
	log.Printf("Object list response code: %s, msg: %s", res.Code, res.Msg)
}

func searchRecordsWithPagination(ctx context.Context, client *apaas.Client) {
	result, err := client.Object.Search.RecordsWithIterator(ctx, apaas.ObjectRecordsIteratorParams{
		ObjectName: "object_store",
		Data: map[string]any{
			"need_total_count": true,
			"page_size":        100,
			"use_page_token":   true,
			"select":           []string{"_id", "store_code", "store_name"},
		},
	})
	if err != nil {
		log.Printf("failed to search records: %v", err)
		return
	}

	fmt.Printf("✅ Total: %d\n", result.Total)
	fmt.Printf("✅ Items retrieved: %d\n", len(result.Items))
	if len(result.Items) > 0 {
		fmt.Printf("✅ First item: %+v\n", result.Items[0])
	}
}

func batchCreateExample(ctx context.Context, client *apaas.Client) {
	records := []map[string]any{
		{"name": "Sample 1", "content": "Content 1"},
		{"name": "Sample 2", "content": "Content 2"},
		// ... 可以添加超过 100 条记录
	}

	result, err := client.Object.Create.RecordsWithIterator(ctx, apaas.ObjectCreateRecordsIteratorParams{
		ObjectName: "object_event_log",
		Records:    records,
		Limit:      100, // 可选，默认 100
	})
	if err != nil {
		log.Printf("failed to create records: %v", err)
		return
	}

	fmt.Printf("✅ Create completed:\n")
	fmt.Printf("   Total: %d\n", result.Total)
	fmt.Printf("   Success: %d\n", result.SuccessCount)
	fmt.Printf("   Failed: %d\n", result.FailedCount)

	// 打印失败的记录
	if result.FailedCount > 0 {
		fmt.Println("❌ Failed records:")
		for _, failed := range result.Failed {
			fmt.Printf("   ID: %s, Error: %s\n", failed.ID, failed.Error)
		}
	}
}

func batchUpdateExample(ctx context.Context, client *apaas.Client) {
	records := []map[string]any{
		{"_id": "id1", "status": "active"},
		{"_id": "id2", "status": "active"},
		// ... 可以添加超过 100 条记录
	}

	result, err := client.Object.Update.RecordsWithIterator(ctx, apaas.ObjectUpdateRecordsIteratorParams{
		ObjectName: "object_store",
		Records:    records,
		Limit:      100, // 可选，默认 100
	})
	if err != nil {
		log.Printf("failed to update records: %v", err)
		return
	}

	fmt.Printf("✅ Update completed:\n")
	fmt.Printf("   Total: %d\n", result.Total)
	fmt.Printf("   Success: %d\n", result.SuccessCount)
	fmt.Printf("   Failed: %d\n", result.FailedCount)

	// 打印失败的记录
	if result.FailedCount > 0 {
		fmt.Println("❌ Failed records:")
		for _, failed := range result.Failed {
			fmt.Printf("   ID: %s, Error: %s\n", failed.ID, failed.Error)
		}
	}
}

func batchDeleteExample(ctx context.Context, client *apaas.Client) {
	ids := []string{"id1", "id2", "id3" /* ... 可以添加超过 100 个 ID */}

	result, err := client.Object.Delete.RecordsWithIterator(ctx, apaas.ObjectDeleteRecordsIteratorParams{
		ObjectName: "object_store",
		IDs:        ids,
		Limit:      100, // 可选，默认 100
	})
	if err != nil {
		log.Printf("failed to delete records: %v", err)
		return
	}

	fmt.Printf("✅ Delete completed:\n")
	fmt.Printf("   Total: %d\n", result.Total)
	fmt.Printf("   Success: %d\n", result.SuccessCount)
	fmt.Printf("   Failed: %d\n", result.FailedCount)

	// 打印失败的记录
	if result.FailedCount > 0 {
		fmt.Println("❌ Failed records:")
		for _, failed := range result.Failed {
			fmt.Printf("   ID: %s, Error: %s\n", failed.ID, failed.Error)
		}
	}
}
