// Package metrics provides Prometheus metrics specific to awl-bootstrap-node.
//
// Most of the useful telemetry for a bootstrap node (relay service, autonat,
// swarm, resource manager, identify, holepunch, ...) comes from libp2p's own
// built-in Prometheus metrics, enabled via libp2p.PrometheusRegisterer. This
// package only adds what libp2p does not expose: DHT routing table size, total
// node bandwidth, node info/uptime, peerstore size, connected peers and DHT
// requests by client, and relayed connection durations with buckets that fit
// the relay limits.
package metrics

import (
	"context"
	"sync/atomic"
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
}

// scrapeHost is the host that the collectors read on every scrape.
var scrapeHost atomic.Pointer[host.Host]

// SetHost sets the host for the metrics computed on scrape: connected peers
// and peerstore size. They are absent until it is set.
func SetHost(h host.Host) {
	scrapeHost.Store(&h)
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

	NodeUptimeSeconds.Set(time.Since(startTime).Seconds())
}
