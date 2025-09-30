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

	devices, _ := slideClient.GetDevices()
	clients, _ := slideClient.GetClients()

	targetDeviceID := "d_680dlpcxgbe9"

	// Find the device
	fmt.Printf("Looking for device: %s\n\n", targetDeviceID)

	for _, device := range devices {
		if device.ID == targetDeviceID {
			fmt.Printf("Found Device:\n")
			fmt.Printf("  ID: %s\n", device.ID)
			fmt.Printf("  Name: %s\n", device.Name)
			fmt.Printf("  ClientID: %s\n", device.ClientID)

			// Find the client
			if device.ClientID != "" {
				for _, client := range clients {
					if client.ID == device.ClientID {
						fmt.Printf("\nClient Info:\n")
						fmt.Printf("  ID: %s\n", client.ID)
						fmt.Printf("  Name: %s\n", client.Name)
					}
				}
			} else {
				fmt.Printf("\nNo ClientID in device record\n")

				// Try smart matching
				deviceUpper := strings.ToUpper(device.Name)
				var prefix string
				for i, ch := range deviceUpper {
					if ch == '-' || (ch >= '0' && ch <= '9') {
						prefix = deviceUpper[:i]
						break
					}
				}

				if prefix == "" {
					prefix = deviceUpper
				}

				fmt.Printf("\nExtracted prefix: %s\n", prefix)
				fmt.Printf("\nTrying smart match:\n")

				for _, client := range clients {
					clientUpper := strings.ToUpper(client.Name)

					// Check starts-with
					if strings.HasPrefix(clientUpper, prefix) {
						fmt.Printf("  ✓ Matches '%s' (starts-with)\n", client.Name)
					}

					// Check initials
					words := strings.Fields(clientUpper)
					var initials string
					for _, word := range words {
						if len(word) > 0 && word != "LLC" && word != "INC" && word != "CORP" && word != "P.C." {
							initials += string(word[0])
						}
					}

					if initials == prefix {
						fmt.Printf("  ✓ Matches '%s' (initials: %s)\n", client.Name, initials)
					}
				}
			}

			break
		}
	}
}