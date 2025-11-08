## Sync Service

A Go-based sync service that keeps PostgreSQL as the source of truth in sync with Elasticsearch, complete with observability, a DLQ (dead-letter queue) workflow, and a lightweight dashboard for manual operations.

---

## Architecture Overview
![Architecture Diagram](image.png)

### Key Components

- **Outbox pattern** (`internal/services`): database mutations enqueue events in `outboxes`.
- **Sync worker** (`internal/workers/sync_worker.go`): polls unprocessed events and indexes documents into Elasticsearch.
- **Dead-letter queue** (`internal/models/dlq.go`): records that failed to sync for manual inspection and retry.
- **Admin API** (`cmd/server/main.go`): exposes metrics, latest outbox rows, DLQ items, and helper actions.
- **Metrics** (`internal/metrics`): Prometheus counters for processed, failed, and DLQ events.
- **Dashboard** (`sync-dashboard/`): React + SWR app that consumes the Admin API for monitoring and manual operations.

Automatic DLQ retries are disabled by default; retries are triggered manually through the Admin API or dashboard.

---

## Getting Started

### Requirements

- Go 1.25+
- Node.js 20+ (for the dashboard)
- Docker & Docker Compose (for PostgreSQL and Elasticsearch)

### 1. Start infrastructure

```bash
docker compose up -d
```

This launches:

- PostgreSQL (`postgres://dev:dev@localhost:5431/syncdb`)
- Elasticsearch on `localhost:9200`
- Kibana on `localhost:5601`

### 2. Configure environment

Create a `.env` file or export the DSN before running the server:

```bash
export POSTGRES_DSN="host=localhost port=5431 user=dev password=dev dbname=syncdb sslmode=disable"
```

### 3. Run the sync service

```bash
go run ./cmd/server
```

On boot the service:

- Connects to Postgres & Elasticsearch
- Runs migrations (`internal/db/migrate.go`)
- Seeds sample data if the database is empty
- Starts the background sync worker
- Exposes the Admin API on `:8080`

### 4. (Optional) Launch the dashboard

```bash
cd sync-dashboard
npm install
npm run dev
```

Visit `http://localhost:5173` to view outbox & DLQ tables and trigger helper actions.

---

## Admin API

| Endpoint | Description |
| --- | --- |
| `GET /metrics` | Prometheus metrics (`sync_processed_total`, `sync_failed_total`, `sync_dlq_total`) |
| `GET /api/outbox` | Latest 100 outbox events (ordered by `id desc`) |
| `GET /api/dlq` | Latest 100 DLQ entries |
| `GET /api/retry/{id}` | Manually retry a DLQ row (re-runs the event through the worker) |
| `POST /api/add-user` | Creates a demo user and enqueues an outbox event |
| `POST /api/update-user` | Updates a random user, demonstrating cascading outbox writes |

> **Note:** `/api/retry/{id}` reuses the standard worker logic; if the retry fails again it will remain unresolved in the DLQ.

---

## Operational Notes

- **Manual DLQ handling:** Automatic retry loops are intentionally disabled (`RetryDLQ` is not started). Use `/api/retry/{id}` or the dashboard button to retry failed events.
- **Bulk indexer lifecycle:** The sync worker keeps a single bulk indexer instance alive for the lifetime of the worker, ensuring efficient flush behaviour.
- **Prometheus/Kibana:** Exposed ports (`:8080`, `:9200`, `:5601`) make it easy to plug in monitoring tools or view indexed documents.
- **Seeding:** Initial sample data (user, hackathon, project) is inserted only when the database is empty.

---

## Project Structure

```
cmd/server/         # main entrypoint & admin API
internal/db/        # connection, migrations, seed data
internal/services/  # domain operations (outbox writes, user updates)
internal/workers/   # sync worker, DLQ repo & retry helpers
internal/elastic/   # client setup & document builders
internal/metrics/   # Prometheus instrumentation
sync-dashboard/     # React dashboard for monitoring/manual actions
```

---
