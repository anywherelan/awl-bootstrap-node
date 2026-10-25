package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var connectedPeersDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, subsystemP2P, "peers_connected"),
	"Connected peers by the client they run, taken from their user agent.",
	[]string{labelNameApp, labelNameVersion, labelNameOS, labelNameArch}, nil,
)

func init() {
	// CollectorFunc describes itself by collecting, and nothing is collected
	// before SetHost: the registry takes it as an unchecked collector.
	prometheus.MustRegister(prometheus.CollectorFunc(collectConnectedPeers))
}

// collectConnectedPeers counts the connected peers on every scrape, so a peer
// that disconnected is gone from the next scrape without any cleanup.
func collectConnectedPeers(ch chan<- prometheus.Metric) {
	h := scrapeHost.Load()
	if h == nil {
		return
	}

	peerstore := (*h).Peerstore()
	counts := make(map[clientLabels]int)
	for _, id := range (*h).Network().Peers() {
		counts[parseUserAgent(peerUserAgent(peerstore, id))]++
	}

	for labels, count := range counts {
		ch <- prometheus.MustNewConstMetric(connectedPeersDesc, prometheus.GaugeValue, float64(count),
			labels.app, labels.version, labels.os, labels.arch)
	}
}
