package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("swimming_event.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	var regCols []struct {
		Cid       int
		Name      string
		Type      string
		NotNull   int
		DfltValue *string
		Pk        int
	}
	db.Raw("PRAGMA table_info(registrations)").Scan(&regCols)
	fmt.Println("=== REGISTRATIONS COLUMNS ===")
	for _, c := range regCols {
		fmt.Printf("Col: %s (%s)\n", c.Name, c.Type)
	}

	var regs []map[string]interface{}
	db.Table("registrations").
		Select("registrations.*, participants.name as p_name, participants.club as p_club, participants.gender as p_gender, participants.age_group as p_age_group, swimming_events.event_code, swimming_events.event_name").
		Joins("LEFT JOIN participants ON participants.id = registrations.participant_id").
		Joins("LEFT JOIN swimming_events ON swimming_events.id = registrations.swimming_event_id").
		Find(&regs)

	fmt.Printf("\nTotal Registrations: %d\n", len(regs))
	for _, r := range regs {
		fmt.Printf("Reg #%v | Swimmer: %v | Club: %v | EvCode: %v (%v) | Status: %v | Active: %v | Heat: %v | Lane: %v | Seed: %v | FinalTime: %v | Rank: %v\n",
			r["id"], r["p_name"], r["p_club"], r["event_code"], r["event_name"], r["status"], r["is_active"], r["heat_number"], r["lane_number"], r["time_seed"], r["final_result_time"], r["final_rank"])
	}
}
