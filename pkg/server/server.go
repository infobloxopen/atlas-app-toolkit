package server

import (
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type Server struct {
	grpcAddr    string
	healthAddr  string
	healthPath  string
	metricsAddr string
	metricsPath string
	grpcServer  *grpc.Server
	httpMux     *http.ServeMux
}

type Option func(*Server) error

func New(options ...Option) (*Server, error) {
	s := &Server{}

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

func WithGRPC(addr string, g *grpc.Server) Option {
	return func(s *Server) error {
		s.grpcAddr = addr
		s.grpcServer = g
		return nil
	}
}

func WithHealth() Option {
	return WithHealthOptions(":8080", "/health")
}

func WithHealthOptions(addr, path string) Option {
	return func(s *Server) error {
		s.healthAddr = addr
		s.healthPath = path
		return nil
	}
}

func WithMetrics() Option {
	return WithMetricsOptions(":9153", "/metrics")
}

func WithMetricsOptions(addr, path string) Option {
	return func(s *Server) error {
		s.metricsAddr = addr
		s.metricsPath = path
		return nil
	}
}

func (s *Server) ServeWithListeners(g, h, m net.Listener) error {
	errChan := make(chan error)
	doneChan := make(chan bool)

	// grpc first
	if s.grpcServer != nil && g != nil {
		go func() {
			if err := s.grpcServer.Serve(g); err != nil {
				errChan <- err
				return
			}
			doneChan <- true
			return
		}()
	}
	if s.healthAddr != "" && h != nil {
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

func (s *Server) Serve() error {
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

func (s *Server) cleanup() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
}
