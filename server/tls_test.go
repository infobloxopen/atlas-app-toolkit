package server

import (
	"path/filepath"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/server/testdata"
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

	_, err := NewTLSConfig(cert, key, ca)
	if err != nil {
		t.Errorf("Failed to create TLSConfig: %s", err)
	}
}

func TestNewTLSClientConfig(t *testing.T) {
	rmFunc, _, _, ca := getPEMFiles(t)
	defer rmFunc()

	_, err := NewTLSClientConfig(ca)
	if err != nil {
		t.Errorf("Failed to create TLSConfig: %s", err)
	}
}
