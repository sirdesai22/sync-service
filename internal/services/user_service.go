package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/gorm"
)

// UpdateUser updates a user and creates outbox entries for:
// 1️⃣ The user itself (UPSERT)
// 2️⃣ All related projects (so they get reindexed)
func UpdateUser(db *gorm.DB, id uuid.UUID, updates map[string]any) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// --- Step 1: Update the user record ---
		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		// --- Step 2: Record outbox event for user itself ---
		var user models.User
		if err := tx.First(&user, "id = ?", id).Error; err != nil {
			return err
		}
		if err := AddOutboxEvent(tx, "user", user.ID, "UPSERT", user); err != nil {
			return err
		}

		// --- Step 3: Find all related projects ---
		var projects []models.Project
		if err := tx.Where("owner_id = ?", id).Find(&projects).Error; err != nil {
			return err
		}
		if len(projects) == 0 {
			log.Println("ℹ️ No projects found for this user; skipping cascade reindex.")
			return nil
		}

		// --- Step 4: Enqueue reindex events for each project ---
		projectIDs := make([]uuid.UUID, 0, len(projects))
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}
		if err := AddBatchOutboxEvents(tx, "project", "UPSERT", projectIDs); err != nil {
			return err
		}

		log.Printf("🔁 Cascade reindex triggered for %d projects of user %s", len(projectIDs), user.Username)
		return nil
	})
}
