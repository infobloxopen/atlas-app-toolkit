package errors

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestMapper(t *testing.T) {
	var err error
	var ctx context.Context

	ctx = context.Background()
	ctx = NewContext(ctx, NewContainer())

	FromContext(ctx).WithMapping(
		func(c *Container, err error) (error, bool) {
			if err.Error() == "<this is unmapped error>" {
				return c.New(
					codes.InvalidArgument,
					"Well, Now it's mapped.",
				).WithDetails(codes.InvalidArgument, "target", "Detail"), true
			}
			return nil, false
		},
	)

	err = FromContext(ctx).Map(fmt.Errorf("<this is unmapped error>"))
	checkErr(t, err, codes.InvalidArgument, "Well, Now it's mapped.")

	err = FromContext(ctx).Map(fmt.Errorf("Blah"))
	checkErr(t, err, codes.Unknown, "Unknown")
}
