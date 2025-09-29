package alerts

import (
	"fmt"
	"log"
	"strings"
	"time"

	"slide-cw-integration/internal/connectwise"
	"slide-cw-integration/internal/database"
	"slide-cw-integration/internal/mapping"
	"slide-cw-integration/internal/slide"
	"slide-cw-integration/pkg/models"
)

type Monitor struct {
	slideClient     *slide.Client
	connectWise     *connectwise.Client
	mappingService  *mapping.Service
	db              *database.DB
	checkInterval   time.Duration
	stopChan        chan bool
}

func NewMonitor(slideClient *slide.Client, connectWise *connectwise.Client, mappingService *mapping.Service, db *database.DB) *Monitor {
	return &Monitor{
		slideClient:    slideClient,
		connectWise:    connectWise,
		mappingService: mappingService,
		db:             db,
		checkInterval:  5 * time.Minute, // Check every 5 minutes
		stopChan:       make(chan bool),
	}
}

func (m *Monitor) Start() error {
	log.Println("Starting alert monitor...")

	go m.monitorLoop()

	return nil
}

func (m *Monitor) Stop() {
	close(m.stopChan)
}

func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.processAlerts(); err != nil {
				log.Printf("Error processing alerts: %v", err)
			}
		case <-m.stopChan:
			log.Println("Alert monitor stopped")
			return
		}
	}
}

func (m *Monitor) processAlerts() error {
	log.Println("Checking for alerts...")

	alerts, err := m.slideClient.GetAlerts()
	if err != nil {
		return fmt.Errorf("failed to get alerts: %w", err)
	}

	for _, alert := range alerts {
		if alert.Resolved {
			continue // Skip already resolved alerts
		}

		if err := m.handleAlert(&alert); err != nil {
			log.Printf("Error handling alert %s: %v", alert.ID, err)
		}
	}

	// Check for manually closed ConnectWise tickets and close corresponding Slide alerts
	if err := m.processClosedTickets(); err != nil {
		log.Printf("Error processing closed tickets: %v", err)
	}

	return nil
}

func (m *Monitor) handleAlert(alert *models.SlideAlert) error {
	clientID := alert.GetParsedClientID()
	log.Printf("Processing alert: %s for client %s", alert.ID, clientID)

	// Check if alert is resolved by checking backup status
	if m.isAlertResolved(alert) {
		log.Printf("Alert %s is resolved, closing...", alert.ID)
		return m.closeAlert(alert)
	}

	// Check if we already have a ticket for this alert
	// If not, create one
	return m.ensureTicketExists(alert)
}

func (m *Monitor) isAlertResolved(alert *models.SlideAlert) bool {
	// For backup-related alerts, check if a successful backup completed after alert timestamp
	if alert.Type == "backup_failed" || alert.Type == "backup_error" {
		backups, err := m.slideClient.GetBackups()
		if err != nil {
			log.Printf("Error getting backups to check resolution: %v", err)
			return false
		}

		// Find successful backups for this device after the alert timestamp
		for _, backup := range backups {
			if backup.DeviceID == alert.DeviceID &&
				backup.Success &&
				backup.CompletedAt != nil &&
				backup.CompletedAt.After(alert.Timestamp) {
				log.Printf("Found successful backup %s after alert %s", backup.ID, alert.ID)
				return true
			}
		}
	}

	return false
}

func (m *Monitor) closeAlert(alert *models.SlideAlert) error {
	// Close alert in Slide API
	if err := m.slideClient.CloseAlert(alert.ID); err != nil {
		return fmt.Errorf("failed to close alert in Slide: %w", err)
	}

	// Close corresponding ticket in ConnectWise if it exists
	mapping, err := m.mappingService.GetAlertTicketMapping(alert.ID)
	if err == nil && mapping != nil {
		if err := m.connectWise.CloseTicket(mapping.TicketID); err != nil {
			log.Printf("Failed to close ConnectWise ticket %d: %v", mapping.TicketID, err)
		} else {
			log.Printf("Closed ConnectWise ticket %d for alert %s", mapping.TicketID, alert.ID)
		}

		// Mark the mapping as closed in database
		if err := m.mappingService.CloseAlertTicketMapping(alert.ID); err != nil {
			log.Printf("Failed to update alert-ticket mapping in database: %v", err)
		}
	}

	log.Printf("Alert %s closed successfully", alert.ID)
	return nil
}

func (m *Monitor) ensureTicketExists(alert *models.SlideAlert) error {
	// Check if ticket already exists for this alert
	existing, err := m.mappingService.GetAlertTicketMapping(alert.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing ticket mapping: %w", err)
	}

	if existing != nil {
		log.Printf("Ticket %d already exists for alert %s", existing.TicketID, alert.ID)
		return nil
	}

	// Get ConnectWise client ID for this alert's client
	clientID := alert.GetParsedClientID()
	cwClientID, err := m.mappingService.GetConnectWiseClientID(clientID)
	if err != nil {
		return fmt.Errorf("failed to get ConnectWise client ID for alert: %w", err)
	}

	// Get ticketing configuration
	config, err := m.db.GetTicketingConfig()
	if err != nil {
		return fmt.Errorf("failed to get ticketing configuration: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no ticketing configuration found - please run setup first")
	}

	// Get device and agent names from alert fields
	deviceName := alert.GetParsedDeviceName()
	agentName := alert.GetParsedAgentName()
	agentHostname := alert.GetParsedAgentHostname()

	// Get the mapped ConnectWise client name (not the Slide account name)
	var clientName string
	mapping, err := m.mappingService.GetClientMapping(clientID)
	if err != nil || mapping == nil {
		log.Printf("Warning: no client mapping found for %s, using parsed name", clientID)
		clientName = alert.GetParsedClientName()
	} else {
		// Use the ConnectWise client name from the mapping
		clientName = mapping.ConnectWiseName
		log.Printf("Using mapped ConnectWise client name: %s", clientName)
	}

	// Fallback to resolving device name via API if not available
	if deviceName == "" {
		_, resolvedDeviceName, err := m.resolveNames(clientID, alert.DeviceID)
		if err != nil {
			log.Printf("Warning: failed to resolve device name for alert %s: %v", alert.ID, err)
		} else {
			deviceName = resolvedDeviceName
		}
	}

	// Final fallback to IDs
	if clientName == "" {
		clientName = clientID
	}
	if deviceName == "" {
		deviceName = alert.DeviceID
	}
	if agentName == "" {
		agentName = alert.AgentID
	}

	// Apply template substitutions
	summary := m.applyTemplate(config.TicketSummary, alert, clientName, deviceName, agentName, agentHostname)
	description := m.applyTemplate(config.TicketTemplate, alert, clientName, deviceName, agentName, agentHostname)

	// Create ticket in ConnectWise using configuration
	var ticket *models.ConnectWiseTicket
	if config != nil {
		ticket, err = m.connectWise.CreateTicketWithConfig(cwClientID, summary, description, config)
	} else {
		// Fallback to default ticket creation
		ticket, err = m.connectWise.CreateTicket(cwClientID, summary, description)
	}

	if err != nil {
		return fmt.Errorf("failed to create ConnectWise ticket: %w", err)
	}

	// Save alert-ticket mapping in database
	if err := m.mappingService.SaveAlertTicketMapping(alert.ID, ticket.ID); err != nil {
		log.Printf("Failed to save alert-ticket mapping (alert: %s, ticket: %d): %v", alert.ID, ticket.ID, err)
	}

	log.Printf("Created ConnectWise ticket %d for alert %s using configuration", ticket.ID, alert.ID)
	return nil
}

// resolveNames gets the human-readable names for client and device IDs
func (m *Monitor) resolveNames(clientID, deviceID string) (clientName, deviceName string, err error) {
	// Get client name from mapping service
	mapping, err := m.mappingService.GetClientMapping(clientID)
	if err != nil || mapping == nil {
		// Try to get client from Slide API
		clients, err := m.slideClient.GetClients()
		if err != nil {
			return "", "", fmt.Errorf("failed to get clients: %w", err)
		}

		for _, client := range clients {
			if client.ID == clientID {
				clientName = client.Name
				break
			}
		}

		if clientName == "" {
			clientName = clientID // Fallback to ID
		}
	} else {
		clientName = mapping.SlideClientName
	}

	// Get device name from Slide API
	devices, err := m.slideClient.GetDevices()
	if err != nil {
		return clientName, "", fmt.Errorf("failed to get devices: %w", err)
	}

	for _, device := range devices {
		if device.ID == deviceID {
			deviceName = device.Name
			break
		}
	}

	if deviceName == "" {
		deviceName = deviceID // Fallback to ID
	}

	return clientName, deviceName, nil
}

// applyTemplate replaces template variables in the given text
func (m *Monitor) applyTemplate(template string, alert *models.SlideAlert, clientName, deviceName, agentName, agentHostname string) string {
	result := template

	// Get parsed values
	alertMessage := alert.GetParsedMessage()
	if alertMessage == "" {
		alertMessage = alert.Message
	}

	// Replace template variables
	result = strings.ReplaceAll(result, "{{alert_id}}", alert.ID)
	result = strings.ReplaceAll(result, "{{alert_type}}", alert.Type)
	result = strings.ReplaceAll(result, "{{alert_message}}", alertMessage)
	result = strings.ReplaceAll(result, "{{alert_timestamp}}", alert.Timestamp.Format("2006-01-02 15:04:05"))
	result = strings.ReplaceAll(result, "{{client_id}}", alert.GetParsedClientID())
	result = strings.ReplaceAll(result, "{{client_name}}", clientName)
	result = strings.ReplaceAll(result, "{{device_id}}", alert.DeviceID)
	result = strings.ReplaceAll(result, "{{device_name}}", deviceName)
	result = strings.ReplaceAll(result, "{{agent_name}}", agentName)
	result = strings.ReplaceAll(result, "{{agent_hostname}}", agentHostname)

	return result
}

// processClosedTickets checks for ConnectWise tickets that have been manually closed
// and closes the corresponding Slide alerts
func (m *Monitor) processClosedTickets() error {
	// Get all open alert-ticket mappings (where closed_at is NULL)
	query := "SELECT alert_id, ticket_id FROM alert_ticket_mappings WHERE closed_at IS NULL"
	rows, err := m.db.GetConn().Query(query)
	if err != nil {
		return fmt.Errorf("failed to query open alert-ticket mappings: %w", err)
	}
	defer rows.Close()

	var openMappings []struct {
		AlertID  string
		TicketID int
	}

	for rows.Next() {
		var mapping struct {
			AlertID  string
			TicketID int
		}
		if err := rows.Scan(&mapping.AlertID, &mapping.TicketID); err != nil {
			log.Printf("Error scanning mapping row: %v", err)
			continue
		}
		openMappings = append(openMappings, mapping)
	}

	log.Printf("Checking %d open ticket mappings for closure", len(openMappings))

	// Check each ticket to see if it's been closed in ConnectWise
	for _, mapping := range openMappings {
		ticket, err := m.connectWise.GetTicket(mapping.TicketID)
		if err != nil {
			log.Printf("Error getting ticket %d status: %v", mapping.TicketID, err)
			continue
		}

		if ticket.IsClosed() {
			log.Printf("Ticket %d is closed in ConnectWise (status: '%s', closedStatus: %t), closing corresponding Slide alert %s",
				mapping.TicketID, ticket.Status.Name, ticket.Status.ClosedStatus, mapping.AlertID)

			// Close the alert in Slide
			if err := m.slideClient.CloseAlert(mapping.AlertID); err != nil {
				log.Printf("Failed to close Slide alert %s: %v", mapping.AlertID, err)
				continue
			}

			// Mark the mapping as closed in database
			if err := m.mappingService.CloseAlertTicketMapping(mapping.AlertID); err != nil {
				log.Printf("Failed to update alert-ticket mapping for %s: %v", mapping.AlertID, err)
			}

			log.Printf("Successfully closed Slide alert %s (ticket %d was closed in ConnectWise)",
				mapping.AlertID, mapping.TicketID)
		}
	}

	return nil
}