package presence

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAnnotator(t *testing.T) {
	for input, expect := range map[string]metadata.MD{
		`{}`: metadata.MD{},
		`{`:  nil,
		`{"one":{"two":"a", "three":[]}, "four": 5}`: {"grpcgateway-field-paths": []string{"Four", "One.Two", "One.Three"}},
	} {
		postReq := &http.Request{
			Method: "POST",
			Body:   ioutil.NopCloser(strings.NewReader(input)),
		}
		md := Annotator(context.Background(), postReq)
		if !reflect.DeepEqual(md, expect) {
			t.Errorf("Did not produce expected metadata %+v, got %+v", expect, md)
		}

	}
}

type dummyReq struct {
	Fields *field_mask.FieldMask
}

func TestUnaryServerInterceptor(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	// Test with good metadata and no present field
	md := runtime.ServerMetadata{
		HeaderMD: metadata.MD{
			fieldPresenceMetaKey: []string{"one.two.three", "one.four"},
		},
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	req := &dummyReq{}
	v, err := interceptor(ctx, req, nil, grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}))
	if v != nil {
		t.Error("Expecting no response object but got one")
	}
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(req.Fields, &field_mask.FieldMask{Paths: []string{"one.two.three", "one.four"}}) {
		t.Error("Didn't properly set the fieldmask in the request")
	}

	// Test with good (but arbitrary) metadata, but a present field to not overwrite
	req = &dummyReq{Fields: &field_mask.FieldMask{Paths: []string{}}}
	v, err = interceptor(ctx, req, nil, grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}))
	if v != nil {
		t.Error("Expecting no response object but got one")
	}
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(req.Fields, &field_mask.FieldMask{Paths: []string{}}) {
		t.Error("Wasn't supposed to alter fieldmask in request but did")
	}
}
