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
		log.Fatal(err)
	}

	var parts []map[string]interface{}
	db.Table("participants").Order("id asc").Find(&parts)
	fmt.Printf("=== ALL PARTICIPANTS IN DB (%d) ===\n", len(parts))
	for _, p := range parts {
		fmt.Printf("ID: %-3v | Name: %-25v | Club: %-25v | Gender: %-6v | DOB: %-12v | AgeGroup: %v\n",
			p["id"], p["name"], p["club"], p["gender"], p["birth_date"], p["age_group"])
	}
}
