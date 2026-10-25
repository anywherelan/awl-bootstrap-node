package metrics

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// The peerstore size is counted on scrape: in the badger-backed peerstore each
// count is a scan over all stored versions of the keys, too expensive for the
// background updater.
var (
	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemPeerstore,
		Name:      "peers_with_addrs",
		Help:      "Number of peers in the peerstore that have known addresses.",
	}, func() float64 { return countPeers(peerstore.Peerstore.PeersWithAddrs) })

	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemPeerstore,
		Name:      "peers_with_keys",
		Help:      "Number of peers in the peerstore that have known public keys.",
	}, func() float64 { return countPeers(peerstore.Peerstore.PeersWithKeys) })
)

func countPeers(peers func(peerstore.Peerstore) peer.IDSlice) float64 {
	h := scrapeHost.Load()
	if h == nil {
		return 0
	}
	return float64(len(peers((*h).Peerstore())))
}
