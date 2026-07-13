package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PeerstorePeersWithAddrs = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemPeerstore,
		Name:      "peers_with_addrs",
		Help:      "Number of peers in the peerstore that have known addresses.",
	})

	PeerstorePeersWithKeys = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemPeerstore,
		Name:      "peers_with_keys",
		Help:      "Number of peers in the peerstore that have known public keys.",
	})
)
