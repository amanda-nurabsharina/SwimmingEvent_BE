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

	fmt.Printf("=== SIMULATING CERTIFICATES (Total Events: %d) ===\n", len(buku))
	count := 0
	for _, ev := range buku {
		for _, h := range ev.Heats {
			if !h.IsEmpty && h.Nama != "" {
				count++
				fmt.Printf("[%2d] RegID: %d | EvCode: %d | %s | %s | TimeResult: %s | TimeSeed: %s\n",
					count, h.RegistrationID, ev.EventCode, h.Nama, ev.EventName, h.Result, h.TimeSeed)
			}
		}
	}
	fmt.Printf("Total Certificates: %d\n", count)
}
