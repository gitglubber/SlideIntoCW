# Diagnostic Tools

This folder contains diagnostic and testing scripts used during development.

## Database Utilities

- `check_db.go` - Check database contents
- `check_clients.go` - List Slide clients
- `query_db.sql` - Sample SQL queries
- `clear-mappings.sql` - SQL to clear all mappings
- `debug-mappings.sql` - SQL to debug mapping issues

## Alert Testing

- `check_alert_device.go` - Check which client owns a specific device
- `check_device_client.go` - Look up device-to-client relationships
- `test_alerts.go` - Test alert retrieval from Slide API
- `test_alerts_debug.go` - Debug alert processing logic
- `debug_alerts_api.go` - Test Slide alerts API endpoint

## Client Mapping

- `list_all_clients.go` - List all Slide clients with initials
- `find_device_client.go` - Find which client a device belongs to
- `smart_match.go` - Test smart device name matching
- `test_bm_match.go` - Test "BM" → "Barnett & Moro" matching
- `add_account_mapping.go` - Manually add account mapping
- `fix_account_mapping.go` - Fix broken account mappings
- `debug_mapping.go` - Debug mapping service

## Ticket Management

- `reset_alert_mapping.go` - Reset alert-ticket mapping (for testing)
- `clear_alert_mappings.go` - Clear all alert-ticket mappings
- `debug_tickets.go` - Debug ticket creation
- `debug_cw_company.go` - Test ConnectWise company lookup
- `update_template.go` - Update ticketing templates

## Other

- `test_clients.go` - Test Slide client API
- `test_db.go` - Test database operations

## Usage

Most of these scripts can be run with:

```bash
go run diag/script-name.go
```

**Note:** These are development tools and are not needed for normal operation of the application.