package metrics

import (
	"context"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// P2PDHTRequestsTotal adds the client of the sender to what
// libp2p_kad_dht_rpc_inbound_messages_total counts: the version shows how
// much of the DHT load comes from old clients.
var P2PDHTRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystemP2P,
	Name:      "dht_requests_total",
	Help:      "DHT requests received, by message type and by the client of the peer that sent them.",
}, []string{"message_type", labelNameApp, labelNameVersion})

// CountDHTRequests returns a DHT option that counts the received requests in
// P2PDHTRequestsTotal. The user agent of the sender is taken from ps.
func CountDHTRequests(ps peerstore.Peerstore) dht.Option {
	// The hook blocks the handling of the request: it has to stay cheap.
	return dht.OnRequestHook(func(_ context.Context, s network.Stream, req *pb.Message) {
		client := parseUserAgent(peerUserAgent(ps, s.Conn().RemotePeer()))
		P2PDHTRequestsTotal.WithLabelValues(messageTypeLabel(req.GetType()), client.app, client.version).Inc()
	})
}

// messageTypeLabel guards against unknown types: the hook sees every message
// before kad-dht rejects the ones it can't handle.
func messageTypeLabel(messageType pb.Message_MessageType) string {
	if _, ok := pb.Message_MessageType_name[int32(messageType)]; !ok {
		return labelOther
	}
	return messageType.String()
}
