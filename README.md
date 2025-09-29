# Slide-ConnectWise Integration

A Go application that integrates Slide backup monitoring API with ConnectWise Manage for automated ticket management based on backup alerts.

## Features

- **🌐 Modern Web UI**: Full-featured web interface for easy configuration and monitoring
- **🚨 Alert Monitoring**: Automatically monitors Slide API for backup failures and other alerts every 5 minutes
- **🗺️ Client Mapping**: Maps Slide clients/accounts to ConnectWise companies with visual interface
- **🎫 Automated Ticketing**: Creates ConnectWise tickets for unresolved alerts using configurable templates
- **🚫 Duplicate Prevention**: Prevents multiple tickets for the same alert
- **🔄 Bidirectional Auto-Resolution**:
  - Detects when backup issues are resolved and closes both alerts and tickets
  - Detects when ConnectWise tickets are manually closed and closes corresponding Slide alerts
- **📋 Rich Templates**: Includes client, device, agent, and error details in tickets
- **⚙️ Easy Configuration**: Web-based configuration with template preview
- **📊 Real-time Dashboard**: Monitor alerts, tickets, and mappings at a glance

## Setup

### Prerequisites

- Go 1.19 or later
- Slide API access (API URL and API Key)
- ConnectWise Manage API access (API URL, Company ID, Public Key, Private Key, Client ID)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd SlideIntoCW
```

2. Install dependencies:
```bash
go mod tidy
```

3. Create `.env` file with your API credentials:
```bash
# Slide API Configuration
SLIDE_API_URL=https://api.slide.tech
SLIDE_API_KEY=your_slide_api_key

# ConnectWise API Configuration
CONNECTWISE_API_URL=https://your-instance.connectwisedev.com/v4_6_release/apis/3.0
CONNECTWISE_COMPANY_ID=your_company_id
CONNECTWISE_PUBLIC_KEY=your_public_key
CONNECTWISE_PRIVATE_KEY=your_private_key
CONNECTWISE_CLIENT_ID=your_client_id

# Optional: Database path (defaults to ./slide_cw_integration.db)
DATABASE_PATH=./slide_cw_integration.db
```

4. Build the application:
```bash
go build -o slide-integrator.exe ./cmd/slide-integrator
```

## Project Structure

```
├── cmd/slide-integrator/     # Main application entry point
├── internal/
│   ├── web/                 # Web UI server and API
│   ├── slide/               # Slide API client
│   ├── connectwise/         # ConnectWise API client
│   ├── database/            # Database operations
│   ├── mapping/             # Client mapping logic
│   ├── tui/                 # Terminal UI (legacy)
│   └── alerts/              # Alert monitoring and processing
├── pkg/models/              # Data models
└── .env.example            # Environment configuration template
```

## Usage

### 🌐 Web UI (Recommended)

Start the web interface with built-in alert monitoring:

```bash
./slide-integrator.exe -web
```

Then open your browser to: **http://localhost:8080**

The web UI provides:
- **Dashboard**: Real-time stats on alerts, mappings, and tickets
- **Client Mappings**: Visual interface to map Slide clients to ConnectWise companies
- **Ticketing Config**: Easy form-based configuration with template preview
- **Alerts View**: Browse and manage all alerts
- **Tickets View**: Monitor alert-to-ticket mappings

Custom port:
```bash
./slide-integrator.exe -web -port 3000
```

### 📋 CLI Commands (Legacy)

For headless server environments, you can still use CLI commands:

1. **Map Clients**: Map Slide clients to ConnectWise companies
```bash
./slide-integrator.exe -map-interactive
```

2. **Setup Ticketing**: Configure ConnectWise board, status, priority, and type
```bash
./slide-integrator.exe -setup-ticketing
```

### Management Commands

- **Show current mappings**:
```bash
./slide-integrator.exe -show-mappings
```

- **Clear all mappings**:
```bash
./slide-integrator.exe -clear-mappings
```

- **Auto-map clients** (uses fuzzy matching):
```bash
./slide-integrator.exe -map-clients
```

### Running the Service

**Option 1: Web UI Mode (Recommended)**
```bash
./slide-integrator.exe -web
```
This starts both the web interface and alert monitoring service together.

**Option 2: CLI-Only Mode**
```bash
./slide-integrator.exe
```
This runs only the alert monitoring service without the web UI.

The service will:
- Check for alerts every 5 minutes
- Create tickets for unresolved alerts
- Monitor for alert resolution and close tickets automatically
- Detect manually closed ConnectWise tickets and close corresponding Slide alerts

### 🎯 **ENHANCED: Interactive Mapping TUI with Smart Search**
The `-map-interactive` command now features **powerful search and intelligent matching**:

**🔍 Advanced Search Features:**
- **Fuzzy Text Search** - Press `s` to search by typing client names
- **Smart Suggestions** - Similar clients automatically shown first
- **Real-time Filtering** - Results update as you type
- **Company Name Cleaning** - Ignores LLC, Inc, Corp differences
- **Complete Client Lists** - Fetches ALL active clients via API pagination
- **Performance Optimized** - Handles thousands of clients smoothly

**🎨 Enhanced Interface:**
- 💡 **Similar clients highlighted** in blue with lightbulb icons
- ✅ **Mapped clients** shown in green with arrow to target
- 🔍 **Search mode** with live input field and cursor
- 📊 **Pagination indicators** showing "X of Y clients"
- 🎯 **Targeted suggestions** based on selected Slide client

**🚀 Improved Workflow:**
1. **Browse or Search** - Use arrows OR press `s` to search Slide clients
2. **Smart Matching** - When you select a Slide client, ConnectWise shows similar clients first
3. **Precise Search** - Use `s` again to search ConnectWise if needed
4. **Quick Selection** - Similar clients highlighted for fast picking

**Navigation:**
- `↑/↓` or `j/k` - Navigate lists
- `s` - **Start search mode**
- `Type` - **Search in real-time**
- `ENTER` - Select client or confirm mapping
- `ESC` - Cancel search or go back
- `TAB` - Switch between mapping and viewing modes
- `q` - Quit

## Configuration

### Ticket Template Variables

The ticket templates support the following variables:

- `{{alert_id}}` - Slide alert ID
- `{{alert_type}}` - Type of alert (e.g., agent_backup_failed)
- `{{alert_message}}` - Error message from the alert
- `{{alert_timestamp}}` - When the alert was created
- `{{client_id}}` - Slide client/account ID
- `{{client_name}}` - Mapped ConnectWise client name
- `{{device_id}}` - Slide device ID
- `{{device_name}}` - Device name from alert
- `{{agent_name}}` - Agent name from alert
- `{{agent_hostname}}` - Agent hostname from alert

### Database Schema

The application uses SQLite with the following tables:

- `client_mappings` - Maps Slide clients to ConnectWise companies
- `alert_ticket_mappings` - Tracks which tickets were created for which alerts
- `ticketing_config` - Stores ConnectWise board, status, priority, and type configuration

## Architecture

```
cmd/slide-integrator/     # Main application entry point
internal/
  ├── alerts/             # Alert monitoring and processing
  ├── connectwise/        # ConnectWise API client
  ├── database/           # SQLite database layer
  ├── mapping/            # Client mapping service
  ├── slide/              # Slide API client
  └── tui/                # Terminal UI for interactive setup
pkg/models/               # Data models and structures
```

## API Integration

### Slide API
- Fetches clients, devices, alerts, and backups
- Closes alerts when issues are resolved
- Handles pagination for large datasets
- Parses complex alert fields for client, device, and agent information

### ConnectWise API
- Creates tickets with full configuration support
- Manages companies, boards, statuses, priorities, and types
- Updates and closes tickets
- Handles API authentication and pagination
- Detects manually closed tickets for bidirectional sync

## How It Works

### Complete Workflow:

1. **Setup Phase**:
   - Run `./slide-integrator.exe -map-interactive` to map Slide accounts to ConnectWise companies
   - Run `./slide-integrator.exe -setup-ticketing` to configure board, status, priority, type

2. **Monitoring Phase** (runs continuously every 5 minutes):
   - **Alert Detection**: Monitors Slide API for unresolved alerts
   - **Ticket Creation**: Creates ConnectWise tickets with rich templates including:
     - Client name (from mapping, not Slide account name)
     - Device and agent information
     - Error messages and timestamps
   - **Duplicate Prevention**: Prevents multiple tickets for the same alert

3. **Resolution Phase** (bidirectional):
   - **Backup Fixed**: When backups succeed after alert → closes both Slide alert and ConnectWise ticket
   - **Manual Close**: When ConnectWise ticket is manually closed → closes corresponding Slide alert
   - **Database Sync**: Updates alert-ticket mappings with closure timestamps

### Alert-to-Ticket Lifecycle:

```
🚨 Slide Alert (Unresolved)
    ↓
📋 Client Mapping Resolution (account → ConnectWise company)
    ↓
🎫 ConnectWise Ticket Creation (with rich template)
    ↓
📊 Database Mapping (alert ↔ ticket relationship)
    ↓
🔄 Continuous Monitoring (every 5 minutes)
    ↓
✅ Resolution Detection:
    • Successful backup → close both alert & ticket
    • Manual ticket close → close corresponding alert
```

## Development

### Adding New Alert Types

1. Update `SlideAlert` model in `pkg/models/models.go`
2. Add handling logic in `internal/alerts/monitor.go`
3. Update ticket templates if needed

### Extending Client Mapping

The fuzzy matching algorithm in `internal/mapping/service.go` can be customized for better automatic client mapping.

## Production Ready!

✅ **Complete Integration** - Handles full alert-to-ticket lifecycle
✅ **Bidirectional Sync** - Works regardless of where closure originates
✅ **Rich Templates** - Includes all relevant alert information
✅ **Interactive Setup** - Easy configuration via TUI
✅ **Robust Error Handling** - Comprehensive logging and error recovery
✅ **Database Persistence** - SQLite for reliable data storage

Just configure your API credentials in `.env` and deploy!