package server

import (
	"context"
	"net"

	"net/http"

	"sync"

	"time"

	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"google.golang.org/grpc"
)

var (
	// ErrInitializeTimeout is returned when an InitializerFunc takes too long to finish during Server.Serve
	ErrInitializeTimeout = errors.New("initialization timed out")
	// DefaultInitializerTimeout is the reasonable default amount of time one would expect initialization to take in the
	// worst case
	DefaultInitializerTimeout = time.Minute
)

// Server is a wrapper struct that will allow you to stand up your GPRC server, API Gateway and health checks within
// the same struct. The recommended way to initialize this is with the NewServer function.
type Server struct {
	initializers      []InitializerFunc
	initializeTimeout time.Duration
	registrars        []func(mux *http.ServeMux) error
	middlewares       []Middleware

	// GRPCServer will be started whenever this is served
	GRPCServer *grpc.Server

	// HTTPServer will be started whenever this is served
	HTTPServer *http.Server

	isAutomaticStop bool
}

//Middleware wrapper
type Middleware func(handler http.Handler) http.Handler

// Option is a functional option for creating a Server
type Option func(*Server) error

// InitializerFunc is a handler that can be passed into WithInitializer to be executed prior to serving
type InitializerFunc func(context.Context) error

// NewServer creates a Server from the given options. All options are processed in the order they are declared.
func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		initializeTimeout: DefaultInitializerTimeout,
		HTTPServer:        &http.Server{},
		registrars:        []func(mux *http.ServeMux) error{},
		isAutomaticStop:   true,
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

	// Revert user input middlewares
	for i, j := 0, len(s.middlewares)-1; i < j; i, j = i+1, j-1 {
		s.middlewares[i], s.middlewares[j] = s.middlewares[j], s.middlewares[i]
	}

	for _, m := range s.middlewares {
		s.HTTPServer.Handler = m(s.HTTPServer.Handler)
	}

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

// WithHealthChecksContext registers the given health checker with this server by registering its endpoints at the root of the
// http server.
func WithHealthChecksContext(checker health.CheckerContext) Option {
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
			_, err := gateway.NewGateway(append([]gateway.Option{
				gateway.WithGatewayOptions(
					runtime.WithIncomingHeaderMatcher(
						gateway.AtlasDefaultHeaderMatcher())),
				gateway.WithMux(mux)},
				options...)...,
			)
			return err
		})
		return nil
	}
}

// WithMiddlewares add opportunity to add different middleware
func WithMiddlewares(middleware ...Middleware) Option {
	return func(s *Server) error {
		s.middlewares = append(s.middlewares, middleware...)
		return nil
	}
}

func WithAutomaticStop(isAutomaticStop bool) Option {
	return func(s *Server) error {
		s.isAutomaticStop = isAutomaticStop
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
//
// If both listeners are nil, then an error is returned
func (s *Server) Serve(grpcL, httpL net.Listener) error {
	if grpcL == nil && httpL == nil {
		return errors.New("both grpcL and httpL are nil")
	}

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
	defer func() {
		if s.isAutomaticStop {
			s.Stop()
		}
	}()
	return <-errC
}

// Stop immediately terminates the grpc and http servers, immediately closing their active listeners
func (s *Server) Stop() error {
	return s.shutdown(context.Background(), false)
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	return s.shutdown(ctx, true)
}

func (s Server) shutdown(ctx context.Context, isGraceful bool) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	doneC := make(chan bool)
	errC := make(chan error)
	go func() {
		defer wg.Done()
		if s.GRPCServer != nil {
			if isGraceful {
				s.GRPCServer.GracefulStop()
			} else {
				s.GRPCServer.Stop()
			}
		}
	}()
	go func() {
		defer wg.Done()
		if s.HTTPServer != nil {
			if isGraceful {
				if err := s.HTTPServer.Shutdown(ctx); err != nil {
					errC <- err
				}
			} else {
				if err := s.HTTPServer.Close(); err != nil {
					errC <- err
				}
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

	for _, initFunc := range s.initializers {
		go func(init InitializerFunc) {
			defer wg.Done()
			if err := init(ctx); err != nil {
				errC <- err
			}
		}(initFunc)
	}

	select {
	case err := <-errC:
		return err
	case <-time.After(s.initializeTimeout):
		return ErrInitializeTimeout
	}
}
