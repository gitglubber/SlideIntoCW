package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./slide_cw_integration.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Reset the closed_at timestamp for alert al_nxjisekv7byi
	// This will make the monitoring loop try to close it again
	result, err := db.Exec("UPDATE alert_ticket_mappings SET closed_at = NULL WHERE alert_id = ?", "al_nxjisekv7byi")
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Reset %d alert-ticket mappings", rowsAffected)

	// Show the current state
	var alertID string
	var ticketID int
	var closedAt interface{}

	err = db.QueryRow("SELECT alert_id, ticket_id, closed_at FROM alert_ticket_mappings WHERE alert_id = ?", "al_nxjisekv7byi").Scan(&alertID, &ticketID, &closedAt)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Alert: %s, Ticket: %d, Closed At: %v", alertID, ticketID, closedAt)
}