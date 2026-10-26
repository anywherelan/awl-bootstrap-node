package metrics

import (
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

// BandwidthStats is the source of the traffic metric.
// It is satisfied by *github.com/anywherelan/awl/p2p.P2p.
type BandwidthStats interface {
	NetworkStatsByProtocol() map[protocol.ID]metrics.Stats
}

var bandwidthDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, subsystemP2P, "bandwidth_bytes_total"),
	"Traffic of libp2p streams in bytes by protocol, without transport overhead. "+
		"Relayed data is counted twice: in on one relay stream and out on the other.",
	[]string{"protocol", "direction"}, nil,
)

// scrapeBandwidth is the traffic source that the collector reads on every scrape.
var scrapeBandwidth atomic.Pointer[BandwidthStats]

func init() {
	// An unchecked collector, as in peers.go.
	prometheus.MustRegister(prometheus.CollectorFunc(collectBandwidth))
}

// SetBandwidthStats sets the source of the traffic metric. The metric is
// absent until it is set.
func SetBandwidthStats(s BandwidthStats) {
	scrapeBandwidth.Store(&s)
}

// collectBandwidth reports the totals of the libp2p bandwidth counter as
// counters: nothing resets or trims the counter, so they only grow.
func collectBandwidth(ch chan<- prometheus.Metric) {
	s := scrapeBandwidth.Load()
	if s == nil {
		return
	}

	for proto, stats := range (*s).NetworkStatsByProtocol() {
		// Traffic of streams before their protocol is negotiated. An empty
		// label value is the same as no label.
		label := string(proto)
		if label == "" {
			label = labelUnknown
		}
		ch <- prometheus.MustNewConstMetric(bandwidthDesc, prometheus.CounterValue, float64(stats.TotalIn), label, "in")
		ch <- prometheus.MustNewConstMetric(bandwidthDesc, prometheus.CounterValue, float64(stats.TotalOut), label, "out")
	}
}
