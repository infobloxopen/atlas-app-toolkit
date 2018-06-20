package errors

import (
	"testing"

	"context"

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
