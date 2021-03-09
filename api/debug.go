package api

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/peerlan/bootstrap-node/entity"
)

// @Tags Debug
// @Summary Get p2p debug info
// @Accept json
// @Produce json
// @Success 200 {object} entity.P2pDebugInfo
// @Router /debug/p2p_info [GET]
func (h *Handler) GetP2pDebugInfo(c echo.Context) (err error) {
	metricsByProtocol := h.p2p.NetworkStatsByProtocol()
	bandwidthByProtocol := make(map[string]entity.BandwidthInfo, len(metricsByProtocol))
	for key, val := range metricsByProtocol {
		bandwidthByProtocol[string(key)] = makeBandwidthInfo(val)
	}

	debugInfo := entity.P2pDebugInfo{
		General: entity.GeneralDebugInfo{
			Uptime: h.p2p.Uptime().String(),
		},
		DHT: entity.DhtDebugInfo{
			RoutingTableSize:    h.p2p.RoutingTableSize(),
			Reachability:        reachabilityToString(h.p2p.Reachability()),
			ListenAddress:       maToStrings(h.p2p.AnnouncedAs()),
			PeersWithAddrsCount: h.p2p.PeersWithAddrsCount(),
			ObservedAddrs:       maToStrings(h.p2p.ObservedAddrs()),
			BootstrapPeers:      h.p2p.BootstrapPeersStatsDetailed(),
		},
		Connections: entity.ConnectionsDebugInfo{
			ConnectedPeersCount:  h.p2p.ConnectedPeersCount(),
			OpenConnectionsCount: h.p2p.OpenConnectionsCount(),
			OpenStreamsCount:     h.p2p.OpenStreamsCount(),
			TotalStreamsInbound:  h.p2p.TotalStreamsInbound(),
			TotalStreamsOutbound: h.p2p.TotalStreamsOutbound(),
			LastTrimAgo:          h.p2p.ConnectionsLastTrimAgo().String(),
		},
		Bandwidth: entity.BandwidthDebugInfo{
			Total:      makeBandwidthInfo(h.p2p.NetworkStats()),
			ByProtocol: bandwidthByProtocol,
			//ByPeer:     h.p2p.NetworkStatsByPeer(),
		},
	}

	return c.JSONPretty(http.StatusOK, debugInfo, "    ")
}

// @Tags Debug
// @Summary Get logs
// @Accept json
// @Produce json
// @Success 200 {string} string "log text"
// @Router /debug/log [GET]
func (h *Handler) GetLog(c echo.Context) (err error) {
	b := h.logBuffer.Bytes()
	return c.Blob(http.StatusOK, echo.MIMETextPlainCharsetUTF8, b)
}

func makeBandwidthInfo(stats metrics.Stats) entity.BandwidthInfo {
	return entity.BandwidthInfo{
		TotalIn:  byteCountIEC(stats.TotalIn),
		TotalOut: byteCountIEC(stats.TotalOut),
		RateIn:   byteCountIEC(int64(stats.RateIn)),
		RateOut:  byteCountIEC(int64(stats.RateOut)),
	}
}

func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func maToStrings(addrs []ma.Multiaddr) []string {
	res := make([]string, 0, len(addrs))
	for i := range addrs {
		res = append(res, addrs[i].String())
	}
	sort.Strings(res)

	return res
}

func reachabilityToString(reachability network.Reachability) string {
	switch reachability {
	case network.ReachabilityUnknown:
		return "unknown"
	case network.ReachabilityPublic:
		return "public"
	case network.ReachabilityPrivate:
		return "private"
	default:
		return ""
	}
}
