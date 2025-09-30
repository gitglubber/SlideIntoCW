package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"slide-cw-integration/internal/slide"
)

func main() {
	godotenv.Load()

	slideClient := slide.NewClient(
		os.Getenv("SLIDE_API_URL"),
		os.Getenv("SLIDE_API_KEY"),
	)

	clients, err := slideClient.GetClients()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Slide Clients from API:")
	fmt.Println("=======================")
	for _, client := range clients {
		fmt.Printf("ID: %s, Name: %s\n", client.ID, client.Name)
	}
}
