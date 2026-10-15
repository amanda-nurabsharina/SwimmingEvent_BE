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

	// Check heats / registrations in buku acara
	var heats []map[string]interface{}
	db.Table("registrations").
		Select("registrations.id, registrations.participant_id, registrations.swimming_event_id, registrations.heat_number, registrations.lane_number, registrations.race_result_time, registrations.final_result_time, registrations.final_rank, participants.name as swimmer_name, participants.club, swimming_events.event_code, swimming_events.event_name, swimming_events.distance, swimming_events.stroke, swimming_events.gender, swimming_events.age_group").
		Joins("JOIN participants ON participants.id = registrations.participant_id").
		Joins("JOIN swimming_events ON swimming_events.id = registrations.swimming_event_id").
		Where("registrations.is_active = 1").
		Order("swimming_events.event_code asc, registrations.heat_number asc, registrations.lane_number asc").
		Find(&heats)

	fmt.Printf("Total Active Registrations: %d\n", len(heats))
	countsByEvent := make(map[string]int)
	for _, h := range heats {
		evKey := fmt.Sprintf("#%v - %v", h["event_code"], h["event_name"])
		countsByEvent[evKey]++
	}
	for evKey, count := range countsByEvent {
		fmt.Printf("Event %s: %d swimmers\n", evKey, count)
	}
}
