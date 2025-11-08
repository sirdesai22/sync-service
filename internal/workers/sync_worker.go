// internal/workers/sync_worker.go
package workers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
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
	if err := elastic.EnsureIndexes(ctx, w.ES); err != nil {
		log.Fatalf("ensure indexes: %v", err)
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processOnce(ctx); err != nil {
				log.Printf("worker error: %v", err)
			}
		}
	}
}

func (w *SyncWorker) processOnce(ctx context.Context) error {
	batch, err := FetchOutboxBatch(ctx, w.DB, 200)
	if err != nil {
		return err
	}
	if len(batch.Events) == 0 {
		return nil
	}

	bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: w.ES, Index: "", FlushBytes: 5 << 20, NumWorkers: 2,
	})

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

	if err := bi.Close(ctx); err != nil {
		return err
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
		DocumentID: docID,
		Index:      index,
		Body:       nil,
		OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
			log.Printf("✅ synced %s id=%s", index, docID)
		},
		OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
			msg := ""
			switch {
			case err != nil:
				msg = err.Error()
			case res.Error.Reason != "":
				msg = fmt.Sprintf("%s: %s", res.Error.Type, res.Error.Reason)
			default:
				msg = fmt.Sprintf("status=%d failed to index", res.Status)
			}

			// ✅ Write failed event to DLQ
			ob := models.Outbox{
				ID:         outboxID,
				EntityType: indexToEntity(index),
				EntityID:   uuid.MustParse(docID),
				Op:         action,
				Payload:    nil,
			}
			PutDLQ(w.DB, ob, msg)
			log.Printf("💀 DLQ created for outbox_id=%d index=%s id=%s reason=%s", outboxID, index, docID, msg)
		},
	}

	if len(body) > 0 {
		item.Body = bytes.NewReader(body)
	}
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
