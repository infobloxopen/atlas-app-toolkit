// This code is derived from https://github.com/coredns/coredns
package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// NewTLSServerConfig returns a TLS config that includes a certificate
// and configures the client auth mode
// If caPath is empty, system CAs will be used
func NewTLSServerConfig(certPath, keyPath, caPath string, clientAuth tls.ClientAuthType) (*tls.Config, error) {
	cert, roots, err := loadCert(certPath, keyPath, caPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		RootCAs:      roots,
		ClientCAs:    roots,
		ClientAuth:   clientAuth,
		NextProtos:   []string{"h2"},
	}, nil
}

// NewTLSClientConfig returns a TLS config for a client connection without a client cert
// If caPath is empty, system CAs will be used
func NewTLSClientConfig(caPath string) (*tls.Config, error) {
	roots, err := loadRoots(caPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{RootCAs: roots}, nil
}

// NewTLSClientCertConfig returns a TLS config for a client connection with a client cert
func NewTLSClientCertConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	cert, roots, err := loadCert(certPath, keyPath, caPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		RootCAs:      roots,
		NextProtos:   []string{"h2"},
	}, nil
}

func loadRoots(caPath string) (*x509.CertPool, error) {
	if caPath == "" {
		return nil, nil
	}

	roots := x509.NewCertPool()
	pem, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", caPath, err)
	}
	ok := roots.AppendCertsFromPEM(pem)
	if !ok {
		return nil, fmt.Errorf("could not read root certs: %s", err)
	}
	return roots, nil
}

func loadCert(certPath, keyPath, caPath string) (*tls.Certificate, *x509.CertPool, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, nil, err
	}

	roots, err := loadRoots(caPath)
	if err != nil {
		return nil, nil, err
	}

	return &cert, roots, nil
}
