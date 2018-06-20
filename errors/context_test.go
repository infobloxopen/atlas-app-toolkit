package errors

import (
	"testing"

	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errfields"
)

func initCtx() context.Context {
	return NewContext(context.Background(), InitContainer())
}

func TestNewContext(t *testing.T) {
	ctx := context.Background()

	// Test empty context.

	if FromContext(ctx) != nil {
		t.Errorf("Invalid FromContext result, expected nil, got %v", FromContext(ctx))
	}

	// Test save in context.

	ctx = NewContext(ctx, InitContainer())
	checkContainer(t, InitContainer(), FromContext(ctx))
}

func TestContextDetail(t *testing.T) {
	ctx := initCtx()

	// Test Detail in context.

	checkContainer(t, Detail(ctx, codes.InvalidArgument, "target", "<detail %d>", 1),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<detail 1>")},
			fields:     nil,
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)

	// Append another detail to context.

	checkContainer(t, Detail(ctx, codes.AlreadyExists, "target", "<detail 2>"),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<detail 1>"),
				errdetails.Newf(codes.AlreadyExists, "target", "<detail 2>")},
			fields:     nil,
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)
}

func TestContextDetails(t *testing.T) {
	ctx := initCtx()

	checkContainer(t,
		Details(
			ctx,
			errdetails.Newf(codes.InvalidArgument, "target", "<detail 1>"),
			errdetails.Newf(codes.AlreadyExists, "target", "<detail 2>"),
		),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<detail 1>"),
				errdetails.Newf(codes.AlreadyExists, "target", "<detail 2>")},
			fields:     nil,
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)
}

func TestContextField(t *testing.T) {
	ctx := initCtx()

	// Test fields in context.

	checkContainer(t,
		Field(ctx, "target", "<field %d>", 1),
		&Container{
			details: nil,
			fields: &errfields.FieldInfo{
				Fields: map[string]*errfields.StringListValue{
					"target": &errfields.StringListValue{
						Values: []string{"<field 1>"},
					},
				},
			},
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)

	// Append another field.

	checkContainer(t,
		Field(ctx, "target", "<field 2>"),
		&Container{
			details: nil,
			fields: &errfields.FieldInfo{
				Fields: map[string]*errfields.StringListValue{
					"target": &errfields.StringListValue{
						Values: []string{"<field 1>", "<field 2>"},
					},
				},
			},
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)

}

func TestContextFields(t *testing.T) {
	ctx := initCtx()

	// Test fields in context.

	checkContainer(t,
		Fields(ctx, map[string][]string{"target": []string{"<field 1>"}}),
		&Container{
			details: nil,
			fields: &errfields.FieldInfo{
				Fields: map[string]*errfields.StringListValue{
					"target": &errfields.StringListValue{
						Values: []string{"<field 1>"},
					},
				},
			},
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)

	// Append another bunch of fields.

	checkContainer(t,
		Fields(ctx, map[string][]string{"target": []string{"<field 2>"}, "target2": []string{"<field 3>"}}),
		&Container{
			details: nil,
			fields: &errfields.FieldInfo{
				Fields: map[string]*errfields.StringListValue{
					"target": &errfields.StringListValue{
						Values: []string{"<field 1>", "<field 2>"},
					},
					"target2": &errfields.StringListValue{
						Values: []string{"<field 3>"},
					},
				},
			},
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)
}

func TestContextSet(t *testing.T) {
	ctx := initCtx()

	// Check IfSet returning nil.

	if v := IfSet(ctx, codes.InvalidArgument, "<general error %d>", 1); v != nil {
		t.Errorf("Invalid IfSet result, expected nil, got %v", v)
	}

	// Check Set returning container with errCode/errMessage and details set.

	checkContainer(t,
		Set(ctx, "target", codes.InvalidArgument, "<error %d>", 1),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<error 1>")},
			fields:     nil,
			errCode:    codes.InvalidArgument,
			errMessage: "<error 1>",
			errSet:     true,
		},
	)

	// Check IfSet overwriting errCode/errMessage.

	checkContainer(t,
		IfSet(ctx, codes.InvalidArgument, "<general error %d>", 1).(*Container),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<error 1>")},
			fields:     nil,
			errCode:    codes.InvalidArgument,
			errMessage: "<general error 1>",
			errSet:     true,
		},
	)
}

func TestContextError(t *testing.T) {
	ctx := initCtx()

	if v := Error(ctx); v != nil {
		t.Errorf("Unexpected Error result: expected nil, got %v", v)
	}

	Detail(ctx, codes.InvalidArgument, "target", "<error 1>")

	checkContainer(
		t,
		Error(ctx).(*Container),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<error 1>")},
			fields:     nil,
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     true,
		},
	)

	// Test that new resets isSet flag and turns error into nil.

	New(ctx, codes.InvalidArgument, "target", "<error 1>")

	if v := Error(ctx); v != nil {
		t.Errorf("Unexpected Error result: expected nil, got %v", v)
	}
}

func TestContextNew(t *testing.T) {
	ctx := initCtx()

	checkContainer(t,
		Set(ctx, "target", codes.InvalidArgument, "<error %d>", 1),
		&Container{
			details: []*errdetails.TargetInfo{
				errdetails.Newf(codes.InvalidArgument, "target", "<error 1>")},
			fields:     nil,
			errCode:    codes.InvalidArgument,
			errMessage: "<error 1>",
			errSet:     true,
		},
	)

	checkContainer(t,
		New(ctx, codes.InvalidArgument, "<error %d>", 2),
		&Container{
			details:    nil,
			fields:     nil,
			errCode:    codes.InvalidArgument,
			errMessage: "<error 2>",
			errSet:     false,
		},
	)
}

func TestMap(t *testing.T) {
	c := InitContainer()
	c.AddMapping(
		NewMapping(CondEq("err1"), NewContainer(codes.InvalidArgument, "<error 1>")),
		NewMapping(fmt.Errorf("err2"), NewContainer(codes.InvalidArgument, "<error 2>")),
	)

	ctx := NewContext(context.Background(), c)

	checkContainer(t,
		Map(ctx, fmt.Errorf("err4")).(*Container),
		InitContainer())

	checkContainer(t,
		Map(ctx, fmt.Errorf("err1")).(*Container),
		NewContainer(codes.InvalidArgument, "<error 1>"))

	checkContainer(t,
		Map(ctx, fmt.Errorf("err2")).(*Container),
		NewContainer(codes.InvalidArgument, "<error 2>"))
}
