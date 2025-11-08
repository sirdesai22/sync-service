package db

import (
	"log"

	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Hackathon{},
		&models.Project{},
		&models.Outbox{},
		&models.DLQ{},
	)
	if err != nil {
		log.Fatalf("❌ migration failed: %v", err)
	}
	log.Println("✅ database migrated successfully")
}
