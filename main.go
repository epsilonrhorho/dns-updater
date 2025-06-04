package main

import (
	"context"
	"log"

	"github.com/epsilonrhorho/dns-updater/ipify"
)

func main() {
	client := ipify.NewClient(nil)
	ip, err := client.GetIP(context.Background())
	if err != nil {
		log.Fatalf("failed to get IP: %v", err)
	}
	log.Println("Public IP:", ip)
}
