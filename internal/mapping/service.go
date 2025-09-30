package mapping

import (
	"fmt"
	"log"
	"strings"

	"slide-cw-integration/internal/database"
	"slide-cw-integration/pkg/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) MapClients(slideClients []models.SlideClient, cwClients []models.ConnectWiseClient) error {
	for _, slideClient := range slideClients {
		// Check if mapping already exists
		existing, err := s.db.GetClientMapping(slideClient.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing mapping for %s: %w", slideClient.ID, err)
		}

		if existing != nil {
			log.Printf("Mapping already exists for Slide client %s -> CW client %d",
				slideClient.Name, existing.ConnectWiseID)
			continue
		}

		// Find matching ConnectWise client
		cwClient := s.findMatchingClient(slideClient, cwClients)
		if cwClient == nil {
			log.Printf("No matching ConnectWise client found for Slide client: %s", slideClient.Name)
			continue
		}

		// Save mapping
		mapping := &models.ClientMapping{
			SlideClientID:   slideClient.ID,
			SlideClientName: slideClient.Name,
			ConnectWiseID:   cwClient.ID,
			ConnectWiseName: cwClient.Name,
		}

		if err := s.db.SaveClientMapping(mapping); err != nil {
			return fmt.Errorf("failed to save mapping for %s: %w", slideClient.Name, err)
		}

		log.Printf("Created mapping: %s (Slide) -> %s (ConnectWise)",
			slideClient.Name, cwClient.Name)
	}

	return nil
}

func (s *Service) GetConnectWiseClientID(slideClientID string) (int, error) {
	mapping, err := s.db.GetClientMapping(slideClientID)
	if err != nil {
		return 0, err
	}

	if mapping == nil {
		return 0, fmt.Errorf("no mapping found for Slide client ID: %s", slideClientID)
	}

	return mapping.ConnectWiseID, nil
}

func (s *Service) SaveAlertTicketMapping(alertID string, ticketID int) error {
	mapping := &models.AlertTicketMapping{
		AlertID:  alertID,
		TicketID: ticketID,
	}
	return s.db.SaveAlertTicketMapping(mapping)
}

func (s *Service) GetAlertTicketMapping(alertID string) (*models.AlertTicketMapping, error) {
	return s.db.GetAlertTicketMapping(alertID)
}

func (s *Service) CloseAlertTicketMapping(alertID string) error {
	return s.db.CloseAlertTicketMapping(alertID)
}

func (s *Service) GetClientMapping(slideClientID string) (*models.ClientMapping, error) {
	return s.db.GetClientMapping(slideClientID)
}

func (s *Service) SaveClientMapping(mapping *models.ClientMapping) error {
	return s.db.SaveClientMapping(mapping)
}

func (s *Service) findMatchingClient(slideClient models.SlideClient, cwClients []models.ConnectWiseClient) *models.ConnectWiseClient {
	slideNameLower := strings.ToLower(slideClient.Name)

	// Try exact match first
	for _, cwClient := range cwClients {
		if strings.ToLower(cwClient.Name) == slideNameLower {
			return &cwClient
		}
	}

	// Clean names for better matching (remove LLC, Inc, etc.)
	cleanSlide := cleanCompanyName(slideNameLower)

	// Try exact match on cleaned names
	for _, cwClient := range cwClients {
		cleanCW := cleanCompanyName(strings.ToLower(cwClient.Name))
		if cleanSlide == cleanCW {
			return &cwClient
		}
	}

	// Only try very conservative partial matches to avoid false positives
	// Must be at least 5 characters to avoid matching too broadly
	if len(cleanSlide) >= 5 {
		for _, cwClient := range cwClients {
			cleanCW := cleanCompanyName(strings.ToLower(cwClient.Name))
			if len(cleanCW) >= 5 {
				// Only match if one name is completely contained in the other
				// AND the shorter name is at least 70% of the longer name length
				if strings.Contains(cleanSlide, cleanCW) {
					ratio := float64(len(cleanCW)) / float64(len(cleanSlide))
					if ratio >= 0.7 {
						return &cwClient
					}
				} else if strings.Contains(cleanCW, cleanSlide) {
					ratio := float64(len(cleanSlide)) / float64(len(cleanCW))
					if ratio >= 0.7 {
						return &cwClient
					}
				}
			}
		}
	}

	// Don't match anything else - better to require manual mapping - Due to too many Columbia river.... blah in my dataset.
	return nil
}

func cleanCompanyName(name string) string {
	// Remove common business suffixes for better matching
	suffixes := []string{", llc", " llc", ", inc", " inc", ", corp", " corp",
		", co", " co", ", ltd", " ltd", ", pllc", " pllc", ", company", " company"}

	lower := strings.ToLower(name)
	for _, suffix := range suffixes {
		lower = strings.TrimSuffix(lower, suffix)
	}

	return strings.TrimSpace(lower)
}