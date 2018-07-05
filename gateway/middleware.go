package gateway

import (
	"context"
	"fmt"
	"reflect"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware if an user's request message
// defines any of collection operators.
//
// Returned middleware populates collection operators from gRPC metadata if
// they defined in a request message.
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

		if req == nil {
			grpclog.Warningf("collection operator interceptor: empty request %+v", req)
			return handler(ctx, req)
		}

		res, err = handler(ctx, req)
		if err != nil {
			return res, err
		}

		// looking for op.PageInfo
		page := new(query.PageInfo)
		if err := unsetOp(res, page); err != nil {
			grpclog.Errorf("collection operator interceptor: failed to set page info - %s", err)
		}

		if err := SetPageInfo(ctx, page); err != nil {
			grpclog.Errorf("collection operator interceptor: failed to set page info - %s", err)
			return nil, err
		}

		return
	}
}

func SetCollectionOps(req, op interface{}) error {
	reqval := reflect.ValueOf(req)

	if reqval.Kind() != reflect.Ptr {
		return fmt.Errorf("request is not a pointer - %s", reqval.Kind())
	}

	reqval = reqval.Elem()

	if reqval.Kind() != reflect.Struct {
		return fmt.Errorf("request value is not a struct - %s", reqval.Kind())
	}

	for i := 0; i < reqval.NumField(); i++ {
		f := reqval.FieldByIndex([]int{i})

		if f.Type() != reflect.TypeOf(op) {
			continue
		}

		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("operation field %+v in request %+v is invalid or cannot be set", op, req)
		}

		if vop := reflect.ValueOf(op); vop.IsValid() {
			f.Set(vop)
		}

	}

	return nil
}

func GetCollectionOp(res, op interface{}) error {
	return getAndUnsetOp(res, op, false)
}

func unsetOp(res, op interface{}) error {
	return getAndUnsetOp(res, op, true)
}

func getAndUnsetOp(res, op interface{}, unset bool) error {
	resval := reflect.ValueOf(res)
	if resval.Kind() != reflect.Ptr {
		return fmt.Errorf("response is not a pointer - %s", resval.Kind())
	}

	resval = resval.Elem()
	if resval.Kind() != reflect.Struct {
		return fmt.Errorf("response value is not a struct - %s", resval.Kind())
	}

	opval := reflect.ValueOf(op)
	if opval.Kind() != reflect.Ptr {
		return fmt.Errorf("operator is not a pointer - %s", opval.Kind())
	}

	for i := 0; i < resval.NumField(); i++ {
		f := resval.FieldByIndex([]int{i})

		if f.Type() != opval.Type() {
			continue
		}

		if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.Ptr {
			return fmt.Errorf("operation field %T in response %+v is invalid or cannot be set", op, res)
		}

		if o := opval.Elem(); o.IsValid() && o.CanSet() && f.Elem().IsValid() {
			o.Set(f.Elem())
		}
		if unset {
			f.Set(reflect.Zero(f.Type()))
		}
	}
	return nil
}
