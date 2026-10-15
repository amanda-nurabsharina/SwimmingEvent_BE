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
		Cid  int
		Name string
		Type string
	}
	db.Raw("PRAGMA table_info(registrations)").Scan(&regCols)
	fmt.Println("=== REGISTRATIONS COLUMNS ===")
	for _, c := range regCols {
		fmt.Printf("Col: %s (%s)\n", c.Name, c.Type)
	}
}
