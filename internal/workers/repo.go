// internal/workers/repo.go
// this file is used to fetch the outbox events and put them in the DLQ
package workers

import (
	"context"
	"log"
	"time"
	"gorm.io/gorm"
	"github.com/sirdesai22/sync-service/internal/models"
	"github.com/sirdesai22/sync-service/internal/metrics"
)

type OutboxBatch struct{ Events []models.Outbox }

func FetchOutboxBatch(ctx context.Context, db *gorm.DB, limit int) (OutboxBatch, error) {
	var evts []models.Outbox
	// FOR UPDATE SKIP LOCKED to allow multiple workers later
	tx := db.WithContext(ctx).Raw(`
		WITH cte AS (
		  SELECT * FROM outboxes
		  WHERE processed = false
		  ORDER BY id ASC
		  LIMIT ?
		  FOR UPDATE SKIP LOCKED
		)
		UPDATE outboxes SET processed = true
		FROM cte
		WHERE outboxes.id = cte.id
		RETURNING cte.*`, limit).Scan(&evts)
	return OutboxBatch{Events: evts}, tx.Error
}

// PutDLQ inserts a failed outbox event into the DLQ table.
func PutDLQ(db *gorm.DB, ob models.Outbox, msg string) {
    metrics.DLQEvents.Inc()
    dlq := models.DLQ{
        OutboxID:   ob.ID,
        EntityType: ob.EntityType,
        EntityID:   ob.EntityID.String(),
        Op:         ob.Op,
        ErrorMsg:   msg,
        Payload:    ob.Payload,
        CreatedAt:  time.Now(),
        Resolved:   false,
    }
    if err := db.Create(&dlq).Error; err != nil {
        log.Printf("❌ Failed to insert into DLQ: %v", err)
    } else {
        log.Printf("💀 DLQ record created for outbox_id=%d", ob.ID)
    }
}
