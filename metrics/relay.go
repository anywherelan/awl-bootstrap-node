package metrics

import (
	"time"

	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// P2PRelayCircuitDurationSeconds repeats libp2p_relaysvc_connection_duration_seconds
// with buckets that fit relayed connections: the libp2p one has the default
// buckets, the largest is 10 s, so it can't show circuits cut by the relay
// limit. The buckets are around 2 and 10 minutes, the current limit and the
// next one.
//
// libp2p reports a circuit as closed also when the relay could not reach the
// destination, so failed attempts land in the first bucket. To get shares of
// established circuits, divide by
// libp2p_relaysvc_connection_request_response_status_total{status="ok"}.
var P2PRelayCircuitDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystemP2P,
	Name:      "relay_circuit_duration_seconds",
	Help:      "How long relayed connections lasted. Attempts that failed to reach the destination are counted too, with a duration under 1 s.",
	Buckets:   []float64{1, 10, 30, 60, 110, 125, 300, 590, 610, 1200, 3600},
})

// RelayMetricsTracer returns the relay metrics tracer of libp2p registered
// with reg, which also observes P2PRelayCircuitDurationSeconds. Pass it with
// relay.WithMetricsTracer: it replaces the tracer libp2p sets up itself.
func RelayMetricsTracer(reg prometheus.Registerer) relay.MetricsTracer {
	return relayMetricsTracer{MetricsTracer: relay.NewMetricsTracer(relay.WithRegisterer(reg))}
}

type relayMetricsTracer struct {
	relay.MetricsTracer
}

func (t relayMetricsTracer) ConnectionClosed(d time.Duration) {
	t.MetricsTracer.ConnectionClosed(d)
	P2PRelayCircuitDurationSeconds.Observe(d.Seconds())
}
