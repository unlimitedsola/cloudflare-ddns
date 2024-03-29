package main

import (
	"cloudflare-ddns/ddns"
	"context"
	"log"
)

var handler = &ddns.Handler{
	OnZoneError: func(zone string, err error) {
		log.Printf("Failed to gather zone info for %s: %s", zone, err)
	},

	OnError: func(name string, err error) {
		log.Printf("Failed to update record for %s: %s", name, err)
	},

	OnUpdate: func(name string, recordType string, previous string, current string) {
		log.Printf("Updated %s record %s from %s to %s", recordType, name, previous, current)
	},

	OnCreate: func(name string, recordType string, current string) {
		log.Printf("Created %s record %s pointed to %s", recordType, name, current)
	},
}

func main() {
	client, err := ddns.New()
	if err != nil {
		log.Fatalf("Failed to start: %s", err)
	}
	ctx := context.Background()
	err = client.Run(ctx, handler)
	if err != nil {
		log.Fatalf("Failed to run: %s", err)
	}
}
