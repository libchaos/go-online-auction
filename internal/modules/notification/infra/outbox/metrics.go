package outbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var notificationEventsRelayedTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "notification_events_relayed_total",
	Help: "Total number of notification outbox events relayed to NATS JetStream",
})
