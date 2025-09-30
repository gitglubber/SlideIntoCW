package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"slide-cw-integration/internal/connectwise"
	"slide-cw-integration/internal/database"
	"slide-cw-integration/internal/mapping"
	"slide-cw-integration/internal/slide"
	"slide-cw-integration/pkg/models"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	slideClient    *slide.Client
	cwClient       *connectwise.Client
	mappingService *mapping.Service
	db             *database.DB
	port           string
}

// matchDeviceToClient matches a device name to a client by checking name prefixes and initials
func matchDeviceToClient(deviceName string, slideClients []models.SlideClient) *models.SlideClient {
	if deviceName == "" {
		return nil
	}

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

	// Try to match prefix to client name initials or starts-with
	for i := range slideClients {
		client := &slideClients[i]
		clientUpper := strings.ToUpper(client.Name)

		// Check if client name starts with prefix
		if strings.HasPrefix(clientUpper, prefix) {
			return client
		}

		// Check if prefix matches initials
		words := strings.Fields(clientUpper)
		var initials string
		for _, word := range words {
			if len(word) > 0 && word != "LLC" && word != "INC" && word != "CORP" && word != "P.C." {
				initials += string(word[0])
			}
		}

		if initials == prefix {
			return client
		}
	}

	return nil
}

func NewServer(slideClient *slide.Client, cwClient *connectwise.Client, mappingService *mapping.Service, db *database.DB, port string) *Server {
	if port == "" {
		port = "8080"
	}
	return &Server{
		slideClient:    slideClient,
		cwClient:       cwClient,
		mappingService: mappingService,
		db:             db,
		port:           port,
	}
}

func (s *Server) Start() error {
	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to create static filesystem: %w", err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	// API routes
	http.HandleFunc("/api/health", s.handleHealth)
	http.HandleFunc("/api/dashboard", s.handleDashboard)

	// Slide clients
	http.HandleFunc("/api/slide/clients", s.handleSlideClients)

	// ConnectWise clients
	http.HandleFunc("/api/connectwise/clients", s.handleConnectWiseClients)
	http.HandleFunc("/api/connectwise/boards", s.handleConnectWiseBoards)
	http.HandleFunc("/api/connectwise/statuses", s.handleConnectWiseStatuses)
	http.HandleFunc("/api/connectwise/priorities", s.handleConnectWisePriorities)
	http.HandleFunc("/api/connectwise/types", s.handleConnectWiseTypes)
	http.HandleFunc("/api/connectwise/members", s.handleConnectWiseMembers)

	// Mappings
	http.HandleFunc("/api/mappings", s.handleMappings)
	http.HandleFunc("/api/mappings/create", s.handleCreateMapping)
	http.HandleFunc("/api/mappings/delete", s.handleDeleteMapping)
	http.HandleFunc("/api/mappings/auto", s.handleAutoMap)

	// Ticketing config
	http.HandleFunc("/api/ticketing/config", s.handleTicketingConfig)
	http.HandleFunc("/api/ticketing/config/save", s.handleSaveTicketingConfig)

	// Alerts
	http.HandleFunc("/api/alerts", s.handleAlerts)
	http.HandleFunc("/api/alerts/close", s.handleCloseAlert)

	// Tickets
	http.HandleFunc("/api/tickets/mappings", s.handleTicketMappings)
	http.HandleFunc("/api/admin/reset-mapping", s.handleResetMapping)

	log.Printf("Web UI server starting on http://localhost:%s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

// Health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Dashboard stats
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts, _ := s.slideClient.GetAlerts()
	unresolvedCount := 0
	for _, alert := range alerts {
		if !alert.Resolved {
			unresolvedCount++
		}
	}

	// Get mapping count
	slideClients, _ := s.slideClient.GetClients()
	mappedCount := 0
	for _, client := range slideClients {
		if mapping, err := s.mappingService.GetClientMapping(client.ID); err == nil && mapping != nil {
			mappedCount++
		}
	}

	// Get ticket mapping count
	query := "SELECT COUNT(*) FROM alert_ticket_mappings WHERE closed_at IS NULL"
	var openTickets int
	s.db.GetConn().QueryRow(query).Scan(&openTickets)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"unresolvedAlerts": unresolvedCount,
		"totalAlerts":      len(alerts),
		"mappedClients":    mappedCount,
		"totalClients":     len(slideClients),
		"openTickets":      openTickets,
	})
}

// Slide clients
func (s *Server) handleSlideClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clients, err := s.slideClient.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(clients)
}

// ConnectWise clients
func (s *Server) handleConnectWiseClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clients, err := s.cwClient.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(clients)
}

// ConnectWise boards
func (s *Server) handleConnectWiseBoards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	boards, err := s.cwClient.GetBoards()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(boards)
}

// ConnectWise statuses
func (s *Server) handleConnectWiseStatuses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	boardIDStr := r.URL.Query().Get("boardId")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	statuses, err := s.cwClient.GetStatuses(boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(statuses)
}

// ConnectWise priorities
func (s *Server) handleConnectWisePriorities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	priorities, err := s.cwClient.GetPriorities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(priorities)
}

// ConnectWise types
func (s *Server) handleConnectWiseTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	boardIDStr := r.URL.Query().Get("boardId")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	types, err := s.cwClient.GetTypes(boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(types)
}

// ConnectWise members
func (s *Server) handleConnectWiseMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	members, err := s.cwClient.GetMembers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(members)
}

// Mappings
func (s *Server) handleMappings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	slideClients, err := s.slideClient.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var mappings []map[string]interface{}
	for _, client := range slideClients {
		mapping, err := s.mappingService.GetClientMapping(client.ID)
		result := map[string]interface{}{
			"slideClientId":   client.ID,
			"slideClientName": client.Name,
			"mapped":          false,
		}

		if err == nil && mapping != nil {
			result["mapped"] = true
			result["connectWiseId"] = mapping.ConnectWiseID
			result["connectWiseName"] = mapping.ConnectWiseName
		}

		mappings = append(mappings, result)
	}

	json.NewEncoder(w).Encode(mappings)
}

// Create mapping
func (s *Server) handleCreateMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SlideClientID   string `json:"slideClientId"`
		SlideClientName string `json:"slideClientName"`
		ConnectWiseID   int    `json:"connectWiseId"`
		ConnectWiseName string `json:"connectWiseName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mapping := &models.ClientMapping{
		SlideClientID:   req.SlideClientID,
		SlideClientName: req.SlideClientName,
		ConnectWiseID:   req.ConnectWiseID,
		ConnectWiseName: req.ConnectWiseName,
	}

	if err := s.mappingService.SaveClientMapping(mapping); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Delete mapping
func (s *Server) handleDeleteMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SlideClientID string `json:"slideClientId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "DELETE FROM client_mappings WHERE slide_client_id = ?"
	_, err := s.db.GetConn().Exec(query, req.SlideClientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Auto-map clients
func (s *Server) handleAutoMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slideClients, err := s.slideClient.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cwClients, err := s.cwClient.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.mappingService.MapClients(slideClients, cwClients); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Ticketing config
func (s *Server) handleTicketingConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	config, err := s.db.GetTicketingConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if config == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}

	json.NewEncoder(w).Encode(config)
}

// Save ticketing config
func (s *Server) handleSaveTicketingConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config models.TicketingConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config.UpdatedAt = time.Now()
	if err := s.db.SaveTicketingConfig(&config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Alerts
func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts, err := s.slideClient.GetAlerts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all devices to resolve device_id to client_id
	devices, err := s.slideClient.GetDevices()
	if err != nil {
		log.Printf("Warning: failed to get devices for alert enrichment: %v", err)
		devices = []models.SlideDevice{} // Continue with empty list
	}

	// Get all Slide clients
	slideClients, err := s.slideClient.GetClients()
	if err != nil {
		log.Printf("Warning: failed to get clients for alert enrichment: %v", err)
		slideClients = []models.SlideClient{}
	}

	// Create device ID to client ID map (devices have ClientID which maps to Slide clients)
	deviceToClient := make(map[string]string)
	for _, device := range devices {
		if device.ClientID != "" {
			deviceToClient[device.ID] = device.ClientID
		}
	}

	// Create client ID to name map
	clientIDToName := make(map[string]string)
	for _, client := range slideClients {
		clientIDToName[client.ID] = client.Name
	}

	// Enrich alerts with mapping info
	var enrichedAlerts []map[string]interface{}
	for _, alert := range alerts {
		// IMPORTANT: For MSP accounts, the alert account_id is the MSP, NOT the end client
		// Strategy 1: Try device ID lookup
		// Strategy 2: Try matching device name to client by prefix/initials
		// Strategy 3: Fall back to alert account (probably MSP account)

		var realClientID string
		var realClientName string
		var matchMethod string

		// Strategy 1: Device ID → Client ID lookup
		if alert.DeviceID != "" {
			if clientID, ok := deviceToClient[alert.DeviceID]; ok && clientID != "" {
				realClientID = clientID
				if name, ok := clientIDToName[clientID]; ok {
					realClientName = name
					matchMethod = "device_id"
				}
			}
		}

		// Strategy 2: Match device name to client by prefix/initials
		if realClientID == "" {
			deviceName := alert.GetParsedDeviceName()
			if deviceName != "" {
				matchedClient := matchDeviceToClient(deviceName, slideClients)
				if matchedClient != nil {
					realClientID = matchedClient.ID
					realClientName = matchedClient.Name
					matchMethod = "device_name"
				}
			}
		}

		// Strategy 3: Fall back to alert's account (MSP account - probably wrong)
		if realClientID == "" {
			realClientID = alert.GetParsedClientID()
			realClientName = alert.GetParsedClientName()
			matchMethod = "alert_account"
		}

		// Get the mapped ConnectWise company name
		var cwCompanyName string
		if realClientID != "" {
			mapping, _ := s.mappingService.GetClientMapping(realClientID)
			if mapping != nil {
				cwCompanyName = mapping.ConnectWiseName
			}
		}

		// If no CW mapping found, use the Slide client name
		if cwCompanyName == "" {
			cwCompanyName = realClientName
		}

		enriched := map[string]interface{}{
			"id":           alert.ID,
			"type":         alert.Type,
			"message":      alert.GetParsedMessage(),
			"timestamp":    alert.Timestamp,
			"resolved":     alert.Resolved,
			"deviceId":     alert.DeviceID,
			"deviceName":   alert.GetParsedDeviceName(),
			"clientId":     realClientID,
			"clientName":   cwCompanyName,
			"slideClientName": realClientName,
			"matchMethod":  matchMethod,
		}

		// Check if ticket exists
		ticketMapping, _ := s.mappingService.GetAlertTicketMapping(alert.ID)
		if ticketMapping != nil {
			enriched["ticketId"] = ticketMapping.TicketID
		}

		enrichedAlerts = append(enrichedAlerts, enriched)
	}

	json.NewEncoder(w).Encode(enrichedAlerts)
}

// Close alert
func (s *Server) handleCloseAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AlertID string `json:"alertId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.slideClient.CloseAlert(req.AlertID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleResetMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AlertID string `json:"alertId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Reset the closed_at timestamp to NULL so the monitoring loop will try again
	_, err := s.db.GetConn().Exec("UPDATE alert_ticket_mappings SET closed_at = NULL WHERE alert_id = ?", req.AlertID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Reset mapping for alert: %s", req.AlertID)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Ticket mappings
func (s *Server) handleTicketMappings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT alert_id, ticket_id, created_at, closed_at FROM alert_ticket_mappings ORDER BY created_at DESC LIMIT 100"
	rows, err := s.db.GetConn().Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var mappings []map[string]interface{}
	for rows.Next() {
		var alertID string
		var ticketID int
		var createdAt time.Time
		var closedAt *time.Time

		if err := rows.Scan(&alertID, &ticketID, &createdAt, &closedAt); err != nil {
			continue
		}

		mapping := map[string]interface{}{
			"alertId":   alertID,
			"ticketId":  ticketID,
			"createdAt": createdAt,
		}

		if closedAt != nil {
			mapping["closedAt"] = closedAt
		}

		// Fetch real-time ticket status from ConnectWise
		ticket, err := s.cwClient.GetTicket(ticketID)
		if err != nil {
			log.Printf("Warning: failed to get ticket %d status: %v", ticketID, err)
			mapping["ticketStatus"] = "Unknown"
			mapping["ticketStatusError"] = true
		} else {
			mapping["ticketStatus"] = ticket.Status.Name
			mapping["ticketClosed"] = ticket.IsClosed()
			mapping["ticketClosedFlag"] = ticket.Status.ClosedStatus

			// If ticket is closed in CW but not marked closed in our DB, flag it
			if ticket.IsClosed() && closedAt == nil {
				mapping["needsSync"] = true
			}
		}

		mappings = append(mappings, mapping)
	}

	json.NewEncoder(w).Encode(mappings)
}