package services

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func AddOutboxEvent(tx *gorm.DB, entityType string, entityID uuid.UUID, op string, payload any) error {
	data, _ := json.Marshal(payload)

	event := models.Outbox{
		EntityType: entityType,
		EntityID:   entityID,
		Op:         op,
		Payload:    datatypes.JSON(data),
	}

	if err := tx.Create(&event).Error; err != nil {
		log.Printf("❌ Failed to create outbox event: %v", err)
		return err
	}

	return nil
}
