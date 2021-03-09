package entity

// Requests

// Responses
type (
	P2pDebugInfo struct {
		General     GeneralDebugInfo
		DHT         DhtDebugInfo
		Connections ConnectionsDebugInfo
		Bandwidth   BandwidthDebugInfo
	}

	GeneralDebugInfo struct {
		Uptime string
	}
	DhtDebugInfo struct {
		RoutingTableSize    int
		Reachability        string
		ListenAddress       []string
		PeersWithAddrsCount int
		ObservedAddrs       []string
		BootstrapPeers      map[string]BootstrapPeerDebugInfo
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
		//ByPeer     map[peer.ID]metrics.Stats
	}
	BandwidthInfo struct {
		TotalIn  string
		TotalOut string
		RateIn   string
		RateOut  string
	}
)
