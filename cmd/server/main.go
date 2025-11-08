package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirdesai22/sync-service/internal/db"
	"github.com/sirdesai22/sync-service/internal/elastic"
	"github.com/sirdesai22/sync-service/internal/models"
	"github.com/sirdesai22/sync-service/internal/services"

	"github.com/sirdesai22/sync-service/internal/metrics"
	"github.com/sirdesai22/sync-service/internal/workers"
	"gorm.io/datatypes"
)

func main() {
	_ = godotenv.Load()

	pg := db.Connect()
	db.Migrate(pg)
	db.Seed(pg)

	metrics.Register()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// ✅ Simulate a user update
	// userID := uuid.MustParse("5a05617f-377e-4d42-832c-ce51fc0c58d8")
	// err := services.UpdateUser(pg, userID, map[string]any{"college": "IIT Delhi"})
	// time.Sleep(2 * time.Second)
	// if err != nil {
	// 	log.Fatalf("❌ user update failed: %v", err)
	// }

	// userID := uuid.MustParse("5a05617f-377e-4d42-832c-ce51fc0c58d8")
	// err := services.UpdateUser(pg, userID, map[string]any{"college": "IIT Suratkhal"})
	// time.Sleep(3 * time.Second)
	// if err != nil {
	// 	log.Fatalf("❌ user update failed: %v", err)
	// }

	es := elastic.Connect()
	worker := &workers.SyncWorker{DB: pg, ES: es}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Run(ctx)
	// go worker.RetryDLQ(ctx) // disabled: retry cycle handled manually via admin API

	// --- test: update user -> outbox event -> worker -> ES
	// var user models.User
	// if err := pg.Select("id").First(&user).Error; err != nil {
	// 	log.Fatalf("failed to fetch user id: %v", err)
	// }
	// _ = services.UpdateUser(pg, user.ID, map[string]any{"college": "IIT Tirupati"})

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/api/outbox", func(w http.ResponseWriter, r *http.Request) {
		var outboxes []models.Outbox
		pg.Order("id desc").Limit(100).Find(&outboxes)
		json.NewEncoder(w).Encode(outboxes)
	})
	mux.HandleFunc("/api/dlq", func(w http.ResponseWriter, r *http.Request) {
		var dlq []models.DLQ
		if err := pg.Order("id desc").Limit(100).Find(&dlq).Error; err != nil {
			log.Printf("DLQ query failed: %v", err)
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		log.Printf("DLQ query returned %d rows", len(dlq))
		json.NewEncoder(w).Encode(dlq)
	})
	mux.HandleFunc("/api/retry/", func(rw http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/api/retry/"):]
		var dlqEntry models.DLQ
		if err := pg.First(&dlqEntry, "id = ?", id).Error; err != nil {
			http.Error(rw, "not found", http.StatusNotFound)
			return
		}

		var ob models.Outbox
		if err := pg.First(&ob, "id = ?", dlqEntry.OutboxID).Error; err != nil {
			http.Error(rw, "outbox missing", http.StatusInternalServerError)
			return
		}

		bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Client: es, Index: "", FlushBytes: 5 << 20, NumWorkers: 2,
		})
		if err := worker.ApplyEvent(ctx, bi, ob); err != nil {
			http.Error(rw, "retry failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		now := time.Now()
		if err := pg.Model(&models.DLQ{}).
			Where("id = ?", id).
			Updates(map[string]any{"resolved": true, "retried_at": &now}).Error; err != nil {
			http.Error(rw, "update failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(rw).Encode(map[string]string{"status": "retried"})
	})

	mux.HandleFunc("/api/add-user", func(w http.ResponseWriter, r *http.Request) {
		skills, _ := json.Marshal([]string{"Go", "React"})
		u := models.User{
			Username: fmt.Sprintf("user_%d", time.Now().Unix()%1000),
			Email:    fmt.Sprintf("demo%d@example.com", time.Now().Unix()%1000),
			Skills:   datatypes.JSON(skills),
			College:  "PESU",
		}
		if err := pg.Create(&u).Error; err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// enqueue outbox event
		_ = services.AddOutboxEvent(pg, "user", u.ID, "UPSERT", u)
		json.NewEncoder(w).Encode(map[string]any{"status": "created", "id": u.ID})
	})

	mux.HandleFunc("/api/update-user", func(w http.ResponseWriter, r *http.Request) {
		// pick a random user to modify
		var u models.User
		if err := pg.Order("random()").First(&u).Error; err != nil {
			http.Error(w, "no users found", 404)
			return
		}
		u.College = "NIT Trichy"
		if err := pg.Save(&u).Error; err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = services.AddOutboxEvent(pg, "user", u.ID, "UPSERT", u)
		json.NewEncoder(w).Encode(map[string]any{"status": "updated", "id": u.ID})
	})

	log.Println("🧭 Admin API running on :8080")
	if err := http.ListenAndServe(":8080", corsMiddleware.Handler(mux)); err != nil {
		log.Fatalf("admin API listener failed: %v", err)
	}
	select {}

	// log.Println("Worker running. Give it a moment to sync…")
	// time.Sleep(3 * time.Second)

	// log.Println("🌍 Sync service initialized and DB ready.")
	// log.Println(pg, es)
}
