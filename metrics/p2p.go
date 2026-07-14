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

	// P2PBandwidthBytesTotal reports cumulative node-wide traffic. libp2p's
	// BandwidthCounter is not wired to Prometheus, so we snapshot it here.
	// Note: relayed traffic is also counted by libp2p_relaysvc_data_transferred_bytes_total.
	// These are cumulative counters exposed as a Gauge snapshot; use rate()/increase() in queries.
	P2PBandwidthBytesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemP2P,
		Name:      "bandwidth_bytes_total",
		Help:      "Total node bandwidth in bytes (snapshot of libp2p BandwidthCounter).",
	}, []string{"direction"})

	P2PBandwidthRateBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemP2P,
		Name:      "bandwidth_rate_bytes",
		Help:      "Current node bandwidth rate in bytes per second.",
	}, []string{"direction"})

	P2PBootstrapPeersConnected = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemP2P,
		Name:      "bootstrap_peers_connected",
		Help:      "Number of connected bootstrap peers.",
	})
)
