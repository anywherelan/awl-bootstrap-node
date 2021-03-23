module github.com/anywherelan/awl-bootstrap-node

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/ipfs/go-datastore v0.4.4
	github.com/ipfs/go-ds-badger v0.2.4
	github.com/ipfs/go-log/v2 v2.0.5
	github.com/labstack/echo/v4 v4.1.16
	github.com/libp2p/go-eventbus v0.2.1
	github.com/libp2p/go-libp2p v0.10.0
	github.com/libp2p/go-libp2p-circuit v0.2.3
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.6.0
	github.com/libp2p/go-libp2p-kad-dht v0.8.1
	github.com/libp2p/go-libp2p-noise v0.1.1
	github.com/libp2p/go-libp2p-peerstore v0.2.6
	github.com/libp2p/go-libp2p-quic-transport v0.6.0
	github.com/libp2p/go-libp2p-swarm v0.2.7
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/libp2p/go-tcp-transport v0.2.0
	github.com/libp2p/go-ws-transport v0.3.1
	github.com/multiformats/go-multiaddr v0.2.2
	go.uber.org/multierr v1.5.0
	go.uber.org/zap v1.15.0
)

replace (
	github.com/ipfs/go-log/v2 => github.com/anywherelan/go-log/v2 v2.0.3-0.20210308150645-ad120b957e42
	github.com/libp2p/go-libp2p-swarm => github.com/anywherelan/go-libp2p-swarm v0.2.8-0.20210308145331-4dade4a1a222
)
