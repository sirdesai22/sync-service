package services

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func getPriority(entityType string) int {
	switch entityType {
	case "user":
		return 0 // Highest priority user
	case "project":
		return 1 // Medium priority project
	case "hackathon":
		return 2 // Lower priority hackathon
	default:
		return 99 // Unknown entities last
	}
}

// AddOutboxEvent inserts one event into the outbox (same as before)
func AddOutboxEvent(tx *gorm.DB, entityType string, entityID uuid.UUID, op string, payload any) error {
	data, _ := json.Marshal(payload)

	event := models.Outbox{
		EntityType: entityType,
		EntityID:   entityID,
		Op:         op,
		Priority: getPriority(entityType),
		Payload:    datatypes.JSON(data),
	}

	if err := tx.Create(&event).Error; err != nil {
		log.Printf("❌ Failed to create outbox event: %v", err)
		return err
	}
	return nil
}

// AddBatchOutboxEvents inserts multiple events efficiently.
// Used for cascading updates (e.g., reindex all projects for a user).
func AddBatchOutboxEvents(tx *gorm.DB, entityType string, op string, ids []uuid.UUID) error {
	for _, id := range ids {
		event := models.Outbox{
			EntityType: entityType,
			EntityID:   id,
			Op:         op,
			Priority: getPriority(entityType),
		}
		if err := tx.Create(&event).Error; err != nil {
			log.Printf("❌ Failed to insert batch outbox for %s: %v", entityType, err)
			return err
		}
	}
	log.Printf("📦 %d outbox events created for %s", len(ids), entityType)
	return nil
}
