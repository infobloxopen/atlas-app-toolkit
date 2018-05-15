package flag

import (
	"crypto/tls"
	"flag"

	atlastls "github.com/infobloxopen/atlas-app-toolkit/pkg/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCFlags struct {
	addr *string
}

type TLSFlags struct {
	cert *string
	key *string
	ca *string
}

func HealthFlags() (*string, *string) {
	addr := flag.String("health-addr", ":8080", "address to use for the health endpoint")
	path := flag.String("health-path", "/health", "path to use for the health endpoint")
	return addr, path
}

func MetricsFlags() (*string, *string) {
	addr := flag.String("metrics-addr", ":9500", "address to use for the metrics endpoint")
	path := flag.String("metrics-path", "/metrics", "path to use for the metrics endpoint")
	return addr, path
}

func NewGRPCFlags() *GRPCFlags {
	f := &GRPCFlags{}
	f.addr = flag.String("addr", ":9000", "address to use for the gRPC endpoint")
	return f
}

func (f *GRPCFlags) Addr() string {
	return *f.addr
}

func NewTLSFlags() *TLSFlags {
	f := &TLSFlags{}
	f.cert = flag.String("cert", "", "path to the server certificate in PEM format")
	f.key = flag.String("key", "", "path to the server private key in PEM format")
	f.ca = flag.String("ca", "", "path to the CA PEM for validating the client cert; system CAs will be used if blank")

	return f
}

func (f *TLSFlags) TLSConfig() (*tls.Config, error) {
	return atlastls.NewTLSConfig(*f.cert, *f.key, *f.ca)
}

func (f *TLSFlags) WithGRPCTLSCreds() (grpc.ServerOption, error) {
	if *f.cert == "" {
		return nil, nil
	}
	tlsConfig, err := f.TLSConfig()
	if err != nil {
		return nil, err
	}
	return grpc.Creds(credentials.NewTLS(tlsConfig)), nil
}
