package main

import (
	"log"
	"github.com/joho/godotenv"
	"github.com/sirdesai22/sync-service/internal/db"
	"github.com/sirdesai22/sync-service/internal/elastic"

	"github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/services"
)

func main() {
	_ = godotenv.Load()

	pg := db.Connect()
	db.Migrate(pg)
	db.Seed(pg)

	// ✅ Simulate a user update
	userID := uuid.MustParse("1a871772-e628-45cd-adef-e26a14303d29")
	err := services.UpdateUser(pg, userID, map[string]any{"college": "NIT Suratkhal"})
	if err != nil {
		log.Fatalf("❌ user update failed: %v", err)
	}


	es := elastic.Connect()

	log.Println("🌍 Sync service initialized and DB ready.")
	log.Println(pg, es)
}