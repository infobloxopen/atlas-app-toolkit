package gateway

import (
	"context"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

var (
	sortingCollection   = "sorting_collection"
	filteringCollection = "filtering_collection"
	pagingCollection    = "paging_collection"
	fieldCollection     = "field_collection"
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
			ctx = CtxSetSorting(ctx, sorting)
		}

		// looking for Filtering
		filtering, err := Filtering(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid filtering operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if filtering != nil {
			ctx = CtxSetFiltering(ctx, filtering)
		}

		// looking for ClientDrivenPagination
		pagination, err := Pagination(ctx)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "collection operator interceptor: invalid pagination operator - %s", err)
			grpclog.Errorln(err)
			return nil, err
		}
		if pagination != nil {
			ctx = CtxSetPagination(ctx, pagination)
		}

		// looking for FieldSelection
		fieldSelection := FieldSelection(ctx)
		if fieldSelection != nil {
			ctx = CtxSetFieldSelection(ctx, fieldSelection)
		}

		res, err = handler(ctx, req)
		return
	}
}

//
// Collection operators setters
//
func CtxSetFiltering(ctx context.Context, f *query.Filtering) context.Context {
	return context.WithValue(ctx, filteringCollection, f)
}

func CtxSetSorting(ctx context.Context, s *query.Sorting) context.Context {
	return context.WithValue(ctx, sortingCollection, s)
}

func CtxSetPagination(ctx context.Context, p *query.Pagination) context.Context {
	return context.WithValue(ctx, pagingCollection, p)
}

func CtxSetFieldSelection(ctx context.Context, f *query.FieldSelection) context.Context {
	return context.WithValue(ctx, fieldCollection, f)
}

//
// Collection operators getters
//
func CtxGetFiltering(ctx context.Context) *query.Filtering {
	if v := ctx.Value(filteringCollection); v != nil {
		if f, ok := v.(*query.Filtering); ok {
			return f
		}
	}
	return nil
}

func CtxGetSorting(ctx context.Context) *query.Sorting {
	if v := ctx.Value(sortingCollection); v != nil {
		if f, ok := v.(*query.Sorting); ok {
			return f
		}
	}
	return nil
}

func CtxGetPagination(ctx context.Context) *query.Pagination {
	if v := ctx.Value(pagingCollection); v != nil {
		if f, ok := v.(*query.Pagination); ok {
			return f
		}
	}
	return nil
}

func CtxGetFieldSelection(ctx context.Context) *query.FieldSelection {
	if v := ctx.Value(fieldCollection); v != nil {
		if f, ok := v.(*query.FieldSelection); ok {
			return f
		}
	}
	return nil
}
