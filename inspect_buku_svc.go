package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"serve-swimming-be/internal/repository"
	"serve-swimming-be/internal/service"
)

func main() {
	db, err := gorm.Open(sqlite.Open("swimming_event.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, nil)

	buku, err := svc.GetBukuAcara(1, "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total Events in Buku Acara: %d\n", len(buku))
	for _, ev := range buku {
		fmt.Printf("Event %d: %s (Dist: %s, Stroke: %s, Gender: %s, AgeGroup: %s, Sched: %s %s) | Heats: %d\n",
			ev.EventCode, ev.EventName, ev.Distance, ev.Stroke, ev.Gender, ev.AgeGroup, ev.ScheduleDate, ev.ScheduleTime, len(ev.Heats))
		for _, h := range ev.Heats {
			if !h.IsEmpty {
				fmt.Printf("   -> RegID: %d | Heat: %d Line: %d | %s (%s) | Seed: %s | Result: %s\n",
					h.RegistrationID, h.Heat, h.Line, h.Nama, h.Club, h.TimeSeed, h.Result)
			}
		}
	}
}
