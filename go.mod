module github.com/anywherelan/awl-bootstrap-node

go 1.16

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.9.0
	github.com/ipfs/go-datastore v0.4.6
	github.com/ipfs/go-ds-badger v0.2.7
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/labstack/echo/v4 v4.6.1
	github.com/libp2p/go-eventbus v0.2.1
	github.com/libp2p/go-libp2p v0.15.1
	github.com/libp2p/go-libp2p-circuit v0.4.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.9.0
	github.com/libp2p/go-libp2p-kad-dht v0.13.1
	github.com/libp2p/go-libp2p-noise v0.2.2
	github.com/libp2p/go-libp2p-peerstore v0.3.0
	github.com/libp2p/go-libp2p-quic-transport v0.11.2
	github.com/libp2p/go-libp2p-swarm v0.5.3
	github.com/libp2p/go-libp2p-tls v0.2.0
	github.com/libp2p/go-tcp-transport v0.2.8
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.4.1
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
)

replace github.com/ipfs/go-log/v2 => github.com/anywherelan/go-log/v2 v2.0.3-0.20210308150645-ad120b957e42
