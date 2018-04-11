package gw

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func TestHeader(t *testing.T) {
	imd := metadata.Pairs("key1", "val1")
	omd := metadata.Pairs("key2", "val2", "grpcgateway-key2", "val2")

	ictx := metadata.NewIncomingContext(context.Background(), imd)
	ctx := metadata.NewOutgoingContext(ictx, omd)

	if v, ok := Header(ctx, "key1"); !ok {
		t.Error("failed to get 'key1'")
	} else if v != "val1" {
		t.Errorf("invalid value of 'key1': %s", v)
	}

	if v, ok := Header(ctx, "key2"); !ok {
		t.Error("failed to get 'key2'")
	} else if v != "val2" {
		t.Errorf("invalid value of 'key2': %s", v)
	}
}

func TestHeaderN(t *testing.T) {
	imd := metadata.Pairs("key1", "val1")
	omd := metadata.Pairs("key2", "val2", "grpcgateway-key2", "val2")

	ictx := metadata.NewIncomingContext(context.Background(), imd)
	ctx := metadata.NewOutgoingContext(ictx, omd)

	if v, ok := HeaderN(ctx, "key1", -1); !ok {
		t.Error("failed to get 'key1'")
	} else if len(v) != 1 || v[0] != "val1" {
		t.Errorf("invalid value of 'key1': %s", v)
	}

	if v, ok := HeaderN(ctx, "key2", 2); !ok {
		t.Error("failed to get 'key2'")
	} else if len(v) != 2 || v[0] != "val2" || v[1] != "val2" {
		t.Errorf("invalid value of 'key2': %s", v)
	}

	if v, ok := HeaderN(ctx, "key1", 0); ok || v != nil {
		t.Errorf("invalid result with n==0: %s, %s", v, ok)
	}

	if v, ok := HeaderN(ctx, "key1", 10); ok || v != nil {
		t.Errorf("invalid result with n>len(md): %s, %s", v, ok)
	}
}

func TestPrefixOutgoingHeaderMatcher(t *testing.T) {
	key := "Content-Type"
	v, ok := PrefixOutgoingHeaderMatcher(key)
	if !ok || v != runtime.MetadataHeaderPrefix+key {
		t.Errorf("header %s is not matched: %s, %s", key, v, ok)
	}
}
