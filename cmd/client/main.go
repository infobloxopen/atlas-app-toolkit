package main

import (
	"log"

	"context"

	"time"

	"github.com/infobloxopen/atlas-app-toolkit/cmd"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	certPath = "cert/cert.pem"
	keyPath  = "cert/key.pem"
	caPath   = certPath
	grpcAddr = "localhost:9091"
)

func main() {
	tlsConfig, err := server.NewTLSClientConfig(caPath)
	if err != nil {
		log.Fatalf("error creating tls config: %v", err)
	}

	dialOption := grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	//dialOption := grpc.WithInsecure()
	conn, err := grpc.Dial(grpcAddr, dialOption)
	if err != nil {
		log.Fatalf("failed to dial server: %s", err)
	}
	defer conn.Close()

	client := cmd.NewHelloClient(conn)

	for range time.NewTicker(time.Second).C {
		resp, err := client.SayHello(context.Background(), &cmd.HelloRequest{Name: "Daniel"})
		if err != nil {
			log.Printf("error saying hello: %v", err)
		} else {
			log.Print(resp.Greeting)
		}
	}
}
