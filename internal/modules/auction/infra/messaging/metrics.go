package messaging

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	bidCommandsPublishedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auction_bid_commands_published_total",
		Help: "Total number of bid commands published to the command stream.",
	})

	bidCommandsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auction_bid_commands_processed_total",
		Help: "Total number of bid commands processed, labelled by result (ok|dup|dlq).",
	}, []string{"result"})

	eventsPublishedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auction_events_published_total",
		Help: "Total number of domain events published to the events stream.",
	})

	websocketBroadcastTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auction_ws_broadcast_total",
		Help: "Total number of events broadcast to WebSocket subscribers.",
	})

	bidCommandPublishDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "auction_bid_command_publish_duration_seconds",
		Help:    "Latency of publishing a bid command to the command stream.",
		Buckets: prometheus.DefBuckets,
	})
)

// IncWebsocketBroadcast increments the counter of events broadcast to WebSocket
// subscribers. It is exported so the websocket hub can record broadcasts without
// importing the Prometheus client directly.
func IncWebsocketBroadcast() {
	websocketBroadcastTotal.Inc()
}
