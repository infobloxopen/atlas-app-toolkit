package server

import (
	"context"
	"net"

	"net/http"

	"sync"

	"time"

	"errors"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"google.golang.org/grpc"
)

var (
	ErrInitializeTimeout      = errors.New("initialization timed out")
	DefaultInitializerTimeout = time.Minute
)

// Server is a wrapper struct that will allow you to stand up your GPRC server, API Gateway and health checks within
// the same struct. The recommended way to initialize this is with the NewServer function.
type Server struct {
	initializers      []InitializerFunc
	initializeTimeout time.Duration
	registrars        []func(mux *http.ServeMux) error

	// GRPCServer will be started whenever this is served
	GRPCServer *grpc.Server

	// HTTPServer will be started whenever this is served
	HTTPServer *http.Server
}

type Option func(*Server) error

type InitializerFunc func(context.Context) error

// NewServer creates a Server from the given options. All options are processed in the order they are declared.
func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		initializeTimeout: DefaultInitializerTimeout,
		HTTPServer:        &http.Server{},
		registrars:        []func(mux *http.ServeMux) error{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	mux := http.NewServeMux()
	for _, register := range s.registrars {
		if err := register(mux); err != nil {
			return nil, err
		}
	}
	s.HTTPServer.Handler = mux

	return s, nil
}

// WithInitializerTimeout set the duration initialization will wait before halting and returning an error
func WithInitializerTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.initializeTimeout = timeout
		return nil
	}
}

// WithInitializer adds an initialization function that will get called prior to serving.
func WithInitializer(initializerFunc InitializerFunc) Option {
	return func(s *Server) error {
		s.initializers = append(s.initializers, initializerFunc)
		return nil
	}
}

// WithGrpcServer adds the given GRPC server to this server. There can only be one GRPC server within a given instance,
// so multiple calls with this option will overwrite the previous ones.
func WithGrpcServer(grpcServer *grpc.Server) Option {
	return func(s *Server) error {
		s.GRPCServer = grpcServer
		return nil
	}
}

// WithHandler registers the given http handler to this server by registering the pattern at the root of the http server
func WithHandler(pattern string, handler http.Handler) Option {
	return func(s *Server) error {
		s.registrars = append(s.registrars, func(mux *http.ServeMux) error {
			mux.Handle(pattern, handler)
			return nil
		})
		return nil
	}
}

// WithHealthChecks registers the given health checker with this server by registering its endpoints at the root of the
// http server.
func WithHealthChecks(checker health.Checker) Option {
	return func(s *Server) error {
		s.registrars = append(s.registrars, func(mux *http.ServeMux) error {
			checker.RegisterHandler(mux)
			return nil
		})
		return nil
	}
}

// WithGateway registers the given gateway options with this server
func WithGateway(options ...gateway.Option) Option {
	return func(s *Server) error {
		s.registrars = append(s.registrars, func(mux *http.ServeMux) error {
			_, err := gateway.NewGateway(append(options, gateway.WithMux(mux))...)
			return err
		})
		return nil
	}
}

// Serve invokes all initializers then serves on the given listeners.
//
// If a listener is left blank, then that particular part will not be served.
//
// If a listener is specified for a part that doesn't have a corresponding server, then an error will be returned. This
// can happen, for instance, whenever a gRPC listener is provided but no gRPC server was set or no option was passed
// into NewServer.
func (s *Server) Serve(grpcL, httpL net.Listener) error {
	if err := s.initialize(); err != nil {
		return err
	}
	errC := make(chan error)

	if httpL != nil {
		if s.HTTPServer == nil {
			return errors.New("httpL is specified, but no HTTPServer is provided")
		}
		go func() { errC <- s.HTTPServer.Serve(httpL) }()
	} else {
		s.HTTPServer = nil
	}

	if grpcL != nil {
		if s.GRPCServer == nil {
			return errors.New("grpcL is specified, but no GRPCServer is provided")
		}
		go func() { errC <- s.GRPCServer.Serve(grpcL) }()
	} else {
		s.GRPCServer = nil
	}

	return <-errC
}

func (s *Server) Stop() error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	doneC := make(chan bool)
	errC := make(chan error)
	go func() {
		defer wg.Done()
		if s.GRPCServer != nil {
			s.GRPCServer.Stop()
		}
	}()
	go func() {
		defer wg.Done()
		if s.HTTPServer != nil {
			if err := s.HTTPServer.Close(); err != nil {
				errC <- err
			}
		}
	}()
	go func() {
		wg.Wait()
		doneC <- true
	}()
	select {
	case err := <-errC:
		return err
	case <-doneC:
		return nil
	}
}

func (s Server) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.initializeTimeout)
	defer cancel()
	errC := make(chan error)

	wg := sync.WaitGroup{}
	wg.Add(len(s.initializers))
	go func() {
		wg.Wait()
		errC <- nil
	}()

	for _, i := range s.initializers {
		go func() {
			defer wg.Done()
			if err := i(ctx); err != nil {
				errC <- err
			}
		}()
	}

	select {
	case err := <-errC:
		return err
	case <-time.After(s.initializeTimeout):
		return ErrInitializeTimeout
	}
}