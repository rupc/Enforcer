package server

import (
	"net"
	"net/http"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/rupc/Enforcer/server/handler"
)

type System struct {
	options    Options
	httpServer *http.Server
	mux        *http.ServeMux
	addr       string
	// versionGauge metrics.Gauge
	logger *flogging.FabricLogger
}

type MetricsOptions struct {
	Provider string
}

type Options struct {
	ListenAddress string
	Version       string
	Logger        *flogging.FabricLogger
	// Metrics       MetricsOptions
}

func NewSystem(o Options) *System {
	logger := o.Logger
	if logger == nil {
		logger = flogging.MustGetLogger("audit.server")
	}

	system := &System{
		logger:  logger,
		options: o,
	}

	system.initializeServer()
	system.initializeAuditService()

	return system
}

func (s *System) Start() error {
	listener, err := s.listen()
	if err != nil {
		return err
	}
	s.addr = listener.Addr().String()

	s.logger.Infof("Audit server started[%s]", s.addr)
	go s.httpServer.Serve(listener)

	return nil
}

func (s *System) listen() (net.Listener, error) {
	listener, err := net.Listen("tcp", s.options.ListenAddress)
	if err != nil {
		return nil, err
	}
	// tlsConfig, err := s.options.TLS.Config()
	// if err != nil {
	//     return nil, err
	// }
	// if tlsConfig != nil {
	//     listener = tls.NewListener(listener, tlsConfig)
	// }
	return listener, nil
}

func (s *System) initializeServer() {
	s.mux = http.NewServeMux()

	s.httpServer = &http.Server{
		Addr:         s.options.ListenAddress,
		Handler:      s.mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 2 * time.Minute,
	}
}

func (s *System) initializeAuditService() error {
	ah := handler.GetAuditHandler()
	nbh := handler.GetNewBlockHandler()
	qh := handler.GetQueryHandler()
	s.mux.Handle("/audit", ah)
	s.mux.Handle("/new_block", nbh)
	s.mux.Handle("/query", qh)
	return nil
}

func (s *System) Addr() string {
	return s.addr
}
