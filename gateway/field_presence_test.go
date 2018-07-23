package gateway

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
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
		`{"one":{"two":"a", "three":[]}, "four": 5}`: {fieldPresenceMetaKey: []string{"Four", "One.Two", "One.Three"}},
	} {
		postReq := &http.Request{
			Method: "POST",
			Body:   ioutil.NopCloser(strings.NewReader(input)),
		}
		md := NewPresenceAnnotator("POST")(context.Background(), postReq)
		if expect == nil && md != nil {
			t.Error("Did not produce expected nil metadata")
			continue
		}
		// Because the order of objects at the same depth is not guaranteed
		sort.Strings(md[fieldPresenceMetaKey])
		sort.Strings(expect[fieldPresenceMetaKey])
		if !reflect.DeepEqual(md, expect) {
			t.Errorf("Did not produce expected metadata %+v, got %+v", expect, md)
		}

	}
}

type dummyReq struct {
	Fields *field_mask.FieldMask
}

func (r *dummyReq) GetFields() *field_mask.FieldMask {
	return r.Fields
}

func TestUnaryServerInterceptor(t *testing.T) {
	dummyInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	interceptor := PresenceClientInterceptor()
	// Test with good metadata and no present field
	md := runtime.ServerMetadata{
		HeaderMD: metadata.MD{
			fieldPresenceMetaKey: []string{"one.two.three", "one.four"},
		},
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	req := &dummyReq{}
	err := interceptor(ctx, "POST", req, nil, nil, dummyInvoker)
	if req == nil {
		t.Error("For some reason it deleted the request object")
	}
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(req.Fields, &field_mask.FieldMask{Paths: []string{"one.two.three", "one.four"}}) {
		t.Error("Didn't properly set the fieldmask in the request")
	}

	// Test with good (but arbitrary) metadata, but a present field to not overwrite
	req = &dummyReq{Fields: &field_mask.FieldMask{Paths: []string{}}}
	err = interceptor(ctx, "POST", req, nil, nil, dummyInvoker)
	if req == nil {
		t.Error("For some reason it deleted the request object")
	}
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(req.Fields, &field_mask.FieldMask{Paths: []string{}}) {
		t.Error("Wasn't supposed to alter fieldmask in request but did")
	}
}
