// internal/workers/sync_worker.go
package workers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	// "github.com/google/uuid"
	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/sirdesai22/sync-service/internal/elastic"
	"github.com/sirdesai22/sync-service/internal/metrics"
	"github.com/sirdesai22/sync-service/internal/models"
	"gorm.io/gorm"
)

type SyncWorker struct {
	DB *gorm.DB
	ES *es.Client
}

func (w *SyncWorker) Run(ctx context.Context) {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        w.ES,
		NumWorkers:    1,
		FlushInterval: 2 * time.Second,
	})
	if err != nil {
		log.Fatalf("Bulk indexer init failed: %v", err)
	}

	// ✅ Close only once, when the worker exits — not after every batch
	defer func() {
		log.Println("Closing BulkIndexer …")
		if err := bi.Close(ctx); err != nil {
			log.Printf("BulkIndexer close error: %v", err)
		}
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Sync worker shutting down")
			return
		case <-ticker.C:
			w.processOnce(ctx, bi) // pass the same BulkIndexer each tick
		}
	}
}

func (w *SyncWorker) processOnce(ctx context.Context, bi esutil.BulkIndexer) error {
	batch, err := FetchOutboxBatch(ctx, w.DB, 200)
	if err != nil {
		return err
	}
	if len(batch.Events) == 0 {
		return nil
	}

	for _, e := range batch.Events {
		if err := w.applyEvent(ctx, bi, e); err != nil {
			// Put back to DLQ (already marked processed to avoid infinite loop)
			metrics.FailedEvents.Inc()
			PutDLQ(w.DB, e, err.Error())
			log.Printf("DLQ outbox_id=%d: %v", e.ID, err)
			continue
		}
		metrics.ProcessedEvents.Inc()
	}

	stats := bi.Stats()
	log.Printf("bulk ok=%d failed=%d", stats.NumFlushed, stats.NumFailed)
	return nil
}

func (w *SyncWorker) ApplyEvent(ctx context.Context, bi esutil.BulkIndexer, e models.Outbox) error {
	return w.applyEvent(ctx, bi, e)
}

func (w *SyncWorker) applyEvent(ctx context.Context, bi esutil.BulkIndexer, e models.Outbox) error {
	switch e.EntityType {
	case "user":
		var u models.User
		if e.Op == "DELETE" {
			return w.add(bi, elastic.IdxUsers, e.EntityID.String(), e.ID, "delete", nil)
		}
		if err := w.DB.First(&u, "id = ?", e.EntityID).Error; err != nil {
			return err
		}
		doc, err := elastic.BuildUserDoc(u)
		if err != nil {
			return err
		}
		return w.add(bi, elastic.IdxUsers, e.EntityID.String(), e.ID, "index", doc)

	case "hackathon":
		var h models.Hackathon
		if e.Op == "DELETE" {
			return w.add(bi, elastic.IdxHackathons, e.EntityID.String(), e.ID, "delete", nil)
		}
		if err := w.DB.First(&h, "id = ?", e.EntityID).Error; err != nil {
			return err
		}
		doc, err := elastic.BuildHackathonDoc(h)
		if err != nil {
			return err
		}
		return w.add(bi, elastic.IdxHackathons, e.EntityID.String(), e.ID, "index", doc)

	case "project":
		var p models.Project
		if e.Op == "DELETE" {
			return w.add(bi, elastic.IdxProjects, e.EntityID.String(), e.ID, "delete", nil)
		}
		if err := w.DB.First(&p, "id = ?", e.EntityID).Error; err != nil {
			return err
		}
		doc, err := elastic.BuildProjectDoc(p)
		if err != nil {
			return err
		}
		return w.add(bi, elastic.IdxProjects, e.EntityID.String(), e.ID, "index", doc)
	}
	return fmt.Errorf("unknown entity_type=%s", e.EntityType)
}

func (w *SyncWorker) add(bi esutil.BulkIndexer, index, docID string, outboxID int64, action string, body []byte) error {
	item := esutil.BulkIndexerItem{
		Action:     action,
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(body),
		OnSuccess: func(_ context.Context, _ esutil.BulkIndexerItem, _ esutil.BulkIndexerResponseItem) {
			log.Printf("✅ synced %s id=%s", index, docID)
		},
		OnFailure: func(_ context.Context, it esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
			log.Printf("💀 OnFailure fired for %s/%s", it.Index, it.DocumentID)

			msg := ""
			if err != nil {
				msg = err.Error()
			} else if res.Error.Reason != "" {
				msg = fmt.Sprintf("%s: %s", res.Error.Type, res.Error.Reason)
			} else {
				msg = fmt.Sprintf("status=%d", res.Status)
			}

			dbErr := w.DB.Session(&gorm.Session{}).Create(&models.DLQ{
				OutboxID:   outboxID,
				EntityType: indexToEntity(index),
				EntityID:   docID,
				Op:         action,
				ErrorMsg:   msg,
				CreatedAt:  time.Now(),
				Resolved:   false,
			}).Error
			if dbErr != nil {
				log.Printf("❌ DLQ insert failed: %v", dbErr)
			} else {
				log.Printf("💾 DLQ row added for outbox=%d", outboxID)
			}
		},
	}

	log.Printf("💾 Adding item to Elasticsearch: %s %s %s", index, action, docID)
	return bi.Add(context.Background(), item)
}

func indexToEntity(index string) string {
	switch index {
	case elastic.IdxUsers:
		return "user"
	case elastic.IdxProjects:
		return "project"
	case elastic.IdxHackathons:
		return "hackathon"
	default:
		return "unknown"
	}
}
