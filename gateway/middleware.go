package gateway

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware if an user's request message
// defines any of collection operators.
//
// Returned middleware parse collection operators from gRPC metadata if
// they defined and stores in context.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		// handle panic
		defer func() {
			if perr := recover(); perr != nil {
				err = status.Errorf(codes.Internal, "collection operators interceptor: %s", perr)
				grpclog.Errorln(err)
				res, err = nil, err
			}
		}()

		// looking for Sorting
		sorting, err := Sorting(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid sorting operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if sorting != nil {
			ctx = NewSortingContext(ctx, sorting)
		}

		// looking for Filtering
		filtering, err := Filtering(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid filtering operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if filtering != nil {
			ctx = NewFilteringContext(ctx, filtering)
		}

		// looking for ClientDrivenPagination
		pagination, err := Pagination(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid pagination operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if pagination != nil {
			ctx = NewPaginationContext(ctx, pagination)
		}

		// looking for FieldSelection
		fieldSelection := FieldSelection(ctx)
		if fieldSelection != nil {
			ctx = NewFieldSelectionContext(ctx, fieldSelection)
		}

		res, err = handler(ctx, req)
		return
	}
}
