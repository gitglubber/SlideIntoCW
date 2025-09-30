package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"slide-cw-integration/internal/slide"
)

func main() {
	godotenv.Load()

	slideClient := slide.NewClient(
		os.Getenv("SLIDE_API_URL"),
		os.Getenv("SLIDE_API_KEY"),
	)

	clients, _ := slideClient.GetClients()

	fmt.Println("All Slide Clients:")
	fmt.Println("==================")

	for _, client := range clients {
		clientUpper := strings.ToUpper(client.Name)
		words := strings.Fields(clientUpper)
		var initials string
		for _, word := range words {
			if len(word) > 0 && word != "LLC" && word != "INC" && word != "CORP" && word != "P.C." {
				initials += string(word[0])
			}
		}

		fmt.Printf("%-40s  Initials: %-8s  ID: %s\n", client.Name, initials, client.ID)
	}

	fmt.Println("\n\nLooking for 'BM' matches:")
	fmt.Println("========================")
	for _, client := range clients {
		if strings.Contains(strings.ToUpper(client.Name), "BM") ||
			strings.Contains(strings.ToUpper(client.Name), "BARNETT") ||
			strings.Contains(strings.ToUpper(client.Name), "MORO") {
			fmt.Printf("  → %s\n", client.Name)
		}
	}
}