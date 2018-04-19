package gw

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProtoMessageErrorHandlerUnknownCode(t *testing.T) {
	err := fmt.Errorf("simple text error")
	v := new(RestError)

	rw := httptest.NewRecorder()
	ProtoMessageErrorHandler(context.Background(), nil, &runtime.JSONBuiltin{}, rw, nil, err)

	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}
	if rw.Code != http.StatusInternalServerError {
		t.Errorf("invalid http status code: %d - expected: %d", rw.Code, http.StatusInternalServerError)
	}

	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal response: %s", err)
	}

	if v.Status.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("invalid http status: %d", v.Status.HTTPStatus)
	}

	if v.Status.Code != code.Code_UNKNOWN.String() {
		t.Errorf("invalid code: %s", v.Status.Code)
	}

	if v.Status.Message != "simple text error" {
		t.Errorf("invalid message: %s", v.Status.Message)
	}
}

func TestProtoMessageErrorHandlerUnimplementedCode(t *testing.T) {
	err := status.Error(codes.Unimplemented, "service not implemented")
	v := new(RestError)

	rw := httptest.NewRecorder()
	ProtoMessageErrorHandler(context.Background(), nil, &runtime.JSONBuiltin{}, rw, nil, err)

	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}
	if rw.Code != http.StatusNotImplemented {
		t.Errorf("invalid status code: %d - expected: %d", rw.Code, http.StatusNotImplemented)
	}

	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal response: %s", err)
	}

	if v.Status.HTTPStatus != http.StatusNotImplemented {
		t.Errorf("invalid http status: %d", v.Status.HTTPStatus)
	}

	if v.Status.Code != "NOT_IMPLEMENTED" {
		t.Errorf("invalid code: %s", v.Status.Code)
	}

	if v.Status.Message != "service not implemented" {
		t.Errorf("invalid message: %s", v.Status.Message)
	}
}
