package gw

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func TestStatus(t *testing.T) {
	// test REST status from gRPC one
	rst := Status(context.Background(), status.New(codes.OK, "success message"))
	if rst.Code != CodeName(codes.OK) {
		t.Errorf("invalid status code: %s - expected: %s", rst.Code, CodeName(codes.OK))
	}
	if rst.HTTPStatus != http.StatusOK {
		t.Errorf("invalid http status code %d - expected: %d", rst.HTTPStatus, http.StatusOK)
	}
	if rst.Message != "success message" {
		t.Errorf("invalid status message: %s - expected: %s", rst.Message, "success message")
	}

	// test REST status from incoming context
	md := metadata.Pairs(
		runtime.MetadataPrefix+"status-code", CodeName(Created),
		runtime.MetadataPrefix+"status-message", "created message",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	rst = Status(ctx, nil)

	if rst.Code != CodeName(Created) {
		t.Errorf("invalid status code: %s - expected: %s", rst.Code, CodeName(Created))
	}
	if rst.HTTPStatus != http.StatusCreated {
		t.Errorf("invalid http status code %d - expected: %d", rst.HTTPStatus, http.StatusCreated)
	}
	if rst.Message != "created message" {
		t.Errorf("invalid status message: %s - expected: %s", rst.Message, "created message")
	}
}

func TestCodeName(t *testing.T) {
	// test renamed code
	if cn := CodeName(codes.Unimplemented); cn != "NOT_IMPLEMENTED" {
		t.Errorf("invalid code name: %s - expected: %s", cn, "NOT_IMPLEMENTED")
	}

	// test custom code
	if cn := CodeName(LongRunning); cn != "LONG_RUNNING_OP" {
		t.Errorf("invalid code name: %s - expected: %s", cn, "LONG_RUNNING_OP")
	}

	// test standard code
	if cn := CodeName(codes.OutOfRange); cn != code.Code_name[int32(code.Code_OUT_OF_RANGE)] {
		t.Errorf("invalid code name: %s - expected: %s", cn, code.Code_name[int32(code.Code_OUT_OF_RANGE)])
	}
}

func TestCode(t *testing.T) {
	// test renamed code
	if c := Code("NOT_IMPLEMENTED"); c != codes.Unimplemented {
		t.Errorf("invalid code: %s - expected: %s", c, codes.Unimplemented)
	}
	// test custom code
	if c := Code("LONG_RUNNING_OP"); c != LongRunning {
		t.Errorf("invalid code: %s - expected: %s", c, LongRunning)
	}
	// test standard code
	if c := Code(code.Code_name[int32(code.Code_OUT_OF_RANGE)]); c != codes.OutOfRange {
		t.Errorf("invalid code: %s - expected: %s", c, codes.OutOfRange)
	}
}

func TestHTTPStatusFromCode(t *testing.T) {
	// test overwritten code
	if sc := HTTPStatusFromCode(codes.Canceled); sc != 499 {
		t.Errorf("invalid http status: %d - expected: %d", sc, 499)
	}
	// test custom code
	if sc := HTTPStatusFromCode(Created); sc != http.StatusCreated {
		t.Errorf("invalid http status: %d - expected: %d", sc, http.StatusCreated)
	}
	// test standard code
	if sc := HTTPStatusFromCode(codes.NotFound); sc != http.StatusNotFound {
		t.Errorf("invalid http status: %d - expected: %d", sc, http.StatusNotFound)
	}
}

func TestPageInfo(t *testing.T) {
	md := metadata.Pairs(
		runtime.MetadataPrefix+pageInfoSizeMetaKey, "10",
		runtime.MetadataPrefix+pageInfoOffsetMetaKey, "100",
		runtime.MetadataPrefix+pageInfoPageTokenMetaKey, "ptoken",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	rst := Status(ctx, nil)

	if rst.Size != "10" {
		t.Errorf("invalid status size: %s - expected: %s", rst.Size, "10")
	}
	if rst.Offset != "100" {
		t.Errorf("invalid status offset: %s - expected: %s", rst.Offset, "100")
	}
	if rst.PageToken != "ptoken" {
		t.Errorf("invalid status page token: %s - expected: %s", rst.PageToken, "ptoken")
	}
}
