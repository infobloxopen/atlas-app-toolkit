package errors

import (
	"context"
	"google.golang.org/grpc/codes"
	"testing"
)

// TestContext function checks the NewContext and FromContext functions.
func TestContext(t *testing.T) {
	var v *Container
	var ctx context.Context

	// Check empty context.
	ctx = context.Background()
	v = FromContext(ctx)
	checkErr(t, v, codes.Unknown, "Unknown")

	// Init error and save it in context.
	ctx = NewContext(ctx, NewContainer().New(codes.Internal, "Internal Server Error"))
	v = FromContext(ctx)
	checkErr(t, v, codes.Internal, "Internal Server Error")

	// Update error in context.
	FromContext(ctx).New(codes.InvalidArgument, "Invalid argument A")
	v = FromContext(ctx)
	checkErr(t, v, codes.InvalidArgument, "Invalid argument A")
}

func TestDetails(t *testing.T) {}

func TestFields(t *testing.T) {}
