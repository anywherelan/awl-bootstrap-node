package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// P2PDHTRoutingTableSize is not provided by libp2p Prometheus metrics
	// (go-libp2p-kad-dht is instrumented via OpenTelemetry and does not export
	// routing table size), so we expose it ourselves.
	P2PDHTRoutingTableSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemP2P,
		Name:      "dht_routing_table_size",
		Help:      "DHT routing table size.",
	})

	P2PBootstrapPeersConnected = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemP2P,
		Name:      "bootstrap_peers_connected",
		Help:      "Number of connected bootstrap peers.",
	})
)
