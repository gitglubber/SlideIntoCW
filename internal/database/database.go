package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
	"slide-cw-integration/pkg/models"
)

type DB struct {
	conn *sql.DB
}

func Initialize() (*DB, error) {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./slide_cw_integration.db"
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}

func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS client_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slide_client_id TEXT UNIQUE NOT NULL,
			slide_client_name TEXT NOT NULL,
			connectwise_id INTEGER NOT NULL,
			connectwise_name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alert_ticket_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alert_id TEXT UNIQUE NOT NULL,
			ticket_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS ticketing_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			board_id INTEGER NOT NULL,
			board_name TEXT NOT NULL,
			status_id INTEGER NOT NULL,
			status_name TEXT NOT NULL,
			priority_id INTEGER NOT NULL,
			priority_name TEXT NOT NULL,
			type_id INTEGER NOT NULL,
			type_name TEXT NOT NULL,
			ticket_summary TEXT NOT NULL DEFAULT 'Slide Alert: {{alert_type}} for {{client_name}}',
			ticket_template TEXT NOT NULL DEFAULT 'Alert Details:\n\nClient: {{client_name}}\nDevice: {{device_name}}\nAlert Type: {{alert_type}}\nMessage: {{alert_message}}\nTimestamp: {{alert_timestamp}}\n\nThis ticket was automatically created by the Slide-ConnectWise integration.',
			auto_assign_tech BOOLEAN DEFAULT FALSE,
			technician_id INTEGER,
			technician_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (db *DB) SaveClientMapping(mapping *models.ClientMapping) error {
	query := `INSERT OR REPLACE INTO client_mappings
		(slide_client_id, slide_client_name, connectwise_id, connectwise_name)
		VALUES (?, ?, ?, ?)`

	_, err := db.conn.Exec(query, mapping.SlideClientID, mapping.SlideClientName,
		mapping.ConnectWiseID, mapping.ConnectWiseName)

	return err
}

func (db *DB) GetClientMapping(slideClientID string) (*models.ClientMapping, error) {
	query := `SELECT id, slide_client_id, slide_client_name, connectwise_id, connectwise_name, created_at
		FROM client_mappings WHERE slide_client_id = ?`

	var mapping models.ClientMapping
	err := db.conn.QueryRow(query, slideClientID).Scan(
		&mapping.ID, &mapping.SlideClientID, &mapping.SlideClientName,
		&mapping.ConnectWiseID, &mapping.ConnectWiseName, &mapping.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &mapping, err
}

func (db *DB) SaveAlertTicketMapping(mapping *models.AlertTicketMapping) error {
	query := `INSERT INTO alert_ticket_mappings (alert_id, ticket_id) VALUES (?, ?)`
	_, err := db.conn.Exec(query, mapping.AlertID, mapping.TicketID)
	return err
}

func (db *DB) GetAlertTicketMapping(alertID string) (*models.AlertTicketMapping, error) {
	query := `SELECT id, alert_id, ticket_id, created_at, closed_at
		FROM alert_ticket_mappings WHERE alert_id = ?`

	var mapping models.AlertTicketMapping
	err := db.conn.QueryRow(query, alertID).Scan(
		&mapping.ID, &mapping.AlertID, &mapping.TicketID,
		&mapping.CreatedAt, &mapping.ClosedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &mapping, err
}

func (db *DB) CloseAlertTicketMapping(alertID string) error {
	query := `UPDATE alert_ticket_mappings SET closed_at = CURRENT_TIMESTAMP WHERE alert_id = ?`
	_, err := db.conn.Exec(query, alertID)
	return err
}

// Ticketing configuration methods
func (db *DB) SaveTicketingConfig(config *models.TicketingConfig) error {
	query := `INSERT OR REPLACE INTO ticketing_config
		(board_id, board_name, status_id, status_name, priority_id, priority_name,
		 type_id, type_name, ticket_summary, ticket_template, auto_assign_tech,
		 technician_id, technician_name, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := db.conn.Exec(query,
		config.BoardID, config.BoardName,
		config.StatusID, config.StatusName,
		config.PriorityID, config.PriorityName,
		config.TypeID, config.TypeName,
		config.TicketSummary, config.TicketTemplate,
		config.AutoAssignTech, config.TechnicianID, config.TechnicianName,
	)

	return err
}

func (db *DB) GetTicketingConfig() (*models.TicketingConfig, error) {
	query := `SELECT id, board_id, board_name, status_id, status_name, priority_id, priority_name,
		type_id, type_name, ticket_summary, ticket_template, auto_assign_tech,
		technician_id, technician_name, created_at, updated_at
		FROM ticketing_config ORDER BY updated_at DESC LIMIT 1`

	var config models.TicketingConfig
	err := db.conn.QueryRow(query).Scan(
		&config.ID, &config.BoardID, &config.BoardName,
		&config.StatusID, &config.StatusName,
		&config.PriorityID, &config.PriorityName,
		&config.TypeID, &config.TypeName,
		&config.TicketSummary, &config.TicketTemplate,
		&config.AutoAssignTech, &config.TechnicianID, &config.TechnicianName,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &config, err
}

func (db *DB) DeleteTicketingConfig() error {
	query := `DELETE FROM ticketing_config`
	_, err := db.conn.Exec(query)
	return err
}