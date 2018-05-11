package appserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewAppServer(t *testing.T) {
	appServer, err := NewAppServer()
	assert.NotNil(t, appServer, "Server Wrapper shouldn't be nil")
	assert.NoError(t, err, "Error should be nil")
	var errorOption Option = func(sw *AppServer) error { return fmt.Errorf("Option Error") }
	appServer, err = NewAppServer(errorOption)
	assert.Nil(t, appServer, "Server Wrapper should be nil")
	assert.Error(t, err, "Error shouldn't be nil")
	var noErrorOption Option = func(sw *AppServer) error { return nil }
	appServer, err = NewAppServer(noErrorOption)
	assert.NotNil(t, appServer, "Server Wrapper shouldn't be nil")
	assert.NoError(t, err, "Error should be nil")
}

func TestWithInitializerCall(t *testing.T) {
	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	initializer := func() error {
		done <- struct{}{}
		return nil
	}
	server := grpc.NewServer()
	grpcOption := WithGRPC(":9099", server)
	initOption := WithInitializer(initializer, time.Second*5)
	appServer, _ := NewAppServer(grpcOption, initOption)
	defer appServer.cleanup()
	go appServer.Serve()
	select {
	case <-ctx.Done():
		cancel()
		close(done)
		t.Errorf("Timeout reached while waiting for initializer will be called")
	case <-done:
		cancel()
		break
	}
}

func TestWithinitializerError(t *testing.T) {
	done := make(chan struct{})
	errCh := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	errString := "Error from initializer"
	initializer := func() error {
		done <- struct{}{}
		return fmt.Errorf(errString)
	}
	server := grpc.NewServer()
	grpcOption := WithGRPC(":9190", server)
	initOption := WithInitializer(initializer, time.Second*5)
	appServer, _ := NewAppServer(grpcOption, initOption)
	defer appServer.cleanup()

	go func() {
		errCh <- appServer.Serve()
	}()
	initTimeoutError := fmt.Sprintf("Initialization timeout expired. Last error: %v", errString)
	select {
	case <-ctx.Done():
		cancel()
		t.Error("Timeout reached while waiting for initializer will be called")
		break
	case <-done:
		select {
		case err := <-errCh:
			cancel()
			assert.EqualErrorf(t, err, initTimeoutError,
				"Error shouldn't be nil and equal to: %s", initTimeoutError)
			break
		case <-ctx.Done():
			cancel()
			t.Error("Timeout reached")
			break
		}
		break
	}
}

// TODO: This test requires AppServer component redesigned a little
func testWithHealthOptionsCall(t *testing.T) {
	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	mux := &http.ServeMux{}
	mux.Handle("/healthztest", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		done <- struct{}{}
	}))
	healthOption := WithHealthOptions(":8098", mux)
	appServer, _ := NewAppServer(healthOption)
	defer appServer.cleanup()

	go func() {
		// This will hang forever here due to internal channels
		// require WithGRPC option presence
		appServer.Serve()
	}()
	client := http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get("http://localhost:8098/healthztest")
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status codes unequal")
	select {
	case <-ctx.Done():
		cancel()
		t.Error("Timeout reached while waiting for health check will be called")
		break
	case <-done:
		cancel()
		break
	}
}
