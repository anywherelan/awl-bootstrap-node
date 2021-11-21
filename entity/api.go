package entity

import (
	"github.com/anywherelan/awl/p2p"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

// Requests

// Responses
type (
	P2pDebugInfo struct {
		General     GeneralDebugInfo
		DHT         DhtDebugInfo
		Connections ConnectionsDebugInfo
		Peerstore   PeerstoreDebugInfo
		Bandwidth   BandwidthDebugInfo
	}

	GeneralDebugInfo struct {
		Version string
		Uptime  string
	}
	DhtDebugInfo struct {
		RoutingTableSize int
		RoutingTable     []kbucket.PeerInfo
		Reachability     string
		ListenAddress    []string
		ObservedAddrs    []string
	}
	BootstrapPeerDebugInfo struct {
		Error       string
		Connections []string
	}
	ConnectionsDebugInfo struct {
		ConnectedPeersCount  int
		OpenConnectionsCount int
		OpenStreamsCount     int64
		TotalStreamsInbound  int64
		TotalStreamsOutbound int64
		LastTrimAgo          string
	}
	BandwidthDebugInfo struct {
		Total      BandwidthInfo
		ByProtocol map[string]BandwidthInfo
	}

	PeerstoreDebugInfo struct {
		PeersWithAddrsCount int
		PeersWithKeysCount  int
		Peers               map[string]Peer
	}
	Peer struct {
		IsConnected     bool
		UserAgent       string
		Bandwidth       BandwidthInfo
		ConnectionsInfo []p2p.ConnectionInfo
		LatencyMs       int64
		Protocols       []string
		PeerstoreAddrs  []string
	}

	BandwidthInfo struct {
		TotalIn  string
		TotalOut string
		RateIn   string
		RateOut  string
	}
)
