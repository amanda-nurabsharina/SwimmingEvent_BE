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

	var results []map[string]interface{}
	db.Table("registrations").
		Select("registrations.id, participants.name, participants.club, swimming_events.event_code, swimming_events.event_name, registrations.race_result_time, registrations.final_result_time, registrations.rank, registrations.final_rank").
		Joins("JOIN participants ON participants.id = registrations.participant_id").
		Joins("JOIN swimming_events ON swimming_events.id = registrations.swimming_event_id").
		Where("registrations.race_result_time != '' OR registrations.final_result_time != ''").
		Find(&results)

	fmt.Printf("Registrations with recorded results: %d\n", len(results))
	for _, r := range results {
		fmt.Printf("RegID: %v | Name: %v | Club: %v | Event: %v (%v) | RaceResult: %v | FinalResult: %v | Rank: %v | FinalRank: %v\n",
			r["id"], r["name"], r["club"], r["event_code"], r["event_name"], r["race_result_time"], r["final_result_time"], r["rank"], r["final_rank"])
	}
}
