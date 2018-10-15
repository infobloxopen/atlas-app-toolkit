package gateway

import (
	"context"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware if an user's request message
// defines any of collection operators.
//
// Returned middleware populates collection operators from gRPC metadata if
// they defined in a request message.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {

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
	_, err := getAndUnsetOp(res, op, false)
	return err
}

func unsetOp(res, op interface{}) error {
	_, err := getAndUnsetOp(res, op, true)
	return err
}

func getAndUnsetOp(res, op interface{}, unset bool) (fieldName string, err error) {
	resval := reflect.ValueOf(res)
	if resval.Kind() != reflect.Ptr {
		return "", fmt.Errorf("response is not a pointer - %s", resval.Kind())
	}

	resval = resval.Elem()
	if resval.Kind() != reflect.Struct {
		return "", fmt.Errorf("response value is not a struct - %s", resval.Kind())
	}

	opval := reflect.ValueOf(op)
	if opval.Kind() != reflect.Ptr {
		return "", fmt.Errorf("operator is not a pointer - %s", opval.Kind())
	}

	for i := 0; i < resval.NumField(); i++ {
		f := resval.FieldByIndex([]int{i})

		if f.Type() != opval.Type() {
			continue
		}

		if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.Ptr {
			return "", fmt.Errorf("operation field %T in response %+v is invalid or cannot be set", op, res)
		}

		if o := opval.Elem(); o.IsValid() && o.CanSet() && f.Elem().IsValid() {
			o.Set(f.Elem())
		}
		fieldName = reflect.TypeOf(res).Elem().Field(i).Name
		if unset {
			f.Set(reflect.Zero(f.Type()))
		}

	}
	return fieldName, nil
}

func GetPageInfo(resp proto.Message) (fieldName string, pg *query.PageInfo, err error) {
	pg = new(query.PageInfo)
	fieldName, err = getAndUnsetOp(resp, pg, false)
	if fieldName == "" {
		pg = nil
	}
	return
}

func GetFiltering(req proto.Message) (fieldName string, f *query.Filtering, err error) {
	f = new(query.Filtering)
	fieldName, err = getAndUnsetOp(req, f, false)
	if fieldName == "" {
		f = nil
	}
	return
}

func GetSorting(req proto.Message) (fieldName string, s *query.Sorting, err error) {
	s = new(query.Sorting)
	fieldName, err = getAndUnsetOp(req, s, false)
	if fieldName == "" {
		s = nil
	}
	return
}

func GetPagination(req proto.Message) (fieldName string, p *query.Pagination, err error) {
	p = new(query.Pagination)
	fieldName, err = getAndUnsetOp(req, p, false)
	if fieldName == "" {
		p = nil
	}
	return
}

func GetFieldSelection(req proto.Message) (fieldName string, fs *query.FieldSelection, err error) {
	fs = new(query.FieldSelection)
	fieldName, err = getAndUnsetOp(req, fs, false)
	if fieldName == "" {
		fs = nil
	}
	return
}
