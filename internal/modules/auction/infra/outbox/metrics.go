package outbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var eventsRelayedTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "auction_outbox_events_relayed_total",
	Help: "Total number of outbox events relayed to the events stream.",
})
