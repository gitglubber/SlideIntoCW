package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"slide-cw-integration/pkg/models"
)

type ticketingState int

const (
	selectingBoard ticketingState = iota
	selectingStatus
	selectingPriority
	selectingType
	configuringSummary
	configuringTemplate
	configuringTechnician
	reviewingConfig
)

type TicketingModel struct {
	boards       []models.ConnectWiseBoard
	statuses     []models.ConnectWiseStatus
	priorities   []models.ConnectWisePriority
	types        []models.ConnectWiseType
	members      []models.ConnectWiseMember

	config       models.TicketingConfig
	currentState ticketingState

	boardCursor     int
	statusCursor    int
	priorityCursor  int
	typeCursor      int
	memberCursor    int

	summaryInput   string
	templateInput  string

	message      string
	loading      bool
	onSave       func(*models.TicketingConfig) error
	onLoadConfig func(boardID int) ([]models.ConnectWiseStatus, []models.ConnectWiseType, error)
}

// Styles
var (
	ticketingHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981")).
		Padding(0, 1)

	ticketingSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1)

	ticketingNormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Padding(0, 1)

	ticketingInputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#374151")).
		Padding(0, 1).
		Width(60)

	ticketingLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)
)

func NewTicketingModel(
	boards []models.ConnectWiseBoard,
	priorities []models.ConnectWisePriority,
	members []models.ConnectWiseMember,
	onSave func(*models.TicketingConfig) error,
	onLoadConfig func(boardID int) ([]models.ConnectWiseStatus, []models.ConnectWiseType, error),
) TicketingModel {
	config := models.TicketingConfig{
		TicketSummary:  "Slide Alert: {{alert_type}} for {{client_name}}",
		TicketTemplate: "Alert Details:\n\nClient: {{client_name}}\nDevice: {{device_name}}\nAlert Type: {{alert_type}}\nMessage: {{alert_message}}\nTimestamp: {{alert_timestamp}}\n\nThis ticket was automatically created by the Slide-ConnectWise integration.",
		AutoAssignTech: false,
	}

	return TicketingModel{
		boards:       boards,
		priorities:   priorities,
		members:      members,
		config:       config,
		currentState: selectingBoard,
		summaryInput: config.TicketSummary,
		templateInput: config.TicketTemplate,
		onSave:       onSave,
		onLoadConfig: onLoadConfig,
	}
}

func (m TicketingModel) Init() tea.Cmd {
	return nil
}

func (m TicketingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			switch m.currentState {
			case selectingBoard:
				if len(m.boards) > 0 && m.boardCursor < len(m.boards) {
					selectedBoard := m.boards[m.boardCursor]
					m.config.BoardID = selectedBoard.ID
					m.config.BoardName = selectedBoard.Name
					m.currentState = selectingStatus
					m.loading = true
					// Load statuses and types for this board
					return m, func() tea.Msg {
						statuses, types, err := m.onLoadConfig(selectedBoard.ID)
						if err != nil {
							return errorMsg{err: err}
						}
						return configLoadedMsg{
							statuses: statuses,
							types:    types,
						}
					}
				}

			case selectingStatus:
				if len(m.statuses) > 0 && m.statusCursor < len(m.statuses) {
					selectedStatus := m.statuses[m.statusCursor]
					m.config.StatusID = selectedStatus.ID
					m.config.StatusName = selectedStatus.Name
					m.currentState = selectingPriority
				}

			case selectingPriority:
				if len(m.priorities) > 0 && m.priorityCursor < len(m.priorities) {
					selectedPriority := m.priorities[m.priorityCursor]
					m.config.PriorityID = selectedPriority.ID
					m.config.PriorityName = selectedPriority.Name
					m.currentState = selectingType
				}

			case selectingType:
				if len(m.types) > 0 && m.typeCursor < len(m.types) {
					selectedType := m.types[m.typeCursor]
					m.config.TypeID = selectedType.ID
					m.config.TypeName = selectedType.Name
					m.currentState = configuringSummary
				}

			case configuringSummary:
				m.config.TicketSummary = m.summaryInput
				m.currentState = configuringTemplate

			case configuringTemplate:
				m.config.TicketTemplate = m.templateInput
				m.currentState = configuringTechnician

			case configuringTechnician:
				if len(m.members) > 0 && m.memberCursor < len(m.members) && m.config.AutoAssignTech {
					selectedMember := m.members[m.memberCursor]
					m.config.TechnicianID = &selectedMember.ID
					m.config.TechnicianName = selectedMember.FirstName + " " + selectedMember.LastName
				}
				m.currentState = reviewingConfig

			case reviewingConfig:
				// Save configuration
				if err := m.onSave(&m.config); err != nil {
					m.message = fmt.Sprintf("Error saving configuration: %v", err)
				} else {
					m.message = "✓ Ticketing configuration saved successfully!"
				}
				return m, nil
			}

		case "esc":
			switch m.currentState {
			case selectingStatus:
				m.currentState = selectingBoard
			case selectingPriority:
				m.currentState = selectingStatus
			case selectingType:
				m.currentState = selectingPriority
			case configuringSummary:
				m.currentState = selectingType
			case configuringTemplate:
				m.currentState = configuringSummary
			case configuringTechnician:
				m.currentState = configuringTemplate
			case reviewingConfig:
				m.currentState = configuringTechnician
			}

		case "up", "k":
			switch m.currentState {
			case selectingBoard:
				if m.boardCursor > 0 {
					m.boardCursor--
				}
			case selectingStatus:
				if m.statusCursor > 0 {
					m.statusCursor--
				}
			case selectingPriority:
				if m.priorityCursor > 0 {
					m.priorityCursor--
				}
			case selectingType:
				if m.typeCursor > 0 {
					m.typeCursor--
				}
			case configuringTechnician:
				if m.memberCursor > 0 {
					m.memberCursor--
				}
			}

		case "down", "j":
			switch m.currentState {
			case selectingBoard:
				if m.boardCursor < len(m.boards)-1 {
					m.boardCursor++
				}
			case selectingStatus:
				if m.statusCursor < len(m.statuses)-1 {
					m.statusCursor++
				}
			case selectingPriority:
				if m.priorityCursor < len(m.priorities)-1 {
					m.priorityCursor++
				}
			case selectingType:
				if m.typeCursor < len(m.types)-1 {
					m.typeCursor++
				}
			case configuringTechnician:
				if m.memberCursor < len(m.members)-1 {
					m.memberCursor++
				}
			}

		case "tab":
			if m.currentState == configuringTechnician {
				m.config.AutoAssignTech = !m.config.AutoAssignTech
			}

		case "backspace":
			switch m.currentState {
			case configuringSummary:
				if len(m.summaryInput) > 0 {
					m.summaryInput = m.summaryInput[:len(m.summaryInput)-1]
				}
			case configuringTemplate:
				if len(m.templateInput) > 0 {
					m.templateInput = m.templateInput[:len(m.templateInput)-1]
				}
			}

		default:
			// Handle text input for summary and template
			if len(msg.String()) == 1 {
				switch m.currentState {
				case configuringSummary:
					m.summaryInput += msg.String()
				case configuringTemplate:
					m.templateInput += msg.String()
				}
			}
		}

	case configLoadedMsg:
		m.loading = false
		m.statuses = msg.statuses
		m.types = msg.types
		m.statusCursor = 0
		m.typeCursor = 0

	case errorMsg:
		m.loading = false
		m.message = fmt.Sprintf("Error loading configuration: %v", msg.err)
	}

	return m, nil
}

type configLoadedMsg struct {
	statuses []models.ConnectWiseStatus
	types    []models.ConnectWiseType
}

type errorMsg struct {
	err error
}

func (m TicketingModel) View() string {
	var b strings.Builder

	b.WriteString(ticketingHeaderStyle.Render("🎫 ConnectWise Ticketing Configuration"))
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(m.message)
		b.WriteString("\n\n")
	}

	switch m.currentState {
	case selectingBoard:
		b.WriteString(m.renderBoardSelection())
	case selectingStatus:
		b.WriteString(m.renderStatusSelection())
	case selectingPriority:
		b.WriteString(m.renderPrioritySelection())
	case selectingType:
		b.WriteString(m.renderTypeSelection())
	case configuringSummary:
		b.WriteString(m.renderSummaryConfiguration())
	case configuringTemplate:
		b.WriteString(m.renderTemplateConfiguration())
	case configuringTechnician:
		b.WriteString(m.renderTechnicianConfiguration())
	case reviewingConfig:
		b.WriteString(m.renderConfigReview())
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelp())
	return b.String()
}

func (m TicketingModel) renderBoardSelection() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render("Select Service Board:"))
	b.WriteString("\n\n")

	for i, board := range m.boards {
		style := ticketingNormalStyle
		prefix := "  "
		if i == m.boardCursor {
			style = ticketingSelectedStyle
			prefix = "► "
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, board.Name)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m TicketingModel) renderStatusSelection() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render(fmt.Sprintf("Select Status for Board: %s", m.config.BoardName)))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ticketingLabelStyle.Render("Loading statuses and types..."))
		return b.String()
	}

	if len(m.statuses) == 0 {
		b.WriteString(ticketingLabelStyle.Render("No active statuses found for this board"))
		return b.String()
	}

	for i, status := range m.statuses {
		style := ticketingNormalStyle
		prefix := "  "
		if i == m.statusCursor {
			style = ticketingSelectedStyle
			prefix = "► "
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, status.Name)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m TicketingModel) renderPrioritySelection() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render("Select Priority:"))
	b.WriteString("\n\n")

	for i, priority := range m.priorities {
		style := ticketingNormalStyle
		prefix := "  "
		if i == m.priorityCursor {
			style = ticketingSelectedStyle
			prefix = "► "
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, priority.Name)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m TicketingModel) renderTypeSelection() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render(fmt.Sprintf("Select Type for Board: %s", m.config.BoardName)))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(ticketingLabelStyle.Render("Loading types..."))
		return b.String()
	}

	if len(m.types) == 0 {
		b.WriteString(ticketingLabelStyle.Render("No active types found for this board"))
		return b.String()
	}

	for i, ticketType := range m.types {
		style := ticketingNormalStyle
		prefix := "  "
		if i == m.typeCursor {
			style = ticketingSelectedStyle
			prefix = "► "
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, ticketType.Name)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m TicketingModel) renderSummaryConfiguration() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render("Configure Ticket Summary Template:"))
	b.WriteString("\n\n")

	b.WriteString(ticketingLabelStyle.Render("Available variables: {{alert_type}}, {{client_name}}, {{device_name}}"))
	b.WriteString("\n\n")

	b.WriteString(ticketingInputStyle.Render(m.summaryInput + "█"))
	b.WriteString("\n")

	return b.String()
}

func (m TicketingModel) renderTemplateConfiguration() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render("Configure Ticket Description Template:"))
	b.WriteString("\n\n")

	b.WriteString(ticketingLabelStyle.Render("Available variables: {{alert_type}}, {{client_name}}, {{device_name}}, {{alert_message}}, {{alert_timestamp}}"))
	b.WriteString("\n\n")

	// Multi-line input display
	lines := strings.Split(m.templateInput, "\n")
	for _, line := range lines {
		b.WriteString(ticketingInputStyle.Render(line))
		b.WriteString("\n")
	}
	b.WriteString(ticketingInputStyle.Render("█"))
	b.WriteString("\n")

	return b.String()
}

func (m TicketingModel) renderTechnicianConfiguration() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render("Configure Auto-Assignment:"))
	b.WriteString("\n\n")

	autoAssignText := "Disabled"
	if m.config.AutoAssignTech {
		autoAssignText = "Enabled"
	}

	b.WriteString(ticketingLabelStyle.Render(fmt.Sprintf("Auto-assign technician: %s (Press TAB to toggle)", autoAssignText)))
	b.WriteString("\n\n")

	if m.config.AutoAssignTech {
		b.WriteString(ticketingHeaderStyle.Render("Select Technician:"))
		b.WriteString("\n\n")

		for i, member := range m.members {
			if member.Inactive {
				continue
			}

			style := ticketingNormalStyle
			prefix := "  "
			if i == m.memberCursor {
				style = ticketingSelectedStyle
				prefix = "► "
			}

			name := fmt.Sprintf("%s %s", member.FirstName, member.LastName)
			if member.Title != "" {
				name += fmt.Sprintf(" (%s)", member.Title)
			}

			b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, name)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m TicketingModel) renderConfigReview() string {
	var b strings.Builder
	b.WriteString(ticketingHeaderStyle.Render("Review Configuration:"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Board: %s\n", m.config.BoardName))
	b.WriteString(fmt.Sprintf("Status: %s\n", m.config.StatusName))
	b.WriteString(fmt.Sprintf("Priority: %s\n", m.config.PriorityName))
	b.WriteString(fmt.Sprintf("Type: %s\n", m.config.TypeName))
	b.WriteString(fmt.Sprintf("Summary Template: %s\n", m.config.TicketSummary))
	b.WriteString(fmt.Sprintf("Auto-assign: %t\n", m.config.AutoAssignTech))
	if m.config.AutoAssignTech && m.config.TechnicianName != "" {
		b.WriteString(fmt.Sprintf("Technician: %s\n", m.config.TechnicianName))
	}
	b.WriteString("\n")

	b.WriteString(ticketingSelectedStyle.Render("Press ENTER to save configuration"))
	b.WriteString("\n")

	return b.String()
}

func (m TicketingModel) renderHelp() string {
	switch m.currentState {
	case selectingBoard, selectingStatus, selectingPriority, selectingType:
		return ticketingLabelStyle.Render("↑/↓: Navigate • ENTER: Select • ESC: Back • q: Quit")
	case configuringSummary, configuringTemplate:
		return ticketingLabelStyle.Render("Type to edit • ENTER: Continue • ESC: Back • q: Quit")
	case configuringTechnician:
		return ticketingLabelStyle.Render("↑/↓: Navigate • TAB: Toggle auto-assign • ENTER: Continue • ESC: Back • q: Quit")
	case reviewingConfig:
		return ticketingLabelStyle.Render("ENTER: Save configuration • ESC: Back • q: Quit")
	default:
		return ticketingLabelStyle.Render("q: Quit")
	}
}