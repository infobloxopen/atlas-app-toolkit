package server

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"fmt"

	"io/ioutil"

	"net"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/server/testdata"
	"github.com/infobloxopen/atlas-app-toolkit/servertest"
	"google.golang.org/grpc"
)

func buildTestServer(t *testing.T, opts ...Option) (grpcAddr string, httpAddr string, cleanup func()) {
	s, err := NewServer(opts...)
	if err != nil {
		t.Fatal(err)
	}

	var grpcL, httpL net.Listener

	if s.GRPCServer != nil {
		grpcL, err = servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
	}
	if s.HTTPServer != nil {
		httpL, err = servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
	}

	go s.Serve(grpcL, httpL)

	if grpcL != nil {
		grpcAddr = grpcL.Addr().String()
	}
	if httpL != nil {
		httpAddr = fmt.Sprintf("http://%s", httpL.Addr().String())
	}
	cleanup = func() {
		if err := s.Stop(); err != nil {
			t.Fatal(err)
		}
	}

	return grpcAddr, httpAddr, cleanup
}

func TestNewServer(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		_, url, cleanup := buildTestServer(t)
		defer cleanup()

		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("wasn't expecting error, but got %v", err)
		}
		if resp.StatusCode != 404 {
			t.Errorf("expected status 404, but got %d\nresponse: %v", resp.StatusCode, resp)
		}
	})

	t.Run("returns error", func(t *testing.T) {
		expected := errors.New("test error")
		if _, actual := NewServer(func(*Server) error { return expected }); actual != expected {
			t.Errorf("expected error %v, but got %v", expected, actual)
		}
	})
}

func TestWithHealthChecks(t *testing.T) {
	tests := []struct {
		name     string
		checkErr error
		testPath string
		expected int
	}{
		{"liveness-pass", nil, "healthz", 200},
		{"liveness-pass", nil, "/healthz", 200},
		{"liveness-fail", errors.New(""), "/healthz", 503},
		{"readiness-pass", nil, "ready", 200},
		{"readiness-pass", nil, "/ready", 200},
		{"readiness-fail", errors.New(""), "ready", 503},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checks := health.NewChecksHandler("healthz", "ready")
			checks.AddLiveness("liveness-test", func() error { return test.checkErr })
			_, url, cleanup := buildTestServer(t, WithHealthChecks(checks))
			defer cleanup()
			resp, err := http.Get(fmt.Sprintf("%s/%s", url, test.testPath))
			if err != nil {
				t.Errorf("not expecting error, but got %v", err)
			}
			if resp.StatusCode != test.expected {
				t.Errorf("expected status code %d, but got %d", test.expected, resp.StatusCode)
			}
		})
	}
}

func TestWithHandler(t *testing.T) {
	h := http.NewServeMux()
	h.HandleFunc("/test/204", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(204)
	})
	_, url, cleanup := buildTestServer(t, WithHandler("/test/", h))
	defer cleanup()
	resp, err := http.Get(fmt.Sprint(url, "/test/204"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("expected status code 204, but got %d", resp.StatusCode)
	}
}

func TestWithGateway(t *testing.T) {
	grpcL, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}
	httpL, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}

	gs := grpc.NewServer()
	server_test.RegisterHelloServer(gs, &server_test.HelloServerImpl{})

	s, err := NewServer(
		WithGrpcServer(gs),
		WithGateway(
			gateway.WithEndpointRegistration("/v1/", server_test.RegisterHelloHandlerFromEndpoint),
			gateway.WithServerAddress(grpcL.Addr().String()),
		),
	)

	go s.Serve(grpcL, httpL)
	defer s.Stop()

	resp, err := http.Get(fmt.Sprint("http://", httpL.Addr().String(), "/v1/hello?name=test"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, but got %d\nresponse: %v", resp.StatusCode, resp)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"greeting":"hello, test!"}`
	actual := string(respBytes)
	if expected != actual {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
}

func TestWithInitializer(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		timeout := 50 * time.Millisecond
		ctxC := make(chan bool)
		s, err := NewServer(WithInitializerTimeout(timeout), WithInitializer(func(ctx context.Context) error {
			<-ctx.Done()
			ctxC <- true
			return nil
		}))
		if err != nil {
			t.Fatal(err)
		}

		if err := s.initialize(); err != ErrInitializeTimeout {
			t.Errorf("expected timeout error, but got %v", err)
		}
		if contextCanceled := <-ctxC; !contextCanceled {
			t.Error("expected timeout to cancel context, but didn't")
		}
	})

	t.Run("init error", func(t *testing.T) {
		expected := errors.New("test error")
		s, err := NewServer(WithInitializer(func(context.Context) error { return expected }))
		if err != nil {
			t.Fatal(err)
		}
		if actual := s.initialize(); actual != expected {
			t.Errorf("expected error %v, but got %v", expected, actual)
		}
	})
}

func TestWithGrpcServer(t *testing.T) {
	gs := grpc.NewServer()
	server_test.RegisterHelloServer(gs, &server_test.HelloServerImpl{})

	gURL, _, cleanup := buildTestServer(t, WithGrpcServer(gs))
	defer cleanup()

	conn, err := grpc.Dial(gURL, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := server_test.NewHelloClient(conn)
	resp, err := client.SayHello(context.Background(), &server_test.HelloRequest{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Greeting != "hello, test!" {
		t.Errorf("expected greeting %q, but got %q", "hello, test!", resp.Greeting)
	}
}

func TestOnlyGrpcServer(t *testing.T) {
	grpcL, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}

	gs := grpc.NewServer()
	server_test.RegisterHelloServer(gs, &server_test.HelloServerImpl{})

	s, err := NewServer(WithGrpcServer(gs))

	go s.Serve(grpcL, nil)
	defer s.Stop()

	conn, err := grpc.Dial(grpcL.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := server_test.NewHelloClient(conn)
	resp, err := client.SayHello(context.Background(), &server_test.HelloRequest{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Greeting != "hello, test!" {
		t.Errorf("expected greeting %q, but got %q", "hello, test!", resp.Greeting)
	}
}

func TestOnlyHttpServer(t *testing.T) {
	h := http.NewServeMux()
	h.HandleFunc("/test/204", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(204)
	})

	httpL, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewServer(WithHandler("/test/", h))
	if err != nil {
		t.Fatal(err)
	}

	go s.Serve(nil, httpL)
	defer s.Stop()

	resp, err := http.Get(fmt.Sprint("http://", httpL.Addr().String(), "/test/204"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("expected status code 204, but got %d\nresponse: %v", resp.StatusCode, resp)
	}
}

func TestServe(t *testing.T) {
	t.Run("both listeners are nil", func(t *testing.T) {
		s, err := NewServer()
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Serve(nil, nil); err == nil {
			t.Error("expected error, but got none")
		}
	})
	t.Run("grpc listener but no server", func(t *testing.T) {
		l, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		s, err := NewServer()
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Serve(l, nil); err == nil {
			t.Error("expected error, but got none")
		}
	})
	t.Run("http listener but no server", func(t *testing.T) {
		l, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		s, err := NewServer()
		if err != nil {
			t.Fatal(err)
		}
		s.HTTPServer = nil
		if err := s.Serve(nil, l); err == nil {
			t.Error("expected error, but got none")
		}
	})
	// if one part fails, this whole server should go down. This test ensures that if the http server crashes for some
	// reason (or is brought explicitly brought down), then the gRPC server will close, too
	t.Run("closing http server closes grpc server", func(t *testing.T) {
		// bunch of setup code
		grpcL, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		httpL, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		grpcServer := grpc.NewServer()
		server_test.RegisterHelloServer(grpcServer, &server_test.HelloServerImpl{})
		s, err := NewServer(WithGrpcServer(grpcServer))
		if err != nil {
			t.Fatal(err)
		}
		// start the server
		go s.Serve(grpcL, httpL)

		// demonstrate that we can reach the gRPC server
		conn, err := grpc.Dial(grpcL.Addr().String(), grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}
		client := server_test.NewHelloClient(conn)
		if _, err := client.SayHello(context.Background(), &server_test.HelloRequest{Name: "test"}); err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}
		// now if we kill the HTTP server, the gRPC server should close, too
		s.HTTPServer.Close()
		if _, err := client.SayHello(context.Background(), &server_test.HelloRequest{Name: "test"}); err == nil {
			t.Fatal("expected grpc server to be closed, but request was successfully sent")
		}
	})
	t.Run("closing grpc server closes http server", func(t *testing.T) {
		// bunch of setup code
		grpcL, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		httpL, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		grpcServer := grpc.NewServer()
		server_test.RegisterHelloServer(grpcServer, &server_test.HelloServerImpl{})
		h := http.NewServeMux()
		h.HandleFunc("/test/204", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(204)
		})
		s, err := NewServer(
			WithGrpcServer(grpcServer),
			WithHandler("/test/", h),
		)
		if err != nil {
			t.Fatal(err)
		}
		// start the server
		go s.Serve(grpcL, httpL)

		resp, err := http.Get(fmt.Sprint("http://", httpL.Addr().String(), "/test/204"))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 204 {
			t.Errorf("expected status code 204, but got %d", resp.StatusCode)
		}

		s.GRPCServer.Stop()

		if _, err = http.Get(fmt.Sprint("http://", httpL.Addr().String(), "/test/204")); err == nil {
			t.Fatal("expected http server to be closed, but request was successfully sent")
		}
	})
}

func TestStop(t *testing.T) {
	s, err := NewServer()
	if err != nil {
		t.Fatal(err)
	}

	doneC := make(chan bool)
	go func() {
		httpL, err := servertest.NewLocalListener()
		if err != nil {
			t.Fatal(err)
		}
		s.Serve(nil, httpL)
		doneC <- true
	}()

	time.Sleep(50 * time.Millisecond)
	s.Stop()
	<-doneC
}
