package db

import (
	"encoding/json"
	"log"
	"time"

	// "github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("🌱 Data already exists, skipping seed.")
		return
	}

	// Wrap in a transaction for atomicity
	db.Transaction(func(tx *gorm.DB) error {
		// 1️⃣ Create user
		skills, _ := json.Marshal([]string{"Go", "React", "AI"})
		user := models.User{
			Username: "prathamesh",
			Email:    "me@example.com",
			Skills:   datatypes.JSON(skills),
			College:  "PESU",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 2️⃣ Create hackathon
		tracks, _ := json.Marshal([]string{"AI", "Web"})
		hackathon := models.Hackathon{
			Name:     "DevFest",
			Location: "Bengaluru",
			Tracks:   datatypes.JSON(tracks),
			StartAt:  time.Now(),
			EndAt:    time.Now().Add(48 * time.Hour),
		}
		if err := tx.Create(&hackathon).Error; err != nil {
			return err
		}

		// 3️⃣ Create project (now we have real IDs)
		project := models.Project{
			Name:         "Voice for All",
			Description:  "AI assistant for mute people",
			HackathonID:  hackathon.ID,
			OwnerID:      user.ID,
			TeamMembers:  datatypes.JSON([]byte("[]")),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := tx.Create(&project).Error; err != nil {
			return err
		}

		log.Println("🌱 Sample data inserted successfully.")
		return nil
	})
}
