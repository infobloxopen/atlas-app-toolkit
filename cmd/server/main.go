package main

import (
	"log"
	"net"
	"net/http"

	"crypto/tls"

	"github.com/infobloxopen/atlas-app-toolkit/cmd"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"google.golang.org/grpc"
)

const (
	certPath = "cert/cert.pem"
	keyPath  = "cert/key.pem"
	caPath   = certPath
	httpAddr = "localhost:9090"
	grpcAddr = "localhost:9091"
)

func main() {
	grpcServer := grpc.NewServer()
	cmd.RegisterHelloServer(grpcServer, cmd.HelloServerImpl{})

	mux := http.NewServeMux()
	mux.HandleFunc("/foobar/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("this is a test endpoint\n"))
	})

	s, err := server.NewServer(server.WithGrpcServer(grpcServer))
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig, err := server.NewTLSConfig(certPath, keyPath, caPath)
	if err != nil {
		log.Fatal(err)
	}
	l = tls.NewListener(l, tlsConfig)

	if err := s.Serve(l, l); err != nil {
		log.Fatal(err)
	}
}
