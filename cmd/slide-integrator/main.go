package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"slide-cw-integration/internal/alerts"
	"slide-cw-integration/internal/connectwise"
	"slide-cw-integration/internal/database"
	"slide-cw-integration/internal/mapping"
	"slide-cw-integration/internal/slide"
	"slide-cw-integration/internal/web"
)

func main() {
	// Parse command line flags
	mapClients := flag.Bool("map-clients", false, "Auto-map Slide clients to ConnectWise clients")
	mapInteractive := flag.Bool("map-interactive", false, "Interactive TUI for manual client mapping")
	showMappingsFlag := flag.Bool("show-mappings", false, "Show current client mappings")
	clearMappingsFlag := flag.Bool("clear-mappings", false, "Clear all client mappings")
	setupTicketing := flag.Bool("setup-ticketing", false, "Interactive setup for ConnectWise ticketing configuration")
	webUI := flag.Bool("web", false, "Start web UI server (runs alert monitor in background)")
	webPort := flag.String("port", "8080", "Web UI port (default: 8080)")
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help {
		showUsage()
		return
	}

	if *mapClients {
		if err := runMapClients(); err != nil {
			log.Fatal("Failed to map clients:", err)
		}
		return
	}

	if *mapInteractive {
		if err := runInteractiveMapping(); err != nil {
			log.Fatal("Failed to run interactive mapping:", err)
		}
		return
	}

	if *showMappingsFlag {
		if err := showMappings(); err != nil {
			log.Fatal("Failed to show mappings:", err)
		}
		return
	}

	if *clearMappingsFlag {
		if err := clearMappings(); err != nil {
			log.Fatal("Failed to clear mappings:", err)
		}
		return
	}

	if *setupTicketing {
		if err := runTicketingSetup(); err != nil {
			log.Fatal("Failed to run ticketing setup:", err)
		}
		return
	}

	if *webUI {
		if err := runWebUI(*webPort); err != nil {
			log.Fatal("Failed to start web UI:", err)
		}
		return
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	log.Println("Starting Slide-ConnectWise Integration Service...")

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
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

	// Initialize mapping service
	mappingService := mapping.NewService(db)

	// Initialize alert monitor
	alertMonitor := alerts.NewMonitor(slideClient, cwClient, mappingService, db)

	// Start services
	if err := alertMonitor.Start(); err != nil {
		log.Fatal("Failed to start alert monitor:", err)
	}

	log.Println("Service started successfully. Press Ctrl+C to stop.")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, stopping services...")

	alertMonitor.Stop()
	log.Println("Service stopped successfully.")
}

func runWebUI(port string) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	log.Println("Starting Slide-ConnectWise Integration with Web UI...")

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		return err
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

	// Initialize mapping service
	mappingService := mapping.NewService(db)

	// Initialize and start alert monitor in background
	alertMonitor := alerts.NewMonitor(slideClient, cwClient, mappingService, db)
	if err := alertMonitor.Start(); err != nil {
		return err
	}
	log.Println("Alert monitor started in background")

	// Initialize and start web server
	webServer := web.NewServer(slideClient, cwClient, mappingService, db, port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutdown signal received, stopping services...")
		alertMonitor.Stop()
		os.Exit(0)
	}()

	log.Printf("Web UI available at http://localhost:%s", port)
	return webServer.Start()
}