package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/gorm"
)

func UpdateUser(db *gorm.DB, id uuid.UUID, updates map[string]any) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Update the user
		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		// Create outbox entry for sync
		user := models.User{}
		if err := tx.First(&user, "id = ?", id).Error; err != nil {
			return err
		}

		if err := AddOutboxEvent(tx, "user", user.ID, "UPSERT", user); err != nil {
			return err
		}

		log.Printf("📤 Outbox event recorded for user %s", user.Username)
		return nil
	})
}
