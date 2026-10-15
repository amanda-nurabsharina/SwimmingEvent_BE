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

	fmt.Println("=== TOURNAMENTS ===")
	var tournaments []map[string]interface{}
	db.Table("tournaments").Find(&tournaments)
	for _, t := range tournaments {
		fmt.Printf("Tourney ID: %v | Name: %v | Active: %v | StartDate: %v | BukuLocked: %v | BukuPublished: %v\n",
			t["id"], t["name"], t["is_active"], t["event_start_date"], t["is_buku_acara_locked"], t["is_buku_acara_published"])
	}

	fmt.Println("\n=== SWIMMING EVENTS ===")
	var events []map[string]interface{}
	db.Table("swimming_events").Order("event_code asc").Find(&events)
	for _, e := range events {
		fmt.Printf("Event ID: %v | Code: %v | TourneyID: %v | Name: %v | Dist: %v | Stroke: %v | Gender: %v | AgeGroup: %v | Category: %v | Sched: %v %v\n",
			e["id"], e["event_code"], e["tournament_id"], e["event_name"], e["distance"], e["stroke"], e["gender"], e["age_group"], e["heat_category"], e["schedule_date"], e["schedule_time"])
	}

	fmt.Println("\n=== REGISTRATIONS ===")
	var regs []map[string]interface{}
	db.Table("registrations").
		Select("registrations.id, registrations.tournament_id, registrations.swimming_event_id, registrations.participant_id, registrations.status, registrations.is_active, registrations.heat_number, registrations.lane_number, registrations.final_rank, registrations.final_result_time, registrations.race_result_time, registrations.time_seed, participants.name as p_name, participants.club as p_club, participants.gender as p_gender, participants.age_group as p_age_group, swimming_events.event_code, swimming_events.event_name, swimming_events.distance, swimming_events.stroke").
		Joins("LEFT JOIN participants ON participants.id = registrations.participant_id").
		Joins("LEFT JOIN swimming_events ON swimming_events.id = registrations.swimming_event_id").
		Order("swimming_events.event_code asc, registrations.heat_number asc, registrations.lane_number asc").
		Find(&regs)
	fmt.Printf("Total Registrations: %d\n", len(regs))
	for _, r := range regs {
		fmt.Printf("RegID: %v | EventCode: %v (%v %v) | Swimmer: %v | Club: %v | Status: %v | Active: %v | Heat: %v Lane: %v | Seed: %v | FinalTime: %v | Rank: %v\n",
			r["id"], r["event_code"], r["distance"], r["stroke"], r["p_name"], r["p_club"], r["status"], r["is_active"], r["heat_number"], r["lane_number"], r["time_seed"], r["final_result_time"], r["final_rank"])
	}
}
