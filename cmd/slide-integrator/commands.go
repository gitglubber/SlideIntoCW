package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"slide-cw-integration/internal/connectwise"
	"slide-cw-integration/internal/database"
	"slide-cw-integration/internal/mapping"
	"slide-cw-integration/internal/slide"
)

func runMapClients() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Initialize API clients
	slideClient := slide.NewClient(
		os.Getenv("SLIDE_API_URL"),
		os.Getenv("SLIDE_API_KEY"),
	)

	cwClient := connectwise.NewClient(
		os.Getenv("CONNECTWISE_API_URL"),
		os.Getenv("CONNECTWISE_COMPANY_ID"),
		os.Getenv("CONNECTWISE_PUBLIC_KEY"),
		os.Getenv("CONNECTWISE_PRIVATE_KEY"),
		os.Getenv("CONNECTWISE_CLIENT_ID"),
	)

	// Get clients from both APIs
	log.Println("Fetching Slide clients...")
	slideClients, err := slideClient.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get Slide clients: %w", err)
	}
	log.Printf("Found %d Slide clients", len(slideClients))

	log.Println("Fetching ConnectWise clients...")
	cwClients, err := cwClient.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get ConnectWise clients: %w", err)
	}
	log.Printf("Found %d ConnectWise clients", len(cwClients))

	// Map clients
	mappingService := mapping.NewService(db)
	if err := mappingService.MapClients(slideClients, cwClients); err != nil {
		return fmt.Errorf("failed to map clients: %w", err)
	}

	log.Println("Client mapping completed successfully!")
	return nil
}

func runInteractiveMapping() error {
	fmt.Println("⚠️  Interactive TUI mapping has been removed.")
	fmt.Println("Please use the web UI instead:")
	fmt.Println("")
	fmt.Println("  ./slide-integrator.exe -web")
	fmt.Println("")
	fmt.Println("Then open http://localhost:8080 in your browser and go to the 'Client Mappings' tab.")
	return nil
}

func showMappings() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	fmt.Println("Current Client Mappings:")
	fmt.Println("========================")

	// Get all Slide clients to check mappings
	slideClient := slide.NewClient(
		os.Getenv("SLIDE_API_URL"),
		os.Getenv("SLIDE_API_KEY"),
	)

	slideClients, err := slideClient.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get Slide clients: %w", err)
	}

	mappingService := mapping.NewService(db)

	for _, client := range slideClients {
		if mapping, err := mappingService.GetClientMapping(client.ID); err == nil && mapping != nil {
			fmt.Printf("✓ %s → %s (CW ID: %d)\n",
				mapping.SlideClientName, mapping.ConnectWiseName, mapping.ConnectWiseID)
		} else {
			fmt.Printf("✗ %s (not mapped)\n", client.Name)
		}
	}

	return nil
}

func clearMappings() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Clear all mappings using raw SQL
	query := "DELETE FROM client_mappings"
	_, err = db.GetConn().Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear mappings: %w", err)
	}

	fmt.Println("✓ All client mappings cleared successfully!")
	return nil
}

func runTicketingSetup() error {
	fmt.Println("⚠️  Interactive TUI ticketing setup has been removed.")
	fmt.Println("Please use the web UI instead:")
	fmt.Println("")
	fmt.Println("  ./slide-integrator.exe -web")
	fmt.Println("")
	fmt.Println("Then open http://localhost:8080 in your browser and go to the 'Ticketing Config' tab.")
	return nil
}

func showUsage() {
	fmt.Println("Slide-ConnectWise Integration Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  slide-integrator -web               # Start web UI server (RECOMMENDED)")
	fmt.Println("  slide-integrator -web -port 3000    # Start web UI on custom port")
	fmt.Println("")
	fmt.Println("CLI Commands:")
	fmt.Println("  slide-integrator                    # Run alert monitoring service only (no UI)")
	fmt.Println("  slide-integrator -map-clients       # Auto-map Slide clients to ConnectWise")
	fmt.Println("  slide-integrator -show-mappings     # Show current client mappings")
	fmt.Println("  slide-integrator -clear-mappings    # Clear all client mappings")
	fmt.Println("  slide-integrator -h                 # Show this help")
	fmt.Println("")
	fmt.Println("Note: TUI commands (-map-interactive, -setup-ticketing) have been replaced by the web UI.")
}
