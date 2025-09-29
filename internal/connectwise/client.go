package connectwise

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"slide-cw-integration/pkg/models"
)

type Client struct {
	baseURL    string
	companyID  string
	publicKey  string
	privateKey string
	clientID   string
	httpClient *http.Client
}

type CompanyResponse struct {
	Data []models.ConnectWiseClient `json:"data"`
}

type TicketCreateRequest struct {
	Summary     string `json:"summary"`
	Company     CompanyRef `json:"company"`
	Board       BoardRef `json:"board"`
	Status      StatusRef `json:"status,omitempty"`
	Priority    PriorityRef `json:"priority,omitempty"`
	Type        TypeRef `json:"type,omitempty"`
	Description string `json:"initialDescription,omitempty"`
}

type CompanyRef struct {
	ID int `json:"id"`
}

type BoardRef struct {
	Name string `json:"name"`
}

type StatusRef struct {
	Name string `json:"name"`
}

type PriorityRef struct {
	Name string `json:"name"`
}

type TypeRef struct {
	Name string `json:"name"`
}

func NewClient(baseURL, companyID, publicKey, privateKey, clientID string) *Client {
	return &Client{
		baseURL:    baseURL,
		companyID:  companyID,
		publicKey:  publicKey,
		privateKey: privateKey,
		clientID:   clientID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetClients() ([]models.ConnectWiseClient, error) {
	var allCompanies []models.ConnectWiseClient
	page := 1
	pageSize := 1000 // Maximum page size for ConnectWise

	for {
		// Build query parameters for active clients with pagination
		// Use proper ConnectWise API filtering for active companies
		endpoint := fmt.Sprintf("/company/companies?conditions=deletedFlag=false&page=%d&pageSize=%d&orderBy=name", page, pageSize)

		var companies []models.ConnectWiseClient
		if err := c.makeRequest("GET", endpoint, nil, &companies); err != nil {
			return nil, fmt.Errorf("failed to get companies (page %d): %w", page, err)
		}

		log.Printf("ConnectWise API page %d: received %d companies", page, len(companies))

		if len(companies) == 0 {
			break // No more companies to fetch
		}

		allCompanies = append(allCompanies, companies...)

		// If we got less than pageSize, we've reached the end
		if len(companies) < pageSize {
			break
		}

		page++
	}

	log.Printf("ConnectWise API: Total active companies retrieved: %d", len(allCompanies))
	return allCompanies, nil
}

func (c *Client) CreateTicket(companyID int, summary, description string) (*models.ConnectWiseTicket, error) {
	ticket := TicketCreateRequest{
		Summary: summary,
		Company: CompanyRef{ID: companyID},
		Board:   BoardRef{Name: "Service Board"}, // Default board
		Status:  StatusRef{Name: "New"},
		Priority: PriorityRef{Name: "Medium"},
		Type:     TypeRef{Name: "Issue"},
		Description: description,
	}

	var result models.ConnectWiseTicket
	if err := c.makeRequest("POST", "/service/tickets", ticket, &result); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	return &result, nil
}

func (c *Client) CreateTicketWithConfig(companyID int, summary, description string, config *models.TicketingConfig) (*models.ConnectWiseTicket, error) {
	ticket := TicketCreateRequest{
		Summary: summary,
		Company: CompanyRef{ID: companyID},
		Board:   BoardRef{Name: config.BoardName},
		Status:  StatusRef{Name: config.StatusName},
		Priority: PriorityRef{Name: config.PriorityName},
		Type:     TypeRef{Name: config.TypeName},
		Description: description,
	}

	log.Printf("Creating ticket with Company ID: %d, Summary: %s", companyID, summary)
	log.Printf("Board: %s, Status: %s, Priority: %s, Type: %s",
		config.BoardName, config.StatusName, config.PriorityName, config.TypeName)

	var result models.ConnectWiseTicket
	if err := c.makeRequest("POST", "/service/tickets", ticket, &result); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	log.Printf("Created ticket %d for company %d (%s)", result.ID, result.Company.ID, result.Company.Name)
	return &result, nil
}

func (c *Client) UpdateTicket(ticketID int, status string) error {
	update := map[string]interface{}{
		"status": StatusRef{Name: status},
	}

	endpoint := fmt.Sprintf("/service/tickets/%d", ticketID)
	return c.makeRequest("PATCH", endpoint, update, nil)
}

func (c *Client) CloseTicket(ticketID int) error {
	return c.UpdateTicket(ticketID, "Closed")
}

// GetTicket retrieves a specific ticket by ID
func (c *Client) GetTicket(ticketID int) (*models.ConnectWiseTicket, error) {
	endpoint := fmt.Sprintf("/service/tickets/%d", ticketID)

	var result models.ConnectWiseTicket
	if err := c.makeRequest("GET", endpoint, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get ticket %d: %w", ticketID, err)
	}

	return &result, nil
}

// GetBoards fetches all active service boards with pagination
func (c *Client) GetBoards() ([]models.ConnectWiseBoard, error) {
	var allBoards []models.ConnectWiseBoard
	page := 1
	pageSize := 1000

	for {
		// Build URL with pagination and filtering for active boards only
		endpoint := fmt.Sprintf("/service/boards?conditions=inactiveFlag=false&page=%d&pageSize=%d", page, pageSize)

		var boards []models.ConnectWiseBoard
		if err := c.makeRequest("GET", endpoint, nil, &boards); err != nil {
			return nil, fmt.Errorf("failed to get boards (page %d): %w", page, err)
		}

		// Add boards from this page to our collection
		allBoards = append(allBoards, boards...)

		// If we got fewer than pageSize results, we're done
		if len(boards) < pageSize {
			break
		}

		page++
	}

	log.Printf("ConnectWise API: Retrieved %d active boards", len(allBoards))
	return allBoards, nil
}

// GetStatuses fetches all statuses for a specific board with pagination and filters out inactive ones
func (c *Client) GetStatuses(boardID int) ([]models.ConnectWiseStatus, error) {
	var allStatuses []models.ConnectWiseStatus
	page := 1
	pageSize := 1000

	for {
		// Build URL with pagination (conditions parameter seems to cause issues for statuses)
		endpoint := fmt.Sprintf("/service/boards/%d/statuses?page=%d&pageSize=%d", boardID, page, pageSize)

		var statuses []models.ConnectWiseStatus
		if err := c.makeRequest("GET", endpoint, nil, &statuses); err != nil {
			return nil, fmt.Errorf("failed to get statuses for board %d (page %d): %w", boardID, page, err)
		}

		// Filter out inactive statuses client-side
		for _, status := range statuses {
			if !status.Inactive {
				allStatuses = append(allStatuses, status)
			}
		}

		// If we got fewer than pageSize results, we're done
		if len(statuses) < pageSize {
			break
		}

		page++
	}

	log.Printf("ConnectWise API: Retrieved %d active statuses for board %d (filtered from server results)", len(allStatuses), boardID)
	return allStatuses, nil
}

// GetPriorities fetches all ticket priorities with pagination and filters out inactive ones
func (c *Client) GetPriorities() ([]models.ConnectWisePriority, error) {
	var allPriorities []models.ConnectWisePriority
	page := 1
	pageSize := 1000

	for {
		// Build URL with pagination (conditions parameter seems to cause issues for priorities)
		endpoint := fmt.Sprintf("/service/priorities?page=%d&pageSize=%d", page, pageSize)

		var priorities []models.ConnectWisePriority
		if err := c.makeRequest("GET", endpoint, nil, &priorities); err != nil {
			return nil, fmt.Errorf("failed to get priorities (page %d): %w", page, err)
		}

		// Filter out inactive priorities client-side
		for _, priority := range priorities {
			if !priority.Inactive {
				allPriorities = append(allPriorities, priority)
			}
		}

		// If we got fewer than pageSize results, we're done
		if len(priorities) < pageSize {
			break
		}

		page++
	}

	log.Printf("ConnectWise API: Retrieved %d active priorities (filtered from server results)", len(allPriorities))
	return allPriorities, nil
}

// GetTypes fetches all active ticket types for a specific board with pagination
func (c *Client) GetTypes(boardID int) ([]models.ConnectWiseType, error) {
	var allTypes []models.ConnectWiseType
	page := 1
	pageSize := 1000

	for {
		// Build URL with pagination and filtering for active types only
		endpoint := fmt.Sprintf("/service/boards/%d/types?conditions=inactiveFlag=false&page=%d&pageSize=%d", boardID, page, pageSize)

		var types []models.ConnectWiseType
		if err := c.makeRequest("GET", endpoint, nil, &types); err != nil {
			return nil, fmt.Errorf("failed to get types for board %d (page %d): %w", boardID, page, err)
		}

		// Add types from this page to our collection
		allTypes = append(allTypes, types...)

		// If we got fewer than pageSize results, we're done
		if len(types) < pageSize {
			break
		}

		page++
	}

	log.Printf("ConnectWise API: Retrieved %d active types for board %d", len(allTypes), boardID)
	return allTypes, nil
}

// GetMembers fetches all active members/technicians with pagination
func (c *Client) GetMembers() ([]models.ConnectWiseMember, error) {
	var allMembers []models.ConnectWiseMember
	page := 1
	pageSize := 1000

	for {
		// Build URL with pagination and filtering for active members only
		endpoint := fmt.Sprintf("/system/members?conditions=inactiveFlag=false&page=%d&pageSize=%d", page, pageSize)

		var members []models.ConnectWiseMember
		if err := c.makeRequest("GET", endpoint, nil, &members); err != nil {
			return nil, fmt.Errorf("failed to get members (page %d): %w", page, err)
		}

		// Add members from this page to our collection
		allMembers = append(allMembers, members...)

		// If we got fewer than pageSize results, we're done
		if len(members) < pageSize {
			break
		}

		page++
	}

	log.Printf("ConnectWise API: Retrieved %d active members", len(allMembers))
	return allMembers, nil
}

func (c *Client) makeRequest(method, endpoint string, payload interface{}, result interface{}) error {
	var body *bytes.Buffer
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	} else {
		body = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// ConnectWise Basic Auth: "companyID+publicKey:privateKey"
	auth := fmt.Sprintf("%s+%s:%s", c.companyID, c.publicKey, c.privateKey)
	authHeader := base64.StdEncoding.EncodeToString([]byte(auth))

	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("clientId", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}