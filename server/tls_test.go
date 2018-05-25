package server

import (
	"fmt"
	"path/filepath"
	"testing"

	"net/http"

	"net/http/httptest"

	"context"

	"github.com/infobloxopen/atlas-app-toolkit/server/testdata"
	"github.com/infobloxopen/atlas-app-toolkit/servertest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getPEMFiles(t *testing.T) (rmFunc func(), cert, key, ca string) {
	tempDir, rmFunc, err := server_test.WritePEMFiles("")
	if err != nil {
		t.Fatalf("Could not write PEM files: %s", err)
	}

	cert = filepath.Join(tempDir, "cert.pem")
	key = filepath.Join(tempDir, "key.pem")
	ca = filepath.Join(tempDir, "ca.pem")

	return
}

func TestNewTLSConfig(t *testing.T) {
	rmFunc, cert, key, ca := getPEMFiles(t)
	defer rmFunc()

	tlsConfig, err := NewTLSConfig(cert, key, ca)
	if err != nil {
		t.Fatalf("Failed to create TLSConfig: %s", err)
	}

	t.Run("can issue https request", func(t *testing.T) {
		h := http.NewServeMux()
		h.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) { writer.WriteHeader(204) })

		s := httptest.NewUnstartedServer(h)
		s.TLS = tlsConfig

		s.StartTLS()
		defer s.Close()

		req := s.Client()
		res, err := req.Get(fmt.Sprint(s.URL, "/test"))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != 204 {
			t.Errorf("expected status code 204, but got %d\nres: %v", res.StatusCode, res)
		}
	})
}

func TestNewTLSClientConfig(t *testing.T) {
	rmFunc, _, _, ca := getPEMFiles(t)
	defer rmFunc()

	_, err := NewTLSClientConfig(ca)
	if err != nil {
		t.Errorf("Failed to create TLSConfig: %s", err)
	}
}

func TestCanIssueRequest(t *testing.T) {
	rmFunc, cert, key, ca := getPEMFiles(t)
	defer rmFunc()

	serverConfig, err := NewTLSConfig(cert, key, ca)
	if err != nil {
		t.Fatalf("Failed to create TLSConfig: %s", err)
	}

	// use our test grpc server with this tls
	gs := grpc.NewServer(grpc.Creds(credentials.NewTLS(serverConfig)))
	server_test.RegisterHelloServer(gs, server_test.HelloServerImpl{})
	gsl, err := servertest.NewLocalListener()
	if err != nil {
		t.Fatal(err)
	}
	go gs.Serve(gsl)
	defer gs.Stop()

	clientConfig, err := NewTLSClientConfig(ca)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := grpc.Dial(gsl.Addr().String(), grpc.WithTransportCredentials(credentials.NewTLS(clientConfig)))
	defer conn.Close()

	gc := server_test.NewHelloClient(conn)
	res, err := gc.SayHello(context.Background(), &server_test.HelloRequest{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "hello, test!"
	if res.Greeting != expected {
		t.Errorf("expected %q, but got %q", expected, res.Greeting)
	}
}
