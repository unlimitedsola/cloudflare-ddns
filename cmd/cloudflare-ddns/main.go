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
	hasChanged, oldIP, newIP, err := client.Update(ctx)
	if err != nil {
		log.Fatalf("Failed to update: %s", err)
	}
	if hasChanged {
		log.Printf("updated existing record %s with %s", oldIP, newIP)
	} else {
		log.Print("no changes detected")
	}
}
