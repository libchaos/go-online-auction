package outbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var paymentEventsRelayedTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "payment_events_relayed_total",
	Help: "Total number of payment outbox events relayed to NATS JetStream",
})
