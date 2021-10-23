package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	http_pprof "net/http/pprof"
	"runtime/pprof"

	"github.com/anywherelan/awl-bootstrap-node/config"
	"github.com/anywherelan/awl/p2p"
	"github.com/anywherelan/awl/ringbuffer"
	"github.com/go-playground/validator/v10"
	"github.com/ipfs/go-log/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handler struct {
	echo      *echo.Echo
	conf      *config.Config
	logger    *log.ZapEventLogger
	p2p       *p2p.P2p
	logBuffer *ringbuffer.RingBuffer
}

func NewHandler(conf *config.Config, p2p *p2p.P2p, logBuffer *ringbuffer.RingBuffer) *Handler {
	return &Handler{
		conf:      conf,
		p2p:       p2p,
		logger:    log.Logger("awl/api"),
		logBuffer: logBuffer,
	}
}

func (h *Handler) SetupAPI() error {
	e := echo.New()
	h.echo = e
	e.HideBanner = true
	e.HidePort = true
	val := validator.New()
	e.Validator = &customValidator{validator: val}

	// Middleware
	if !h.conf.DevMode() {
		e.Use(middleware.Recover())
	}

	// Routes

	// Debug
	e.GET(GetP2pDebugInfoPath, h.GetP2pDebugInfo)
	e.GET(GetDebugLogPath, h.GetLog)

	if h.conf.DevMode() {
		e.Any(V0Prefix+"debug/pprof/", echo.WrapHandler(http.HandlerFunc(http_pprof.Index)))
		e.Any(V0Prefix+"debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(http_pprof.Profile)))
		e.Any(V0Prefix+"debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(http_pprof.Trace)))
		e.Any(V0Prefix+"debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(http_pprof.Cmdline)))
		e.Any(V0Prefix+"debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(http_pprof.Symbol)))

		for _, p := range pprof.Profiles() {
			name := p.Name()
			e.Any(V0Prefix+"debug/pprof/"+name, echo.WrapHandler(http_pprof.Handler(name)))
		}
	}

	// Start
	address := h.conf.HttpListenAddress
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("unable to bind address %s: %v", address, err)
	}
	h.echo.Listener = listener
	h.logger.Infof("starting web server on http://%s", listener.Addr().String())
	go func() {
		if err := e.StartServer(e.Server); err != nil && err != http.ErrServerClosed {
			h.logger.Warnf("shutting down web server %s: %s", address, err)
		}
	}()

	return nil
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.echo.Server.Shutdown(ctx)
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

type Error struct {
	Message string `json:"error"`
}

func (e Error) Error() string {
	return e.Message
}

func ErrorMessage(message string) Error {
	return Error{Message: message}
}
