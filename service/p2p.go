package service

import (
	"time"

	"github.com/anywherelan/awl-bootstrap-node/config"
	"github.com/anywherelan/awl-bootstrap-node/p2p"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	protectedPeerTag = "known"
)

type P2pService struct {
	p2pServer *p2p.P2p
	conf      *config.Config
	logger    *log.ZapEventLogger
	startedAt time.Time
}

func NewP2p(server *p2p.P2p, conf *config.Config) *P2pService {
	p := &P2pService{
		p2pServer: server,
		conf:      conf,
		logger:    log.Logger("awl/service/p2p"),
		startedAt: time.Now(),
	}

	// Protect friendly peers from disconnecting
	conf.RLock()

	for _, peerAddr := range conf.GetBootstrapPeers() {
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			continue
		}
		p.ProtectPeer(peerInfo.ID)
	}
	conf.RUnlock()

	return p
}

func (s *P2pService) IsConnected(peerID peer.ID) bool {
	return s.p2pServer.IsConnected(peerID)
}

func (s *P2pService) ProtectPeer(id peer.ID) {
	s.p2pServer.ChangeProtectedStatus(id, protectedPeerTag, true)
}

func (s *P2pService) UnprotectPeer(id peer.ID) {
	s.p2pServer.ChangeProtectedStatus(id, protectedPeerTag, false)
}

func (s *P2pService) PeerAddresses(peerID peer.ID) []string {
	conns := s.p2pServer.ConnsToPeer(peerID)
	addrs := make([]string, 0, len(conns))
	for _, conn := range conns {
		addrs = append(addrs, conn.RemoteMultiaddr().String())
	}
	return addrs
}

// BootstrapPeersStats returns total peers count and connected count.
func (s *P2pService) BootstrapPeersStats() (int, int) {
	total, connected := 0, 0
	for _, peerAddr := range s.conf.GetBootstrapPeers() {
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			continue
		}
		total += 1
		if s.p2pServer.IsConnected(peerInfo.ID) {
			connected += 1
		}
	}

	return total, connected
}

func (s *P2pService) ConnectedPeersCount() int {
	return s.p2pServer.ConnectedPeersCount()
}

func (s *P2pService) RoutingTableSize() int {
	return s.p2pServer.RoutingTableSize()
}

func (s *P2pService) PeersWithAddrsCount() int {
	return s.p2pServer.PeersWithAddrsCount()
}

func (s *P2pService) AnnouncedAs() []ma.Multiaddr {
	return s.p2pServer.AnnouncedAs()
}

func (s *P2pService) OpenConnectionsCount() int {
	return s.p2pServer.OpenConnectionsCount()
}

func (s *P2pService) OpenStreamsCount() int64 {
	return s.p2pServer.OpenStreamsCount()
}

func (s *P2pService) TotalStreamsInbound() int64 {
	return s.p2pServer.TotalStreamsInbound()
}

func (s *P2pService) TotalStreamsOutbound() int64 {
	return s.p2pServer.TotalStreamsOutbound()
}

func (s *P2pService) ConnectionsLastTrimAgo() time.Duration {
	lastTrim := s.p2pServer.ConnectionsLastTrim()
	if lastTrim.IsZero() {
		lastTrim = s.startedAt
	}
	return time.Since(lastTrim)
}

func (s *P2pService) Reachability() network.Reachability {
	return s.p2pServer.Reachability()
}

func (s *P2pService) ObservedAddrs() []ma.Multiaddr {
	addrs := s.p2pServer.OwnObservedAddrs()
	return addrs
}

func (s *P2pService) NetworkStats() metrics.Stats {
	return s.p2pServer.NetworkStats()
}

func (s *P2pService) NetworkStatsByProtocol() map[protocol.ID]metrics.Stats {
	return s.p2pServer.NetworkStatsByProtocol()
}

func (s *P2pService) NetworkStatsByPeer() map[peer.ID]metrics.Stats {
	return s.p2pServer.NetworkStatsByPeer()
}

func (s *P2pService) NetworkStatsForPeer(peerID peer.ID) metrics.Stats {
	return s.p2pServer.NetworkStatsForPeer(peerID)
}

func (s *P2pService) Uptime() time.Duration {
	return time.Since(s.startedAt)
}
