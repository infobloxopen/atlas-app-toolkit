package gateway

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAnnotator(t *testing.T) {
	tests := map[string]metadata.MD{}
	tests[``] = metadata.MD{fieldPresenceMetaKey: nil}
	tests[`{}`] = metadata.MD{fieldPresenceMetaKey: nil}
	tests[`{`] = nil
	tests[`{"one": {}}`] = metadata.MD{fieldPresenceMetaKey: []string{"One"}}
	tests[`{"one":{"two":"a", "three":[]}, "four": 5}`] = metadata.MD{fieldPresenceMetaKey: []string{"Four$One.Two$One.Three"}}
	tests[`{"objects":[{"one": {"two":"a", "three":[]}, "four": 5}, {"one":{"two":"a", "three":[]}, "four": 5}]}`] = metadata.MD{fieldPresenceMetaKey: []string{"Four$One.Two$One.Three", "Four$One.Two$One.Three"}}
	tests[`{
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
}`] = metadata.MD{fieldPresenceMetaKey: []string{"Name$Burden.Duration$Burden.Weight$Burden.Breaks$Burden.Replacements.Hero.Name$Burden.Replacements.Hero.Duration$Burden.Replacements.Hero.Lineage.Mother$Burden.Replacements.Hero.Lineage.Father$Burden.Replacements.Mortals"}}

	for input, expect := range tests {
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
		if !isEqualFieldMasks(md[fieldPresenceMetaKey], expect[fieldPresenceMetaKey]) {
			t.Errorf("Did not produce expected metadata %+v, got %+v", expect, md)
		}

	}
}

func isEqualFieldMasks(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		fmt.Println("len(s1) != len(s2)", len(s1), len(s2))
		return false
	}

	for i := 0; i < len(s1); i++ {
		mask1, mask2 := strings.Split(s1[i], "$"), strings.Split(s2[i], "$")
		sort.Strings(mask1)
		sort.Strings(mask2)

		if !reflect.DeepEqual(mask1, mask2) {
			fmt.Println("!reflect.DeepEqual(mask1, mask2)")
			return false
		}
	}

	return true
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
			fieldPresenceMetaKey: []string{"one.two.three$one.four"},
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
	t.Run("sets FieldMask to empty if nil", func(t *testing.T) {
		md := runtime.ServerMetadata{
			HeaderMD: metadata.MD{
				fieldPresenceMetaKey: nil,
			},
		}
		ctx1 := runtime.NewServerMetadataContext(context.Background(), md)
		req := &dummyReq{}
		if err := interceptor(ctx1, "POST", req, nil, nil, dummyInvoker); err != nil {
			t.Fatal(err)
		}
		if req == nil {
			t.Fatal("For some reason it deleted the request object")
		}
		got, want := req.SomeFieldMaskField, &field_mask.FieldMask{}
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

type RequestWithFieldMask struct {
	FieldMask *field_mask.FieldMask
}

func TestOverrideFieldMaskOption(t *testing.T) {
	defaultInvoker := func(context.Context, string, interface{}, interface{}, *grpc.ClientConn, ...grpc.CallOption) error {
		return nil
	}

	f := func(ctx context.Context, req *RequestWithFieldMask, expected *field_mask.FieldMask, overrideEnabled bool) {
		interceptor := PresenceClientInterceptor()
		if overrideEnabled {
			interceptor = PresenceClientInterceptor(WithOverrideFieldMask)
		}

		if err := interceptor(ctx, "", req, nil, nil, defaultInvoker); err != nil {
			t.Logf("Unexpected error %s\n", err)
			t.Fail()
		}

		result := req.FieldMask

		if !isEqualFieldMasks(expected.Paths, result.Paths) {
			t.Logf("Unexpected result field mask, expect %+v, got %+v\n", expected.Paths, result.Paths)
			t.Fail()
		}

	}

	r1 := &RequestWithFieldMask{FieldMask: &field_mask.FieldMask{Paths: []string{"One"}}}
	f(context.Background(), r1, &field_mask.FieldMask{Paths: []string{"One"}}, true)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{fieldPresenceMetaKey: []string{"One"}})
	r2 := &RequestWithFieldMask{}
	f(ctx, r2, &field_mask.FieldMask{Paths: []string{"One"}}, true)

	ctx = metadata.NewIncomingContext(context.Background(), metadata.MD{fieldPresenceMetaKey: []string{"Two"}})
	r3 := &RequestWithFieldMask{FieldMask: &field_mask.FieldMask{Paths: []string{"One"}}}
	f(ctx, r3, &field_mask.FieldMask{Paths: []string{"Two"}}, true)

	ctx = metadata.NewIncomingContext(context.Background(), metadata.MD{fieldPresenceMetaKey: []string{"Two"}})
	r4 := &RequestWithFieldMask{FieldMask: &field_mask.FieldMask{Paths: []string{"One"}}}
	f(ctx, r4, &field_mask.FieldMask{Paths: []string{"One"}}, false)
}

type RequestWithMultiFieldMask struct {
	FieldMasks []*field_mask.FieldMask
}

func TestOverrideMultipleFieldMasksOption(t *testing.T) {
	defaultInvoker := func(context.Context, string, interface{}, interface{}, *grpc.ClientConn, ...grpc.CallOption) error {
		return nil
	}

	f := func(ctx context.Context, req *RequestWithMultiFieldMask, expected []*field_mask.FieldMask, overrideEnabled bool) {
		interceptor := PresenceClientInterceptor()
		if overrideEnabled {
			interceptor = PresenceClientInterceptor(WithOverrideFieldMask)
		}

		if err := interceptor(ctx, "", req, nil, nil, defaultInvoker); err != nil {
			t.Logf("Unexpected error %s\n", err)
			t.Fail()
		}

		result := req.FieldMasks
		if len(result) != len(expected) {
			t.Logf("Unexpected field masks expect %+v, got %+v", expected, result)
			t.Fail()
			return
		}

		for i := 0; i < len(expected); i++ {
			if !isEqualFieldMasks(expected[i].Paths, result[i].Paths) {
				t.Logf("Unexpected result field mask on index %d, expect %+v, got %+v\n", i, expected[0].Paths, result[0].Paths)
				t.Fail()
			}
		}

	}

	r1 := &RequestWithMultiFieldMask{FieldMasks: []*field_mask.FieldMask{{Paths: []string{"One"}}, {Paths: []string{"Two"}}}}
	f(context.Background(), r1, []*field_mask.FieldMask{{Paths: []string{"One"}}, {Paths: []string{"Two"}}}, true)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{fieldPresenceMetaKey: []string{"One", "Two"}})
	r2 := &RequestWithMultiFieldMask{}
	f(ctx, r2, []*field_mask.FieldMask{{Paths: []string{"One"}}, {Paths: []string{"Two"}}}, true)

	ctx = metadata.NewIncomingContext(context.Background(), metadata.MD{fieldPresenceMetaKey: []string{"Four", "Five"}})
	r3 := &RequestWithMultiFieldMask{FieldMasks: []*field_mask.FieldMask{{Paths: []string{"One"}}, {Paths: []string{"Two"}}}}
	f(ctx, r3, []*field_mask.FieldMask{{Paths: []string{"Four"}}, {Paths: []string{"Five"}}}, true)

	ctx = metadata.NewIncomingContext(context.Background(), metadata.MD{fieldPresenceMetaKey: []string{"Four", "Five"}})
	r4 := &RequestWithMultiFieldMask{FieldMasks: []*field_mask.FieldMask{{Paths: []string{"One"}}, {Paths: []string{"Two"}}}}
	f(ctx, r4, []*field_mask.FieldMask{{Paths: []string{"One"}}, {Paths: []string{"Two"}}}, false)
}
