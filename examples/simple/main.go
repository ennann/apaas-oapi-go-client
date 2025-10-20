package main

import (
	"context"
	"log"
	"time"

	"github.com/apaas/apaas-sdk/go-client/apaas"
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

	// Example: list objects (requires valid credentials)
	/*
		res, err := client.Object.List(ctx, apaas.ObjectListParams{
			Offset: 0,
			Limit:  20,
		})
		if err != nil {
			log.Fatalf("failed to list objects: %v", err)
		}
		log.Printf("Object list response code: %s", res.Code)
	*/
}
