package server

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"fmt"

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

	grpcL, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}
	httpL, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}

	go s.Serve(grpcL, httpL)

	close := func() {
		if err := s.Stop(); err != nil {
			t.Fatal(err)
		}
	}
	return grpcL.Addr().String(), fmt.Sprintf("http://%s", httpL.Addr().String()), close
}

func TestNewServer(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		_, url, close := buildTestServer(t)
		defer close()

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

func TestWithHandler(t *testing.T) {
	h := http.NewServeMux()
	h.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(204)
		writer.Write([]byte("test"))
	})
	_, url, close := buildTestServer(t, WithHandler(h))
	defer close()
	resp, err := http.Get(fmt.Sprint(url, "/test"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("expected status code 204, but got %d", resp.StatusCode)
	}
}

func TestWithHealthChecks(t *testing.T) {
	tests := []struct {
		name     string
		checkErr error
		testPath string
		expected int
	}{
		{"liveness-pass", nil, "healthz", 200},
		{"liveness-fail", errors.New(""), "healthz", 503},
		{"readiness-pass", nil, "ready", 200},
		{"readiness-fail", errors.New(""), "ready", 503},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checks := health.NewChecksHandler("healthz", "ready")
			checks.AddLiveness("liveness-test", func() error { return test.checkErr })
			_, url, close := buildTestServer(t, WithHealthChecks(checks))
			defer close()
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

		if err := s.Initialize(); err != ErrInitializeTimeout {
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
		if actual := s.Initialize(); actual != expected {
			t.Errorf("expected error %v, but got %v", expected, actual)
		}
	})
}

func TestWithGrpcServer(t *testing.T) {
	gs := grpc.NewServer()
	server_test.RegisterHelloServer(gs, &server_test.HelloServerImpl{})

	gURL, _, close := buildTestServer(t, WithGrpcServer(gs))
	defer close()

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
