package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("cmd/server/swimming_event.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== TOURNAMENTS IN cmd/server/swimming_event.db ===")
	var tourneys []map[string]interface{}
	db.Table("tournaments").Find(&tourneys)
	for _, t := range tourneys {
		fmt.Printf("Tourney #%v: %v | StartDate: %v | Active: %v\n", t["id"], t["name"], t["event_start_date"], t["is_active"])
	}

	fmt.Println("\n=== EVENTS IN cmd/server/swimming_event.db ===")
	var events []map[string]interface{}
	db.Table("swimming_events").Order("event_code asc").Find(&events)
	for _, e := range events {
		fmt.Printf("Event %v: %v | Dist: %v | Stroke: %v | Gender: %v | AgeGroup: %v\n",
			e["event_code"], e["event_name"], e["distance"], e["stroke"], e["gender"], e["age_group"])
	}

	fmt.Println("\n=== REGISTRATIONS / PARTICIPANTS IN cmd/server/swimming_event.db ===")
	var parts []map[string]interface{}
	db.Table("participants").Find(&parts)
	fmt.Printf("Total Participants: %d\n", len(parts))
	for _, p := range parts {
		fmt.Printf("P #%v: %v (%v, %v, %v)\n", p["id"], p["name"], p["club"], p["gender"], p["age_group"])
	}
}
