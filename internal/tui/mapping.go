package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"slide-cw-integration/pkg/models"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Padding(1, 0)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981")).
		Padding(0, 1)

	mappedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Padding(0, 1)

	unmappedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#374151")).
		Padding(0, 1)

	searchInputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1F2937")).
		Padding(0, 1)

	suggestionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Padding(0, 1)

	highlightStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FBBF24")).
		Bold(true)
)

type state int

const (
	selectingSlideClient state = iota
	searchingSlideClient
	selectingCWClient
	searchingCWClient
	confirmMapping
	viewMappings
)

type MappingModel struct {
	slideClients        []models.SlideClient
	cwClients           []models.ConnectWiseClient
	existingMappings    map[string]*models.ClientMapping
	selectedSlide       int
	selectedSlideClient *models.SlideClient // Store the actual selected client
	selectedCW          int
	currentState        state
	slideClientCursor   int
	cwClientCursor      int
	slideSearchQuery    string
	cwSearchQuery       string
	filteredSlideClients []models.SlideClient
	filteredCWClients   []models.ConnectWiseClient
	pendingMapping      *pendingMapping
	message             string
	onSave              func(slideClientID string, cwClientID int) error
}

type pendingMapping struct {
	slideClient models.SlideClient
	cwClient    models.ConnectWiseClient
}

func NewMappingModel(
	slideClients []models.SlideClient,
	cwClients []models.ConnectWiseClient,
	existingMappings map[string]*models.ClientMapping,
	onSave func(slideClientID string, cwClientID int) error,
) MappingModel {
	m := MappingModel{
		slideClients:        slideClients,
		cwClients:           cwClients,
		existingMappings:    existingMappings,
		currentState:        selectingSlideClient,
		onSave:              onSave,
		filteredSlideClients: slideClients,
		filteredCWClients:   cwClients,
	}
	return m
}

func (m MappingModel) Init() tea.Cmd {
	return nil
}

func (m MappingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			switch m.currentState {
			case selectingSlideClient, searchingSlideClient:
				m.currentState = viewMappings
			case viewMappings:
				m.currentState = selectingSlideClient
				// Reset ConnectWise state when returning to slide selection
				m.cwClientCursor = 0
				m.cwSearchQuery = ""
				m.filteredCWClients = m.cwClients
			}

		case "/", "s":
			// Start search mode (/ or s key)
			switch m.currentState {
			case selectingSlideClient:
				m.currentState = searchingSlideClient
				m.slideSearchQuery = ""
				m.slideClientCursor = 0
			case selectingCWClient:
				m.currentState = searchingCWClient
				m.cwSearchQuery = ""
				m.cwClientCursor = 0
			}

		case "enter":
			switch m.currentState {
			case selectingSlideClient:
				if len(m.filteredSlideClients) > 0 && m.slideClientCursor < len(m.filteredSlideClients) {
					// Use the filtered client directly
					selectedClient := m.filteredSlideClients[m.slideClientCursor]
					slideIndex := m.findSlideClientIndex(selectedClient.ID)
					if slideIndex == -1 {
						m.message = fmt.Sprintf("Error: Could not find slide client %s", selectedClient.Name)
						return m, nil
					}
					m.selectedSlide = slideIndex
					m.selectedSlideClient = &selectedClient // Store the actual selected client
					m.currentState = selectingCWClient
					m.cwClientCursor = 0
					m.cwSearchQuery = "" // Clear any previous search
					// Pre-filter CW clients based on selected Slide client
					m.filterCWClients("")
				}

			case selectingCWClient:
				if len(m.filteredCWClients) > 0 && m.cwClientCursor < len(m.filteredCWClients) {
					// Use the filtered client directly
					selectedCWClient := m.filteredCWClients[m.cwClientCursor]
					m.selectedCW = m.findCWClientIndex(selectedCWClient.ID)
					// Use the directly stored selected client for accurate mapping
					var slideClient models.SlideClient
					if m.selectedSlideClient != nil {
						slideClient = *m.selectedSlideClient
					} else {
						slideClient = m.slideClients[m.selectedSlide]
					}
					m.pendingMapping = &pendingMapping{
						slideClient: slideClient,
						cwClient:    m.cwClients[m.selectedCW],
					}
					m.currentState = confirmMapping
				}

			case searchingSlideClient:
				if len(m.filteredSlideClients) > 0 && m.slideClientCursor < len(m.filteredSlideClients) {
					selectedClient := m.filteredSlideClients[m.slideClientCursor]
					slideIndex := m.findSlideClientIndex(selectedClient.ID)
					if slideIndex == -1 {
						m.message = fmt.Sprintf("Error: Could not find slide client %s", selectedClient.Name)
						return m, nil
					}
					m.selectedSlide = slideIndex
					m.selectedSlideClient = &selectedClient // Store the actual selected client
					m.currentState = selectingCWClient
					m.cwClientCursor = 0
					m.cwSearchQuery = "" // Clear any previous search
					m.filterCWClients("")
				}

			case searchingCWClient:
				if len(m.filteredCWClients) > 0 && m.cwClientCursor < len(m.filteredCWClients) {
					selectedCWClient := m.filteredCWClients[m.cwClientCursor]
					m.selectedCW = m.findCWClientIndex(selectedCWClient.ID)
					// Use the directly stored selected client for accurate mapping
					var slideClient models.SlideClient
					if m.selectedSlideClient != nil {
						slideClient = *m.selectedSlideClient
					} else {
						slideClient = m.slideClients[m.selectedSlide]
					}
					m.pendingMapping = &pendingMapping{
						slideClient: slideClient,
						cwClient:    m.cwClients[m.selectedCW],
					}
					m.currentState = confirmMapping
				}

			case confirmMapping:
				// Save the mapping
				err := m.onSave(m.pendingMapping.slideClient.ID, m.pendingMapping.cwClient.ID)
				if err != nil {
					m.message = fmt.Sprintf("Error: %v", err)
				} else {
					// Update existing mappings in memory
					newMapping := &models.ClientMapping{
						SlideClientID:   m.pendingMapping.slideClient.ID,
						SlideClientName: m.pendingMapping.slideClient.Name,
						ConnectWiseID:   m.pendingMapping.cwClient.ID,
						ConnectWiseName: m.pendingMapping.cwClient.Name,
					}
					m.existingMappings[m.pendingMapping.slideClient.ID] = newMapping
					m.message = fmt.Sprintf("✓ Mapped %s → %s (ID:%d)",
						m.pendingMapping.slideClient.Name,
						m.pendingMapping.cwClient.Name,
						m.pendingMapping.cwClient.ID)
				}
				m.currentState = selectingSlideClient
				m.pendingMapping = nil
				m.filterSlideClients("")
				// Reset ConnectWise state as well
				m.cwClientCursor = 0
				m.cwSearchQuery = ""
				m.filteredCWClients = m.cwClients
			}

		case "esc":
			switch m.currentState {
			case searchingSlideClient:
				m.currentState = selectingSlideClient
				m.slideSearchQuery = ""
				m.filterSlideClients("")
			case searchingCWClient:
				m.currentState = selectingCWClient
				m.cwSearchQuery = ""
				m.filterCWClients("")
			case selectingCWClient:
				m.currentState = selectingSlideClient
				// Reset ConnectWise state when going back to slide selection
				m.cwClientCursor = 0
				m.cwSearchQuery = ""
				m.filteredCWClients = m.cwClients
			case confirmMapping:
				m.currentState = selectingCWClient
				m.pendingMapping = nil
			}

		case "backspace":
			switch m.currentState {
			case searchingSlideClient:
				if len(m.slideSearchQuery) > 0 {
					m.slideSearchQuery = m.slideSearchQuery[:len(m.slideSearchQuery)-1]
					m.filterSlideClients(m.slideSearchQuery)
					m.slideClientCursor = 0
				}
			case searchingCWClient:
				if len(m.cwSearchQuery) > 0 {
					m.cwSearchQuery = m.cwSearchQuery[:len(m.cwSearchQuery)-1]
					m.filterCWClients(m.cwSearchQuery)
					m.cwClientCursor = 0
				}
			}

		case "up", "k":
			switch m.currentState {
			case selectingSlideClient, searchingSlideClient:
				if m.slideClientCursor > 0 {
					m.slideClientCursor--
				}
			case selectingCWClient, searchingCWClient:
				if m.cwClientCursor > 0 {
					m.cwClientCursor--
				}
			}

		case "down", "j":
			switch m.currentState {
			case selectingSlideClient, searchingSlideClient:
				if m.slideClientCursor < len(m.filteredSlideClients)-1 {
					m.slideClientCursor++
				}
			case selectingCWClient, searchingCWClient:
				if m.cwClientCursor < len(m.filteredCWClients)-1 {
					m.cwClientCursor++
				}
			}

		default:
			// Handle text input in search mode
			switch m.currentState {
			case searchingSlideClient:
				// Accept alphanumeric characters and common punctuation for search
				key := msg.String()
				if len(key) == 1 {
					// Allow letters, numbers, spaces, and basic punctuation
					char := key[0]
					if (char >= 'a' && char <= 'z') ||
					   (char >= 'A' && char <= 'Z') ||
					   (char >= '0' && char <= '9') ||
					   char == ' ' || char == '.' || char == '-' || char == '_' || char == '&' {
						m.slideSearchQuery += key
						m.filterSlideClients(m.slideSearchQuery)
						m.slideClientCursor = 0
					}
				}
			case searchingCWClient:
				// Accept alphanumeric characters and common punctuation for search
				key := msg.String()
				if len(key) == 1 {
					// Allow letters, numbers, spaces, and basic punctuation
					char := key[0]
					if (char >= 'a' && char <= 'z') ||
					   (char >= 'A' && char <= 'Z') ||
					   (char >= '0' && char <= '9') ||
					   char == ' ' || char == '.' || char == '-' || char == '_' || char == '&' {
						m.cwSearchQuery += key
						m.filterCWClients(m.cwSearchQuery)
						m.cwClientCursor = 0
					}
				}
			}
		}
	}

	return m, nil
}

func (m MappingModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📋 Manual Client Mapping"))
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("✓ " + m.message))
		b.WriteString("\n\n")
	}

	switch m.currentState {
	case selectingSlideClient:
		b.WriteString(m.renderSlideClientSelection())
	case selectingCWClient:
		b.WriteString(m.renderCWClientSelection())
	case confirmMapping:
		b.WriteString(m.renderConfirmMapping())
	case viewMappings:
		b.WriteString(m.renderExistingMappings())
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m MappingModel) renderSlideClientSelection() string {
	var b strings.Builder

	if m.currentState == searchingSlideClient {
		b.WriteString(headerStyle.Render("🔍 Search Slide Clients:"))
		b.WriteString("\n")
		searchBox := fmt.Sprintf("Search: %s█", m.slideSearchQuery)
		b.WriteString(searchInputStyle.Render(searchBox))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Type to search, ESC to cancel"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(headerStyle.Render("Select Slide Client to Map:"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Press 's' to search"))
		b.WriteString("\n\n")
	}

	clients := m.filteredSlideClients
	if len(clients) == 0 {
		b.WriteString(unmappedStyle.Render("No clients found"))
		return b.String()
	}

	// Show up to 15 clients to prevent overwhelming display
	maxDisplay := 15
	startIndex := 0
	endIndex := len(clients)

	if len(clients) > maxDisplay {
		// Calculate window around cursor
		half := maxDisplay / 2
		startIndex = m.slideClientCursor - half
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex = startIndex + maxDisplay
		if endIndex > len(clients) {
			endIndex = len(clients)
			startIndex = endIndex - maxDisplay
			if startIndex < 0 {
				startIndex = 0
			}
		}
		clients = clients[startIndex:endIndex]
	}

	for i, client := range clients {
		var style lipgloss.Style
		prefix := "  "

		// The actual index in the filtered list
		actualIndex := startIndex + i

		if actualIndex == m.slideClientCursor {
			style = selectedStyle
			prefix = "► "
		} else {
			style = normalStyle
		}

		// Check if already mapped
		status := ""
		if mapping, exists := m.existingMappings[client.ID]; exists {
			status = fmt.Sprintf(" → %s", mapping.ConnectWiseName)
			style = style.Foreground(lipgloss.Color("#10B981"))
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s%s", prefix, client.Name, status)))
		b.WriteString("\n")
	}

	// Show count info
	if len(m.filteredSlideClients) > maxDisplay {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			fmt.Sprintf("Showing %d of %d clients", len(clients), len(m.filteredSlideClients))))
	}

	return b.String()
}

func (m MappingModel) renderCWClientSelection() string {
	var b strings.Builder

	// Use the directly stored selected client for more reliability
	var selectedSlide models.SlideClient
	if m.selectedSlideClient != nil {
		selectedSlide = *m.selectedSlideClient
	} else {
		// Fallback to index-based selection if needed
		if m.selectedSlide < 0 || m.selectedSlide >= len(m.slideClients) {
			b.WriteString(headerStyle.Render("Select ConnectWise Client for: [Invalid Selection]"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("Error: Invalid slide client selection"))
			b.WriteString("\n\n")
			return b.String()
		}
		selectedSlide = m.slideClients[m.selectedSlide]
	}

	if m.currentState == searchingCWClient {
		b.WriteString(headerStyle.Render(fmt.Sprintf("🔍 Search ConnectWise Clients for: %s", selectedSlide.Name)))
		b.WriteString("\n")
		searchBox := fmt.Sprintf("Search: %s█", m.cwSearchQuery)
		b.WriteString(searchInputStyle.Render(searchBox))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Type to search, ESC to cancel"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(headerStyle.Render(fmt.Sprintf("Select ConnectWise Client for: %s", selectedSlide.Name)))
		b.WriteString("\n")
		b.WriteString(suggestionStyle.Render("💡 Similar clients shown first"))
		b.WriteString(" • ")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Press 's' to search"))
		b.WriteString("\n\n")
	}

	clients := m.filteredCWClients
	if len(clients) == 0 {
		b.WriteString(unmappedStyle.Render("No clients found"))
		return b.String()
	}

	// Show up to 20 clients for ConnectWise (usually more clients)
	maxDisplay := 20
	startIndex := 0
	endIndex := len(clients)

	if len(clients) > maxDisplay {
		// Calculate window around cursor
		half := maxDisplay / 2
		startIndex = m.cwClientCursor - half
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex = startIndex + maxDisplay
		if endIndex > len(clients) {
			endIndex = len(clients)
			startIndex = endIndex - maxDisplay
			if startIndex < 0 {
				startIndex = 0
			}
		}
		clients = clients[startIndex:endIndex]
	}

	// Get similar clients for highlighting (only if not searching)
	var similarClients map[int]bool
	if m.currentState != searchingCWClient {
		similarClients = make(map[int]bool)
		similar := m.findSimilarCWClients(selectedSlide.Name)
		for _, sc := range similar {
			similarClients[sc.ID] = true
		}
	}

	for i, client := range clients {
		var style lipgloss.Style
		prefix := "  "

		// The actual index in the filtered list
		actualIndex := startIndex + i

		if actualIndex == m.cwClientCursor {
			style = selectedStyle
			prefix = "► "
		} else {
			style = normalStyle
		}

		// Highlight suggested/similar clients
		if similarClients != nil && similarClients[client.ID] {
			style = style.Foreground(lipgloss.Color("#3B82F6"))
			if actualIndex != m.cwClientCursor {
				prefix = "💡 "
			}
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s (ID: %d)", prefix, client.Name, client.ID)))
		b.WriteString("\n")
	}

	// Show count and pagination info
	if len(m.filteredCWClients) > maxDisplay {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			fmt.Sprintf("Showing %d of %d clients (use ↑/↓ to scroll)", len(clients), len(m.filteredCWClients))))
	}

	return b.String()
}

func (m MappingModel) renderConfirmMapping() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Confirm Mapping:"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Slide Client: %s\n", m.pendingMapping.slideClient.Name))
	b.WriteString(fmt.Sprintf("  ConnectWise Client: %s (ID: %d)\n\n", m.pendingMapping.cwClient.Name, m.pendingMapping.cwClient.ID))

	b.WriteString(selectedStyle.Render("Press ENTER to confirm, ESC to cancel"))

	return b.String()
}

func (m MappingModel) renderExistingMappings() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Existing Mappings:"))
	b.WriteString("\n\n")

	if len(m.existingMappings) == 0 {
		b.WriteString(unmappedStyle.Render("No mappings exist yet"))
		return b.String()
	}

	for _, mapping := range m.existingMappings {
		b.WriteString(mappedStyle.Render(fmt.Sprintf("✓ %s → %s", mapping.SlideClientName, mapping.ConnectWiseName)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m MappingModel) renderHelp() string {
	switch m.currentState {
	case selectingSlideClient:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"↑/↓: Navigate • ENTER: Select • s: Search • TAB: View mappings • q: Quit")
	case searchingSlideClient:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"Type to search • ↑/↓: Navigate • ENTER: Select • ESC: Cancel search • q: Quit")
	case selectingCWClient:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"↑/↓: Navigate • ENTER: Select • s: Search • ESC: Back • q: Quit")
	case searchingCWClient:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"Type to search • ↑/↓: Navigate • ENTER: Select • ESC: Cancel search • q: Quit")
	case confirmMapping:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"ENTER: Confirm • ESC: Cancel • q: Quit")
	case viewMappings:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			"TAB: Back to mapping • q: Quit")
	}
	return ""
}

// Helper methods for filtering and searching

func (m *MappingModel) filterSlideClients(query string) {
	if query == "" {
		m.filteredSlideClients = m.slideClients
		return
	}

	// Create searchable strings
	searchStrings := make([]string, len(m.slideClients))
	for i, client := range m.slideClients {
		searchStrings[i] = client.Name
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, searchStrings)

	// Build filtered list
	m.filteredSlideClients = make([]models.SlideClient, len(matches))
	for i, match := range matches {
		m.filteredSlideClients[i] = m.slideClients[match.Index]
	}
}

func (m *MappingModel) filterCWClients(query string) {
	// Use the directly stored selected client for more reliability
	var selectedSlideClient models.SlideClient
	if m.selectedSlideClient != nil {
		selectedSlideClient = *m.selectedSlideClient
	} else {
		// Fallback to index-based selection if needed
		if m.selectedSlide < 0 || m.selectedSlide >= len(m.slideClients) {
			// Fallback to showing all clients if selection is invalid
			m.filteredCWClients = m.cwClients
			return
		}
		selectedSlideClient = m.slideClients[m.selectedSlide]
	}

	var baseClients []models.ConnectWiseClient
	var baseSearchStrings []string

	if query == "" {
		// When no search query, show similar clients first based on selected Slide client
		similar := m.findSimilarCWClients(selectedSlideClient.Name)
		remaining := m.getRemainingCWClients(similar)
		baseClients = append(similar, remaining...)
	} else {
		baseClients = m.cwClients
	}

	// Create searchable strings
	baseSearchStrings = make([]string, len(baseClients))
	for i, client := range baseClients {
		baseSearchStrings[i] = client.Name
	}

	if query == "" {
		m.filteredCWClients = baseClients
		return
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, baseSearchStrings)

	// Build filtered list
	m.filteredCWClients = make([]models.ConnectWiseClient, len(matches))
	for i, match := range matches {
		m.filteredCWClients[i] = baseClients[match.Index]
	}
}

func (m *MappingModel) findSimilarCWClients(slideName string) []models.ConnectWiseClient {
	var similar []models.ConnectWiseClient
	slideNameLower := strings.ToLower(slideName)

	// Remove common business suffixes for better matching
	cleanSlideName := cleanCompanyName(slideNameLower)

	for _, cwClient := range m.cwClients {
		cwNameLower := strings.ToLower(cwClient.Name)
		cleanCWName := cleanCompanyName(cwNameLower)

		// Check various similarity conditions
		if strings.Contains(cleanSlideName, cleanCWName) ||
		   strings.Contains(cleanCWName, cleanSlideName) ||
		   strings.HasPrefix(cleanSlideName, cleanCWName) ||
		   strings.HasPrefix(cleanCWName, cleanSlideName) {
			similar = append(similar, cwClient)
		}
	}

	return similar
}

func (m *MappingModel) getRemainingCWClients(exclude []models.ConnectWiseClient) []models.ConnectWiseClient {
	excludeMap := make(map[int]bool)
	for _, client := range exclude {
		excludeMap[client.ID] = true
	}

	var remaining []models.ConnectWiseClient
	for _, client := range m.cwClients {
		if !excludeMap[client.ID] {
			remaining = append(remaining, client)
		}
	}

	return remaining
}

func cleanCompanyName(name string) string {
	// Remove common business suffixes for better matching
	suffixes := []string{", llc", " llc", ", inc", " inc", ", corp", " corp",
		", co", " co", ", ltd", " ltd", ", pllc", " pllc"}

	lower := strings.ToLower(name)
	for _, suffix := range suffixes {
		lower = strings.TrimSuffix(lower, suffix)
	}

	return strings.TrimSpace(lower)
}

func (m *MappingModel) findSlideClientIndex(id string) int {
	for i, client := range m.slideClients {
		if client.ID == id {
			return i
		}
	}
	return -1 // Return -1 to indicate not found
}

func (m *MappingModel) findCWClientIndex(id int) int {
	for i, client := range m.cwClients {
		if client.ID == id {
			return i
		}
	}
	return 0
}