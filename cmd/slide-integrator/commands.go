package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"slide-cw-integration/internal/connectwise"
	"slide-cw-integration/internal/database"
	"slide-cw-integration/internal/mapping"
	"slide-cw-integration/internal/slide"
	"slide-cw-integration/internal/tui"
	"slide-cw-integration/pkg/models"
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
	log.Println("Fetching clients from both APIs...")

	slideClients, err := slideClient.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get Slide clients: %w", err)
	}

	cwClients, err := cwClient.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get ConnectWise clients: %w", err)
	}

	// Get existing mappings
	mappingService := mapping.NewService(db)
	existingMappings := make(map[string]*models.ClientMapping)
	for _, slideClient := range slideClients {
		if mapping, err := mappingService.GetClientMapping(slideClient.ID); err == nil && mapping != nil {
			existingMappings[slideClient.ID] = mapping
		}
	}

	// Create TUI model
	model := tui.NewMappingModel(
		slideClients,
		cwClients,
		existingMappings,
		func(slideClientID string, cwClientID int) error {
			// Find the client names for the mapping
			var slideClientName, cwClientName string
			for _, sc := range slideClients {
				if sc.ID == slideClientID {
					slideClientName = sc.Name
					break
				}
			}
			for _, cc := range cwClients {
				if cc.ID == cwClientID {
					cwClientName = cc.Name
					break
				}
			}

			mapping := &models.ClientMapping{
				SlideClientID:   slideClientID,
				SlideClientName: slideClientName,
				ConnectWiseID:   cwClientID,
				ConnectWiseName: cwClientName,
			}
			return mappingService.SaveClientMapping(mapping)
		},
	)

	// Run the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
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

	// Initialize ConnectWise client
	cwClient := connectwise.NewClient(
		os.Getenv("CONNECTWISE_API_URL"),
		os.Getenv("CONNECTWISE_COMPANY_ID"),
		os.Getenv("CONNECTWISE_PUBLIC_KEY"),
		os.Getenv("CONNECTWISE_PRIVATE_KEY"),
		os.Getenv("CONNECTWISE_CLIENT_ID"),
	)

	log.Println("Fetching ConnectWise configuration data...")

	// Fetch boards, priorities, and members
	boards, err := cwClient.GetBoards()
	if err != nil {
		return fmt.Errorf("failed to get boards: %w", err)
	}

	priorities, err := cwClient.GetPriorities()
	if err != nil {
		return fmt.Errorf("failed to get priorities: %w", err)
	}

	members, err := cwClient.GetMembers()
	if err != nil {
		return fmt.Errorf("failed to get members: %w", err)
	}

	log.Printf("Found %d boards, %d priorities, %d members", len(boards), len(priorities), len(members))

	// Create TUI model
	model := tui.NewTicketingModel(
		boards,
		priorities,
		members,
		func(config *models.TicketingConfig) error {
			return db.SaveTicketingConfig(config)
		},
		func(boardID int) ([]models.ConnectWiseStatus, []models.ConnectWiseType, error) {
			// Load statuses and types for the selected board
			statuses, err := cwClient.GetStatuses(boardID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get statuses for board %d: %w", boardID, err)
			}

			types, err := cwClient.GetTypes(boardID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get types for board %d: %w", boardID, err)
			}

			log.Printf("Loaded %d statuses and %d types for board %d", len(statuses), len(types), boardID)
			return statuses, types, nil
		},
	)

	// Run the TUI
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func showUsage() {
	fmt.Println("Slide-ConnectWise Integration Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  slide-integrator                    # Run the alert monitoring service")
	fmt.Println("  slide-integrator -web               # Start web UI server (recommended)")
	fmt.Println("  slide-integrator -web -port 3000    # Start web UI on custom port")
	fmt.Println("  slide-integrator -map-clients       # Auto-map Slide clients to ConnectWise clients")
	fmt.Println("  slide-integrator -map-interactive   # Interactive TUI for manual client mapping")
	fmt.Println("  slide-integrator -show-mappings     # Show current client mappings")
	fmt.Println("  slide-integrator -clear-mappings    # Clear all client mappings")
	fmt.Println("  slide-integrator -setup-ticketing   # Interactive setup for ConnectWise ticketing configuration")
	fmt.Println("  slide-integrator -h                 # Show this help")
}