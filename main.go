package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anywherelan/awl/p2p"
	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"go.uber.org/zap/zapcore"

	"github.com/anywherelan/awl-bootstrap-node/config"
)

// @title Anywherelan bootstrap node API
// @version 0.1
// @description Anywherelan bootstrap node API

// @Host localhost:9090
// @BasePath /api/v0/

//go:generate swag init --parseDependency
//go:generate rm -f docs/docs.go docs/swagger.json
func main() {
	configName := flag.String("config-name", "config.yaml", "Config filepath")
	generateConfig := flag.Bool("generate-config", false, "Enable generating config instead of starting the server")
	flag.Parse()

	// go run main.go -generate-config -config-name=config.yaml
	if *generateConfig {
		generateExampleConfig(*configName)
		return
	}

	app := New()
	logger := app.SetupLoggerAndConfig()
	ctx, ctxCancel := context.WithCancel(context.Background())

	err := app.Init(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	logger.Infof("received signal %s", <-quit)
	finishedCh := make(chan struct{})
	go func() {
		select {
		case <-time.After(3 * time.Second):
			logger.Fatal("exit timeout reached: terminating")
		case <-finishedCh:
			// ok
		case sig := <-quit:
			logger.Fatalf("duplicate exit signal %s: terminating", sig)
		}
	}()
	ctxCancel()
	app.Close()

	finishedCh <- struct{}{}
	logger.Info("exited normally")
}

func generateExampleConfig(configName string) {
	_, err := os.Stat(configName)
	if os.IsNotExist(err) {
		// ok
	} else if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Config file %s already exists, specify different name with flag -config-name=new-config.yaml\n", configName)
		os.Exit(1)
	}

	// disable logs
	_ = os.Setenv("QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING", "true")
	log.SetupLogging(zapcore.NewNopCore(), func(name string) zapcore.Level {
		return zapcore.FatalLevel
	})

	conf := config.NewExampleConfig()

	peerstore, _ := pstoremem.NewPeerstore()
	p2pSrv := p2p.NewP2p(context.Background())
	host, err := p2pSrv.InitHost(p2p.HostConfig{
		Peerstore:    peerstore,
		DHTDatastore: dssync.MutexWrap(ds.NewMapDatastore()),
	})
	if err != nil {
		fmt.Printf("Error initializing test host: %s\n", err)
		os.Exit(1)
	}

	privKey := host.Peerstore().PrivKey(host.ID())
	conf.SetIdentity(privKey, host.ID())
	_ = p2pSrv.Close()

	err = config.SaveConfig(conf, configName)
	if err != nil {
		fmt.Printf("Error saving config file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated example config file: %s\n", configName)
}
