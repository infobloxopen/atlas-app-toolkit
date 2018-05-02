package appserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

// AppServer structure holds various parameters needed
// for GRPC server to function properly including optional items
// like health checks and metrics
type AppServer struct {
	grpcAddr      string
	healthAddr    string
	initTimeout   time.Duration
	initializer   Initializer
	healthHandler http.Handler
	metricsAddr   string
	metricsPath   string
	grpcServer    *grpc.Server
	httpMux       *http.ServeMux
}

type Option func(*AppServer) error
type Initializer func() error

func NewAppServer(options ...Option) (*AppServer, error) {
	s := &AppServer{}

	// call option functions on instance to set options on it
	for _, opt := range options {
		err := opt(s)
		if err != nil {
			s.cleanup()
			return nil, err
		}
	}

	return s, nil
}

func WithInitializer(initializer Initializer, initFailedTimeout time.Duration) Option {
	return func(s *AppServer) error {
		s.initializer = initializer
		s.initTimeout = initFailedTimeout
		return nil
	}
}

func WithGRPC(addr string, g *grpc.Server) Option {
	return func(s *AppServer) error {
		s.grpcAddr = addr
		s.grpcServer = g
		return nil
	}
}

func WithHealth(handler http.Handler) Option {
	return WithHealthOptions(":8080", handler)
}

func WithHealthOptions(addr string, handler http.Handler) Option {
	return func(s *AppServer) error {
		s.healthAddr = addr
		s.healthHandler = handler
		return nil
	}
}

func WithMetrics() Option {
	return WithMetricsOptions(":9153", "/metrics")
}

func WithMetricsOptions(addr, path string) Option {
	return func(s *AppServer) error {
		s.metricsAddr = addr
		s.metricsPath = path
		return nil
	}
}

func (s *AppServer) ServeWithListeners(g, h, m net.Listener) error {
	errChan := make(chan error)
	doneChan := make(chan bool)

	if s.healthAddr != "" && h != nil && s.healthHandler != nil {
		go http.ListenAndServe(s.healthAddr, s.healthHandler)
	}

	// grpc first
	if s.grpcServer != nil && g != nil {
		go func() {
			if s.initializer != nil {
				var err error
				initCheckContext, cancel := context.WithTimeout(context.Background(), s.initTimeout)

				for {
					select {
					case <-time.After(time.Second * 3):
						if err = s.initializer(); err != nil {
							continue
						}
						cancel()
					case <-initCheckContext.Done():
						err = fmt.Errorf("Initialization timeout expired. Last error: %v", err)
						cancel()
					}
					break
				}
				if err != nil {
					errChan <- err
					return
				}
			}
			if err := s.grpcServer.Serve(g); err != nil {
				errChan <- err
				return
			}
			doneChan <- true
			return
		}()
	}
	if s.metricsAddr != "" && m != nil {
	}

	// first one to error or done will end it all
	select {
	case err := <-errChan:
		return err
	case _ = <-doneChan:
		return nil
	}
}

func (s *AppServer) Serve() error {
	var g, h, m net.Listener
	var err error
	if s.grpcAddr != "" {
		g, err = net.Listen("tcp", s.grpcAddr)
		if err != nil {
			return err
		}
	}
	if s.healthAddr != "" {
		h, err = net.Listen("tcp", s.healthAddr)
		if err != nil {
			return err
		}
	}
	if s.metricsAddr != "" {
		m, err = net.Listen("tcp", s.metricsAddr)
		if err != nil {
			return err
		}
	}
	return s.ServeWithListeners(g, h, m)
}

func (s *AppServer) cleanup() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
}
