// Package metrics provides Prometheus metrics specific to awl-bootstrap-node.
//
// Most of the useful telemetry for a bootstrap node (relay service, autonat,
// swarm, resource manager, identify, holepunch, ...) comes from libp2p's own
// built-in Prometheus metrics, enabled via libp2p.PrometheusRegisterer. This
// package only adds the few gauges that libp2p does not expose: DHT routing
// table size, total node bandwidth and node info/uptime.
package metrics

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/metrics"
)

const (
	namespace = "awl_bootstrap"

	subsystemNode      = "node"
	subsystemP2P       = "p2p"
	subsystemPeerstore = "peerstore"
)

// P2pMetrics is an interface for getting p2p stats used by the background updater.
// It is satisfied by *github.com/anywherelan/awl/p2p.P2p.
type P2pMetrics interface {
	RoutingTableSize() int
	NetworkStats() metrics.Stats
	BootstrapPeersStats() (total int, connected int)
	Host() host.Host
}

// StartBackgroundUpdater periodically updates gauge-type metrics from their data sources.
func StartBackgroundUpdater(ctx context.Context, p2pMetrics P2pMetrics) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	updateGauges(p2pMetrics, startTime)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updateGauges(p2pMetrics, startTime)
		}
	}
}

func updateGauges(p2pMetrics P2pMetrics, startTime time.Time) {
	P2PDHTRoutingTableSize.Set(float64(p2pMetrics.RoutingTableSize()))

	_, bootstrapConnected := p2pMetrics.BootstrapPeersStats()
	P2PBootstrapPeersConnected.Set(float64(bootstrapConnected))

	stats := p2pMetrics.NetworkStats()
	P2PBandwidthBytesTotal.WithLabelValues("in").Set(float64(stats.TotalIn))
	P2PBandwidthBytesTotal.WithLabelValues("out").Set(float64(stats.TotalOut))
	P2PBandwidthRateBytes.WithLabelValues("in").Set(stats.RateIn)
	P2PBandwidthRateBytes.WithLabelValues("out").Set(stats.RateOut)

	peerstore := p2pMetrics.Host().Peerstore()
	PeerstorePeersWithAddrs.Set(float64(len(peerstore.PeersWithAddrs())))
	PeerstorePeersWithKeys.Set(float64(len(peerstore.PeersWithKeys())))

	NodeUptimeSeconds.Set(time.Since(startTime).Seconds())
}
