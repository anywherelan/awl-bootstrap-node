package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Anywherelan bootstrap node API
// @version 0.1
// @description Anywherelan bootstrap node API

// @Host localhost:9090
// @BasePath /api/v0/

//go:generate swag init --parseDependency
//go:generate rm -f docs/docs.go docs/swagger.json
func main() {
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
