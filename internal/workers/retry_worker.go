package workers

import (
	"context"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/sirdesai22/sync-service/internal/models"
	"github.com/sirdesai22/sync-service/internal/metrics"
)

func (w *SyncWorker) RetryDLQ(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var dlqs []models.DLQ
			if err := w.DB.Where("resolved = false").Limit(50).Find(&dlqs).Error; err != nil {
				log.Printf("DLQ fetch error: %v", err)
				continue
			}
			for _, d := range dlqs {
				log.Printf("♻️ Retrying DLQ id=%d entity=%s op=%s", d.ID, d.EntityType, d.Op)
				ob := models.Outbox{
					ID:         d.OutboxID,
					EntityType: d.EntityType,
					Op:         d.Op,
				}
				// reapply event
				bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
					Client: w.ES, Index: "", FlushBytes: 5 << 20, NumWorkers: 2,
				})
				if err := w.applyEvent(ctx, bi, ob); err == nil {
					now := time.Now()
					w.DB.Model(&models.DLQ{}).Where("id = ?", d.ID).Updates(map[string]any{
						"resolved":  true,
						"retried_at": &now,
					})
					metrics.ProcessedEvents.Inc()
					log.Printf("✅ DLQ id=%d resolved", d.ID)
				}
			}
		}
	}
}
