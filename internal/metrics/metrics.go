package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ProcessedEvents = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "sync_processed_total", Help: "Total processed outbox events"},
	)
	FailedEvents = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "sync_failed_total", Help: "Total failed outbox events"},
	)
	DLQEvents = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "sync_dlq_total", Help: "Total events inserted into DLQ"},
	)
)

func Register() {
	prometheus.MustRegister(ProcessedEvents, FailedEvents, DLQEvents)
}
