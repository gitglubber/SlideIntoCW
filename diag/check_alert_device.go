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

	alerts, _ := slideClient.GetAlerts()
	clients, _ := slideClient.GetClients()

	targetDeviceID := "d_680dlpcxgbe9"

	fmt.Printf("Looking for alerts with device: %s\n\n", targetDeviceID)

	for _, alert := range alerts {
		if alert.DeviceID == targetDeviceID {
			deviceName := alert.GetParsedDeviceName()
			accountID := alert.GetParsedClientID()
			accountName := alert.GetParsedClientName()

			fmt.Printf("Alert ID: %s\n", alert.ID)
			fmt.Printf("  Device ID: %s\n", alert.DeviceID)
			fmt.Printf("  Device Name: %s\n", deviceName)
			fmt.Printf("  Account ID: %s\n", accountID)
			fmt.Printf("  Account Name: %s\n", accountName)
			fmt.Printf("  Alert Type: %s\n", alert.Type)
			fmt.Printf("  Resolved: %t\n\n", alert.Resolved)

			// Try smart matching
			if deviceName != "" {
				deviceUpper := strings.ToUpper(deviceName)
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

				fmt.Printf("  Extracted prefix: '%s'\n", prefix)
				fmt.Printf("  Smart match candidates:\n")

				for _, client := range clients {
					clientUpper := strings.ToUpper(client.Name)

					// Check starts-with
					if strings.HasPrefix(clientUpper, prefix) {
						fmt.Printf("    ✓ '%s' (starts-with)\n", client.Name)
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
						fmt.Printf("    ✓ '%s' (initials: %s)\n", client.Name, initials)
					}
				}
			}

			fmt.Println("\n---")
			break
		}
	}
}