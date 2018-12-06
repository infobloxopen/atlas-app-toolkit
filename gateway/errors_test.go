package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/errors"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProtoMessageErrorHandlerUnknownCode(t *testing.T) {
	err := fmt.Errorf("simple text error")
	v := &RestErrs{}

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

	if v.Error[0]["message"] != "simple text error" {
		t.Errorf("invalid message: %s", v.Error[0]["message"])
	}
}

func TestProtoMessageErrorHandlerUnimplementedCode(t *testing.T) {
	err := status.Error(codes.Unimplemented, "service not implemented")
	v := new(RestErrs)

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

	if v.Error[0]["message"] != "service not implemented" {
		t.Errorf("invalid message: %s", v.Error[0]["message"])
	}
}

func TestWriteErrorContainer(t *testing.T) {
	err := errors.
		NewContainer(codes.InvalidArgument, "Invalid 'x' value.").
		WithDetail(codes.InvalidArgument, "resource", "x could be one of.").
		WithDetail(codes.AlreadyExists, "resource", "x btw already exists.").
		WithField("x", "Check correct value of 'x'.")

	v := new(RestErrs)

	rw := httptest.NewRecorder()
	ProtoMessageErrorHandler(context.Background(), nil, &runtime.JSONBuiltin{}, rw, nil, err)

	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal response: %s", err)
	}

	if v.Error[0]["message"] != "Invalid 'x' value." {
		t.Errorf("invalid message: %s", v.Error[0]["message"])
	}

	if len(v.Error[0]["details"].([]interface{})) != 2 {
		t.Errorf("invalid details length: %d", len(v.Error[0]["details"].([]interface{})))
	}

	details := []interface{}{
		map[string]interface{}{
			"code":    "INVALID_ARGUMENT",
			"target":  "resource",
			"message": "x could be one of.",
		},
		map[string]interface{}{
			"code":    "ALREADY_EXISTS",
			"target":  "resource",
			"message": "x btw already exists.",
		},
	}

	if !reflect.DeepEqual(
		v.Error[0]["details"],
		details,
	) {
		t.Errorf("invalid details value: %v", v.Error[0]["details"])
	}

	fields := map[string][]string{
		"x": []string{"Check correct value of 'x'."}}

	vMap := v.Error[0]["fields"].(map[string]interface{})

	if vMap["x"].([]interface{})[0] != fields["x"][0] {
		t.Errorf("invalid fields value: %v", v.Error[0]["fields"])
	}

}
