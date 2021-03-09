package protocol

import (
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	Version = "0.1.0"

	// REMOVE
	PortForwardingMethod protocol.ID = "/peerlan/" + Version + "/forward/"
	AuthMethod           protocol.ID = "/peerlan/" + Version + "/auth/"
	GetStatusMethod      protocol.ID = "/peerlan/" + Version + "/status/"
)
