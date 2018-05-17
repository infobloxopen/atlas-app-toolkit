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

type Server struct {
	initializers      []InitializerFunc
	initializeTimeout time.Duration

	GRPCServer *grpc.Server
	HTTPServer *http.Server
	mux        *http.ServeMux
}

type Option func(*Server) error

type InitializerFunc func(context.Context) error

func WithInitializerTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.initializeTimeout = timeout
		return nil
	}
}

func WithInitializer(initializerFunc InitializerFunc) Option {
	return func(s *Server) error {
		s.initializers = append(s.initializers, initializerFunc)
		return nil
	}
}

func WithGrpcServer(grpcServer *grpc.Server) Option {
	return func(s *Server) error {
		s.GRPCServer = grpcServer
		return nil
	}
}

func WithHealthChecks(checker health.Checker) Option {
	return func(s *Server) error {
		s.mux.Handle("/", checker.Handler())
		return nil
	}
}

func WithGateway(options ...gateway.Option) Option {
	return func(s *Server) error {
		gwMux, err := gateway.NewGateway(options...)
		if err != nil {
			return err
		}
		s.mux.Handle("/", gwMux)
		return nil
	}
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		initializeTimeout: DefaultInitializerTimeout,
		HTTPServer:        &http.Server{},
		mux:               &http.ServeMux{},
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Serve calls invokes all initializers then serves on the given listener
func (s *Server) Serve(grpcL, httpL net.Listener) error {
	if err := s.Initialize(); err != nil {
		return err
	}

	doneC := make(chan bool)
	errC := make(chan error)
	go func() {
		defer func() { doneC <- true }()
		s.HTTPServer.Handler = s.mux
		if err := s.HTTPServer.Serve(httpL); err != nil {
			errC <- err
		}
	}()
	go func() {
		defer func() { doneC <- true }()
		if s.GRPCServer != nil {
			if err := s.GRPCServer.Serve(grpcL); err != nil {
				errC <- err
			}
		}
	}()

	select {
	case err := <-errC:
		return err
	case <-doneC:
		return nil
	}
}

func (s *Server) Stop() error {
	doneC := make(chan bool)
	errC := make(chan error)
	go func() {
		defer func() { doneC <- true }()
		if s.GRPCServer != nil {
			s.GRPCServer.Stop()
		}
	}()
	go func() {
		defer func() { doneC <- true }()
		if s.HTTPServer != nil {
			if err := s.HTTPServer.Close(); err != nil {
				errC <- err
			}
		}
	}()

	select {
	case err := <-errC:
		return err
	case <-doneC:
		return nil
	}
}

func (s Server) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.initializeTimeout)
	defer cancel()
	errC := make(chan error)
	doneC := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(len(s.initializers))
	go func() {
		wg.Wait()
		doneC <- true
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
	case <-doneC:
		return nil
	}
}
