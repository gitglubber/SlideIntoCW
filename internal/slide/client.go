package slide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"slide-cw-integration/pkg/models"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type DeviceResponse struct {
	Data []models.SlideDevice `json:"data"`
}

type ClientResponse struct {
	Data []models.SlideClient `json:"data"`
}

type AlertResponse struct {
	Data []models.SlideAlert `json:"data"`
}

type BackupResponse struct {
	Data []models.SlideBackup `json:"data"`
}

func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "https://api.slide.tech"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetDevices() ([]models.SlideDevice, error) {
	var response DeviceResponse
	if err := c.makeRequest("GET", "/v1/device", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	return response.Data, nil
}

func (c *Client) GetClients() ([]models.SlideClient, error) {
	var response ClientResponse
	if err := c.makeRequest("GET", "/v1/client", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}
	return response.Data, nil
}

func (c *Client) GetAlerts() ([]models.SlideAlert, error) {
	var response AlertResponse
	if err := c.makeRequest("GET", "/v1/alert", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	return response.Data, nil
}

func (c *Client) GetBackups() ([]models.SlideBackup, error) {
	var response BackupResponse
	if err := c.makeRequest("GET", "/v1/backup", nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get backups: %w", err)
	}
	return response.Data, nil
}

func (c *Client) CloseAlert(alertID string) error {
	// Try setting both status and resolved fields
	payload := map[string]interface{}{
		"status":   "resolved",
		"resolved": true,
	}

	endpoint := fmt.Sprintf("/v1/alert/%s", alertID)

	log.Printf("Closing alert %s with payload: %+v", alertID, payload)
	err := c.makeRequest("PATCH", endpoint, payload, nil)
	if err != nil {
		log.Printf("Error closing alert %s: %v", alertID, err)
		return err
	}
	log.Printf("Successfully sent close request for alert %s", alertID)
	return nil
}

func (c *Client) GetDevice(deviceID string) (*models.SlideDevice, error) {
	endpoint := fmt.Sprintf("/v1/device/%s", deviceID)
	var device models.SlideDevice
	if err := c.makeRequest("GET", endpoint, nil, &device); err != nil {
		return nil, fmt.Errorf("failed to get device %s: %w", deviceID, err)
	}
	return &device, nil
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

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

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