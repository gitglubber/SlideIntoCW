package models

import (
	"encoding/json"
	"strings"
	"time"
)

// SlideDevice represents a device from the Slide API
type SlideDevice struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
	Type     string `json:"type"`
}

// SlideClient represents a client from the Slide API
type SlideClient struct {
	ID   string `json:"client_id"`
	Name string `json:"name"`
}

// SlideAlert represents an alert from the Slide API
type SlideAlert struct {
	ID          string    `json:"alert_id"`
	DeviceID    string    `json:"device_id"`
	ClientID    string    `json:"client_id"`
	Type        string    `json:"alert_type"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	Resolved    bool      `json:"resolved"`
	AgentID     string    `json:"agent_id"`
	AlertFields string    `json:"alert_fields"`
}

// AlertFieldsData represents the parsed alert_fields JSON
type AlertFieldsData struct {
	Account struct {
		AccountID string `json:"account_id"`
		Name      string `json:"name"`
	} `json:"account"`
	BackupErrorMessage string `json:"backup_error_message"`
	Agent struct {
		Name     string   `json:"name"`
		AgentID  string   `json:"agent_id"`
		Hostname string   `json:"hostname"`
		Addresses []string `json:"addresses"`
	} `json:"agent"`
	Device struct {
		Name     string   `json:"name"`
		Hostname string   `json:"hostname"`
		Addresses []string `json:"addresses"`
	} `json:"device"`
}

// GetParsedClientID extracts the client ID from alert fields
func (a *SlideAlert) GetParsedClientID() string {
	if a.ClientID != "" {
		return a.ClientID
	}

	if a.AlertFields == "" {
		return ""
	}

	var fields AlertFieldsData
	if err := json.Unmarshal([]byte(a.AlertFields), &fields); err != nil {
		return ""
	}

	return fields.Account.AccountID
}

// GetParsedMessage extracts the error message from alert fields
func (a *SlideAlert) GetParsedMessage() string {
	if a.Message != "" {
		return a.Message
	}

	if a.AlertFields == "" {
		return ""
	}

	var fields AlertFieldsData
	if err := json.Unmarshal([]byte(a.AlertFields), &fields); err != nil {
		return ""
	}

	return fields.BackupErrorMessage
}

// GetParsedClientName extracts the client name from alert fields
func (a *SlideAlert) GetParsedClientName() string {
	if a.AlertFields == "" {
		return ""
	}

	var fields AlertFieldsData
	if err := json.Unmarshal([]byte(a.AlertFields), &fields); err != nil {
		return ""
	}

	return fields.Account.Name
}

// GetParsedDeviceName extracts the device name from alert fields
func (a *SlideAlert) GetParsedDeviceName() string {
	if a.AlertFields == "" {
		return ""
	}

	var fields AlertFieldsData
	if err := json.Unmarshal([]byte(a.AlertFields), &fields); err != nil {
		return ""
	}

	return fields.Device.Name
}

// GetParsedAgentName extracts the agent name from alert fields
func (a *SlideAlert) GetParsedAgentName() string {
	if a.AlertFields == "" {
		return ""
	}

	var fields AlertFieldsData
	if err := json.Unmarshal([]byte(a.AlertFields), &fields); err != nil {
		return ""
	}

	return fields.Agent.Name
}

// GetParsedAgentHostname extracts the agent hostname from alert fields
func (a *SlideAlert) GetParsedAgentHostname() string {
	if a.AlertFields == "" {
		return ""
	}

	var fields AlertFieldsData
	if err := json.Unmarshal([]byte(a.AlertFields), &fields); err != nil {
		return ""
	}

	return fields.Agent.Hostname
}

// SlideBackup represents a backup from the Slide API
type SlideBackup struct {
	ID           string    `json:"id"`
	DeviceID     string    `json:"device_id"`
	AgentID      string    `json:"agent_id"`
	ClientID     string    `json:"client_id"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// ConnectWiseClient represents a client in ConnectWise
type ConnectWiseClient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ConnectWiseTicket represents a ticket in ConnectWise
type ConnectWiseTicket struct {
	ID      int    `json:"id"`
	Summary string `json:"summary"`
	Status  struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		ClosedStatus bool `json:"closedStatus"`
	} `json:"status"`
	Company struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"company"`
}

// IsClosed returns true if the ticket is in a closed status
func (t *ConnectWiseTicket) IsClosed() bool {
	// Check the closedStatus flag first
	if t.Status.ClosedStatus {
		return true
	}

	// Also check for common closed status names (case-insensitive)
	// Strip special characters like '>' from status names
	statusName := strings.ToLower(strings.TrimSpace(t.Status.Name))
	statusName = strings.TrimLeft(statusName, ">")
	statusName = strings.TrimSpace(statusName)

	closedStatuses := []string{
		"closed",
		"cancelled",
		"canceled",
		"completed",
		"resolved",
		"done",
		"finished",
		"complete",
	}

	for _, closedStatus := range closedStatuses {
		if statusName == closedStatus {
			return true
		}
	}

	return false
}

// ClientMapping represents the mapping between Slide and ConnectWise clients
type ClientMapping struct {
	ID                int    `json:"id" db:"id"`
	SlideClientID     string `json:"slide_client_id" db:"slide_client_id"`
	SlideClientName   string `json:"slide_client_name" db:"slide_client_name"`
	ConnectWiseID     int    `json:"connectwise_id" db:"connectwise_id"`
	ConnectWiseName   string `json:"connectwise_name" db:"connectwise_name"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// AlertTicketMapping represents the mapping between alerts and tickets
type AlertTicketMapping struct {
	ID        int       `json:"id" db:"id"`
	AlertID   string    `json:"alert_id" db:"alert_id"`
	TicketID  int       `json:"ticket_id" db:"ticket_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty" db:"closed_at"`
}

// ConnectWise configuration models
type ConnectWiseBoard struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Inactive    bool   `json:"inactiveFlag,omitempty"`
}

type ConnectWiseStatus struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	BoardID     int    `json:"boardId,omitempty"`
	SortOrder   int    `json:"sortOrder,omitempty"`
	DisplayOnBoard bool `json:"displayOnBoard,omitempty"`
	Inactive    bool   `json:"inactiveFlag,omitempty"`
	ClosedStatus bool  `json:"closedStatus,omitempty"`
}

type ConnectWisePriority struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	SortOrder   int    `json:"sortOrder,omitempty"`
	Inactive    bool   `json:"inactiveFlag,omitempty"`
}

type ConnectWiseType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	BoardID     int    `json:"boardId,omitempty"`
	Inactive    bool   `json:"inactiveFlag,omitempty"`
}

type ConnectWiseMember struct {
	ID            int    `json:"id"`
	Identifier    string `json:"identifier"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Title         string `json:"title,omitempty"`
	EmailAddress  string `json:"emailAddress,omitempty"`
	Inactive      bool   `json:"inactiveFlag,omitempty"`
}

// TicketingConfig represents the ticketing configuration
type TicketingConfig struct {
	ID              int    `json:"id" db:"id"`
	BoardID         int    `json:"board_id" db:"board_id"`
	BoardName       string `json:"board_name" db:"board_name"`
	StatusID        int    `json:"status_id" db:"status_id"`
	StatusName      string `json:"status_name" db:"status_name"`
	PriorityID      int    `json:"priority_id" db:"priority_id"`
	PriorityName    string `json:"priority_name" db:"priority_name"`
	TypeID          int    `json:"type_id" db:"type_id"`
	TypeName        string `json:"type_name" db:"type_name"`
	TicketSummary   string `json:"ticket_summary" db:"ticket_summary"`
	TicketTemplate  string `json:"ticket_template" db:"ticket_template"`
	AutoAssignTech  bool   `json:"auto_assign_tech" db:"auto_assign_tech"`
	TechnicianID    *int   `json:"technician_id,omitempty" db:"technician_id"`
	TechnicianName  string `json:"technician_name" db:"technician_name"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}