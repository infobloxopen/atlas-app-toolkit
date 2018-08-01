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
		`{
  "name": "atlas",
  "burden": {
    "duration": "forever",
    "weight": "earth",
    "breaks": [],
    "replacements": {
			"hero": {
	        "name": "hercules",
	        "duration": "temporary",
					"lineage": {
						"mother": "alcmena",
						"father": "zeus"
					}
	      },
			"mortals": []
		}
  }
}`: {fieldPresenceMetaKey: []string{"Name", "Burden.Duration", "Burden.Weight",
			"Burden.Breaks", "Burden.Replacements.Hero.Name", "Burden.Replacements.Hero.Duration",
			"Burden.Replacements.Hero.Lineage.Mother", "Burden.Replacements.Hero.Lineage.Father", "Burden.Replacements.Mortals"}},
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
	SomeFieldMaskField *field_mask.FieldMask
}

type testReqWithoutFieldMask struct {
	foo string
	bar *dummyReq
	baz *int
}

func TestUnaryServerInterceptor(t *testing.T) {
	dummyInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}
	interceptor := PresenceClientInterceptor()
	md := runtime.ServerMetadata{
		HeaderMD: metadata.MD{
			fieldPresenceMetaKey: []string{"one.two.three", "one.four"},
		},
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)

	t.Run("sets FieldMask if nil", func(t *testing.T) {
		req := &dummyReq{}
		if err := interceptor(ctx, "POST", req, nil, nil, dummyInvoker); err != nil {
			t.Fatal(err)
		}
		if req == nil {
			t.Fatal("For some reason it deleted the request object")
		}
		got, want := req.SomeFieldMaskField, &field_mask.FieldMask{Paths: []string{"one.two.three", "one.four"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Didn't properly set the fieldmask in the request.\ngot :%v\nwant:%v", got, want)
		}
	})
	t.Run("doesn't set FieldMask if not nil", func(t *testing.T) {
		// Test with good (but arbitrary) metadata, but a present field to not overwrite
		req := &dummyReq{SomeFieldMaskField: &field_mask.FieldMask{Paths: []string{}}}
		err := interceptor(ctx, "POST", req, nil, nil, dummyInvoker)
		if req == nil {
			t.Error("For some reason it deleted the request object")
		}
		if err != nil {
			t.Error(err.Error())
		}
		if !reflect.DeepEqual(req.SomeFieldMaskField, &field_mask.FieldMask{Paths: []string{}}) {
			t.Error("Wasn't supposed to alter fieldmask in request but did")
		}
	})
	t.Run("works if no FieldMask in request", func(t *testing.T) {
		req := &testReqWithoutFieldMask{foo: "bar"}
		if err := interceptor(ctx, "POST", req, nil, nil, dummyInvoker); err != nil {
			t.Error(err.Error())
		}
	})
}
