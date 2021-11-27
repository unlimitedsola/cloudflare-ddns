package main

import (
	"cloudflare-ddns/ddns"
	"context"
	"log"
)

func main() {
	client, err := ddns.New()
	if err != nil {
		log.Fatalf("Failed to start: %s", err)
	}
	ctx := context.Background()
	handler := func(result ddns.UpdateResult, err error) {
		if err != nil {
			log.Print(err)
			return
		}
		if result.Updated {
			log.Printf("Updated %s from %s to %s", result.Name, result.Previous, result.Current)
			return
		}
	}
	err = client.Run(ctx, handler)
	if err != nil {
		log.Fatalf("Failed to start: %s", err)
	}
}
