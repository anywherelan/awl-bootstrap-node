package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anywherelan/awl-bootstrap-node/api"
	"github.com/anywherelan/awl-bootstrap-node/config"
	"github.com/anywherelan/awl/p2p"
	"github.com/anywherelan/awl/ringbuffer"
	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logBufSize = 30 * 1024
)

type Application struct {
	LogBuffer *ringbuffer.RingBuffer
	logger    *log.ZapEventLogger
	Conf      *config.Config
	ctx       context.Context

	Api       *api.Handler
	p2pServer *p2p.P2p
}

func New() *Application {
	return &Application{}
}

func (a *Application) Init(ctx context.Context) error {
	a.ctx = ctx
	p2pSrv := p2p.NewP2p(ctx)
	host, err := p2pSrv.InitHost(a.makeP2pHostConfig())
	if err != nil {
		return err
	}
	a.p2pServer = p2pSrv

	privKey := host.Peerstore().PrivKey(host.ID())
	a.Conf.SetIdentity(privKey, host.ID())
	a.logger.Infof("Host created. We are: %s", host.ID().String())
	a.logger.Infof("Listen interfaces: %v", host.Addrs())

	err = p2pSrv.Bootstrap()
	if err != nil {
		return err
	}

	handler := api.NewHandler(a.Conf, a.p2pServer, a.LogBuffer)
	a.Api = handler
	err = handler.SetupAPI()
	if err != nil {
		return fmt.Errorf("failed to setup api: %v", err)
	}

	if a.Conf.P2pNode.ExchangeIdentityWithPeersInBackground {
		a.exchangeIdentityWithPeersInBackground(p2pSrv)
	}

	return nil
}

func (a *Application) SetupLoggerAndConfig() *log.ZapEventLogger {
	// Config
	conf, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("ERROR anywherelan: failed to read config file, creating new one: %v\n", err)
		conf = config.NewConfig()
	}

	// Logger
	a.LogBuffer = ringbuffer.New(logBufSize)
	syncer := zapcore.NewMultiWriteSyncer(
		zapcore.Lock(zapcore.AddSync(os.Stdout)),
		zapcore.AddSync(a.LogBuffer),
	)

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	zapCore := zapcore.NewCore(consoleEncoder, syncer, zapcore.InfoLevel)

	lvl := conf.LogLevel()
	opts := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel)}
	if conf.DevMode() {
		opts = append(opts, zap.Development())
	}

	log.SetupLogging(zapCore, func(name string) zapcore.Level {
		if strings.HasPrefix(name, "awl") {
			return lvl
		}
		switch name {
		case "swarm2", "relay", "connmgr", "autonat":
			return zapcore.WarnLevel
		default:
			return zapcore.InfoLevel
		}
	},
		opts...,
	)

	a.logger = log.Logger("awl")
	a.Conf = conf

	return a.logger
}

func (a *Application) Close() {
	if a.Api != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := a.Api.Shutdown(ctx)
		if err != nil {
			a.logger.Errorf("closing api server: %v", err)
		}
	}
	if a.p2pServer != nil {
		err := a.p2pServer.Close()
		if err != nil {
			a.logger.Errorf("closing p2p server: %v", err)
		}
	}
	a.Conf.Save()
}

func (a *Application) makeP2pHostConfig() p2p.HostConfig {
	var datastore ds.Batching
	datastore, err := badger.NewDatastore(a.Conf.PeerstoreDir(), nil)
	if err != nil {
		a.logger.DPanicf("could not create badger datastore: %v", err)
		datastore = dssync.MutexWrap(ds.NewMapDatastore())
	}
	var customPeerstore peerstore.Peerstore
	customPeerstore, err = pstoreds.NewPeerstore(a.ctx, datastore, pstoreds.DefaultOpts())
	if err != nil {
		a.logger.DPanicf("could not create peerstore: %v", err)
		customPeerstore, err = pstoremem.NewPeerstore()
		if err != nil {
			panic(err)
		}
	}

	relayResourcesCfg := relay.Resources{
		Limit: &relay.RelayLimit{
			Duration: 2 * time.Minute,
			// the same as in `backgroundOutboundHandler` func in awl/service/tunnel.go
			Data: 1024 * 1024 * (8 + 1),
		},

		ReservationTTL: time.Hour,

		MaxReservations: 512,
		MaxCircuits:     16,
		// vpn interface MTU + protocol message size packet
		BufferSize: 3500 + 8,

		MaxReservationsPerPeer: 8,
		MaxReservationsPerIP:   8,
		MaxReservationsPerASN:  32,
	}

	return p2p.HostConfig{
		PrivKeyBytes:   a.Conf.PrivKey(),
		ListenAddrs:    a.Conf.GetListenAddresses(),
		UserAgent:      config.UserAgent,
		BootstrapPeers: a.Conf.GetBootstrapPeers(),
		Libp2pOpts: []libp2p.Option{
			libp2p.EnableRelay(),
			libp2p.EnableRelayService(relay.WithResources(relayResourcesCfg)),
			libp2p.EnableHolePunching(),
			libp2p.EnableNATService(),
			libp2p.AutoNATServiceRateLimit(0, 2, time.Second),
			libp2p.ForceReachabilityPublic(),
		},
		ConnManager: struct {
			LowWater    int
			HighWater   int
			GracePeriod time.Duration
		}{
			LowWater:    0,
			HighWater:   0,
			GracePeriod: time.Minute,
		},
		Peerstore:    customPeerstore,
		DHTDatastore: datastore,
		DHTOpts: []dht.Option{
			dht.Mode(dht.ModeServer),
		},
	}
}

func (a *Application) exchangeIdentityWithPeersInBackground(p2pSrv *p2p.P2p) {
	checkPeer := func(id peer.ID) bool {
		return len(p2pSrv.Host().Peerstore().Addrs(id)) != 0
	}
	p2pSrv.SubscribeConnectionEvents(func(_ network.Network, conn network.Conn) {
		id := conn.RemotePeer()
		if checkPeer(id) {
			return
		}
		p2pSrv.IDService().IdentifyWait(conn)
	}, nil)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				// ok
			}

			allConns := p2pSrv.Host().Network().Conns()
			for _, conn := range allConns {
				id := conn.RemotePeer()
				if checkPeer(id) {
					continue
				}

				// TODO: try to remove peer from IDService before re-identifying
				p2pSrv.IDService().IdentifyWait(conn)
			}
		}
	}()
}
