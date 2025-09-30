package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"slide-cw-integration/internal/database"
	"slide-cw-integration/internal/mapping"
	"slide-cw-integration/internal/slide"
)

func main() {
	godotenv.Load()

	slideClient := slide.NewClient(
		os.Getenv("SLIDE_API_URL"),
		os.Getenv("SLIDE_API_KEY"),
	)

	clients, _ := slideClient.GetClients()
	devices, _ := slideClient.GetDevices()
	alerts, _ := slideClient.GetAlerts()

	// Check: do the alert fields have device names?
	fmt.Println("=== ALERT DEVICE NAMES ===")
	for i, alert := range alerts {
		if i >= 3 {
			break
		}
		deviceName := alert.GetParsedDeviceName()
		fmt.Printf("Alert %s: Device=%s, DeviceID=%s\n", alert.ID, deviceName, alert.DeviceID)
	}

	// Check: can we match device names to client names?
	fmt.Println("\n=== DEVICES LIST ===")
	for i, device := range devices {
		if i >= 10 {
			break
		}
		fmt.Printf("Device: ID=%s, Name=%s, ClientID=%s\n", device.ID, device.Name, device.ClientID)
	}

	// Strategy: Match device name prefix to client name
	fmt.Println("\n=== DEVICE NAME MATCHING ===")
	for i, alert := range alerts {
		if i >= 5 {
			break
		}
		deviceName := alert.GetParsedDeviceName()
		
		// Try to match device name to a client
		var matchedClient string
		for _, client := range clients {
			clientNameUpper := strings.ToUpper(client.Name)
			deviceNameUpper := strings.ToUpper(deviceName)
			
			// Check if device name contains client name or abbreviation
			words := strings.Fields(clientNameUpper)
			for _, word := range words {
				if len(word) >= 3 && strings.Contains(deviceNameUpper, word) {
					matchedClient = client.Name
					break
				}
			}
			if matchedClient != "" {
				break
			}
		}
		
		fmt.Printf("Device '%s' → Matched Client: %s\n", deviceName, matchedClient)
	}

	// Check mappings
	db, _ := database.Initialize()
	defer db.Close()
	mappingService := mapping.NewService(db)

	fmt.Println("\n=== TESTING MAPPING LOOKUP ===")
	for _, client := range clients {
		mapping, _ := mappingService.GetClientMapping(client.ID)
		if mapping != nil {
			fmt.Printf("Client %s (ID: %s) → CW: %s\n", client.Name, client.ID, mapping.ConnectWiseName)
		}
	}
}
