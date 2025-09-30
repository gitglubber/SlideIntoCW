package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./slide_cw_integration.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT slide_client_id, slide_client_name, connectwise_name FROM client_mappings")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Database Mappings:")
	fmt.Println("==================")
	for rows.Next() {
		var slideID, slideName, cwName string
		rows.Scan(&slideID, &slideName, &cwName)
		fmt.Printf("Slide ID: %s, Name: %s → CW: %s\n", slideID, slideName, cwName)
	}
}
