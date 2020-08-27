package main

import (
	"cloudflare-ddns/internal/ddns"
	"log"
)

func main() {
	client, err := ddns.New()
	if err != nil {
		log.Fatalf("Failed to start: %s", err)
	}
	hasChanged, oldIP, newIP, err := client.Update()
	if err != nil {
		log.Fatalf("Failed to update: %s", err)
	}
	if hasChanged {
		log.Printf("updated existing record %s with %s", oldIP, newIP)
	} else {
		log.Print("no changes detected")
	}
}
