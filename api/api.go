package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	http_pprof "net/http/pprof"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/anywherelan/awl/p2p"
	"github.com/anywherelan/awl/ringbuffer"
	"github.com/go-playground/validator/v10"
	"github.com/ipfs/go-log/v2"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/anywherelan/awl-bootstrap-node/config"
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

const (
	// ReadTimeout covers only reading the request: the server clears the
	// deadline once the body is read. No WriteTimeout: pprof profile and
	// trace respond for 30+ seconds.
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 20 * time.Second
	idleTimeout       = 2 * time.Minute

	// Compressing shorter responses saves nothing.
	gzipMinLength = 1024
)

// global so the /metrics collector is registered only once even if SetupAPI runs multiple times (tests).
var metricsHandler = echoprometheus.NewHandler()

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
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		// pprof profiles are gzipped already.
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, V0Prefix+"debug/pprof/")
		},
		MinLength: gzipMinLength,
	}))

	// Routes

	// Metrics
	e.GET("/metrics", metricsHandler)

	// Debug
	e.GET(GetP2pDebugInfoPath, h.GetP2pDebugInfo)
	e.GET(GetDebugLogPath, h.GetLog)

	if !h.conf.DisablePprof {
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
	e.Server.ReadHeaderTimeout = readHeaderTimeout
	e.Server.ReadTimeout = readTimeout
	e.Server.IdleTimeout = idleTimeout
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
