package application

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anywherelan/awl-bootstrap-node/api"
	"github.com/anywherelan/awl-bootstrap-node/config"
	"github.com/anywherelan/awl-bootstrap-node/p2p"
	"github.com/anywherelan/awl-bootstrap-node/ringbuffer"
	"github.com/anywherelan/awl-bootstrap-node/service"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
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

	p2pServer  *p2p.P2p
	host       host.Host
	P2pService *service.P2pService
}

func (a *Application) Init(ctx context.Context) error {
	p2pSrv := p2p.NewP2p(ctx, a.Conf)
	host, err := p2pSrv.InitHost()
	if err != nil {
		return err
	}
	a.p2pServer = p2pSrv
	a.host = host

	privKey := host.Peerstore().PrivKey(host.ID())
	a.Conf.SetIdentity(privKey, host.ID())
	a.logger.Infof("Host created. We are: %s", host.ID().String())
	a.logger.Infof("Listen interfaces: %v", host.Addrs())

	err = p2pSrv.Bootstrap()
	if err != nil {
		return err
	}

	a.P2pService = service.NewP2p(p2pSrv, a.Conf)

	handler := api.NewHandler(a.Conf, a.P2pService, a.LogBuffer)
	handler.SetupAPI()

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
		switch {
		case strings.HasPrefix(name, "awl"):
			return lvl
		case name == "swarm2":
			// TODO решить какой выставлять
			//return zapcore.InfoLevel // REMOVE
			return zapcore.ErrorLevel
		case name == "relay":
			return zapcore.WarnLevel
			//return zapcore.ErrorLevel
		case name == "connmgr":
			return zapcore.WarnLevel
		case name == "autonat":
			return zapcore.WarnLevel
		default:
			//return lvl
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
	if a.p2pServer != nil {
		err := a.p2pServer.Close()
		if err != nil {
			a.logger.Errorf("closing p2p server: %v", err)
		}
	}
	a.Conf.Save()
}

func New() *Application {
	return &Application{}
}
