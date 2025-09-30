package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"slide-cw-integration/internal/slide"
)

func matchDeviceToClient(deviceName string, clients []string) string {
	deviceUpper := strings.ToUpper(deviceName)
	
	// Extract prefix (letters before hyphen or numbers)
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
	
	fmt.Printf("Device '%s' → Prefix: '%s'\n", deviceName, prefix)
	
	// Try to match prefix to client name initials
	for _, client := range clients {
		clientUpper := strings.ToUpper(client)
		
		// Check if prefix matches initials
		words := strings.Fields(clientUpper)
		var initials string
		for _, word := range words {
			if len(word) > 0 && word != "LLC" && word != "INC" && word != "CORP" {
				initials += string(word[0])
			}
		}
		
		if initials == prefix {
			return client
		}
		
		// Also check if client name starts with prefix
		if strings.HasPrefix(clientUpper, prefix) {
			return client
		}
	}
	
	return ""
}

func main() {
	godotenv.Load()

	slideClient := slide.NewClient(
		os.Getenv("SLIDE_API_URL"),
		os.Getenv("SLIDE_API_KEY"),
	)

	clients, _ := slideClient.GetClients()
	alerts, _ := slideClient.GetAlerts()

	clientNames := make([]string, len(clients))
	for i, c := range clients {
		clientNames[i] = c.Name
	}

	fmt.Println("=== SMART DEVICE MATCHING ===\n")
	seen := make(map[string]bool)
	for _, alert := range alerts {
		deviceName := alert.GetParsedDeviceName()
		if seen[deviceName] || deviceName == "" {
			continue
		}
		seen[deviceName] = true
		
		matched := matchDeviceToClient(deviceName, clientNames)
		if matched != "" {
			fmt.Printf("  ✓ Matched to: %s\n\n", matched)
		} else {
			fmt.Printf("  ✗ No match found\n\n")
		}
	}
}
