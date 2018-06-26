package gateway

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type testRequest struct {
	Sorting        *query.Sorting
	Pagination     *query.Pagination
	Filtering      *query.Filtering
	FieldSelection *query.FieldSelection
}

type testResponse struct {
	PageInfo *query.PageInfo
}

func TestSorting(t *testing.T) {
	// sort parameters is not specified
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), hreq)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &testRequest{}
	repl := &testResponse{}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		tstReq := req.(*testRequest)
		if tstReq.Sorting != nil {
			t.Fatalf("invalid error: %s, %s - expected: nil, nil", tstReq.Sorting, err)
		}
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err != nil {
		t.Fatalf("invalid error: %s, for CollectionOperationsInterceptor", err)
	}

	// invalid sort parameters
	hreq, err = http.NewRequest(http.MethodGet, "http://app.com?_order_by=name dasc, age desc&someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), hreq)
	ctx = metadata.NewIncomingContext(context.Background(), md)

	invoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err == nil {
		t.Fatal("no error returned")
	}
	if s, ok := status.FromError(err); !ok {
		t.Fatal("no status error retunred")
	} else if s.Code() != codes.InvalidArgument {
		t.Errorf("invalid status code: %s - expected: %s", s.Code(), codes.InvalidArgument)
	}

	// valid sort parameters
	hreq, err = http.NewRequest(http.MethodGet, "http://app.com?_order_by=name asc, age desc&someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), hreq)
	ctx = metadata.NewIncomingContext(context.Background(), md)

	invoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		tstReq := req.(*testRequest)
		if tstReq.Sorting == nil {
			t.Fatalf("invalid error: %s, %s - expected: nil, nil", tstReq.Sorting, err)
		}
		s := tstReq.Sorting
		if len(s.GetCriterias()) != 2 {
			t.Fatalf("invalid number of sort criterias: %d - expected: 2", len(s.GetCriterias()))
		}
		if c := s.GetCriterias(); c[0].GoString() != "name ASC" || c[0].Tag != "name" || c[0].Order != query.SortCriteria_ASC {
			t.Errorf("invalid sort criteria: %v - expected: %v", c[0], query.SortCriteria{"name", query.SortCriteria_ASC})
		}
		if c := s.GetCriterias(); c[1].GoString() != "age DESC" || c[1].Tag != "age" || c[1].Order != query.SortCriteria_DESC {
			t.Errorf("invalid sort criteria: %v - expected: %v", c[1], query.SortCriteria{"age", query.SortCriteria_DESC})
		}
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err != nil {
		t.Fatalf("failed to extract sorting parameters from context: %s", err)
	}

}

func TestPagination(t *testing.T) {
	// valid pagination testRequest
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_limit=20&_offset=10&_page_token=ptoken", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), hreq)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &testRequest{}
	repl := &testResponse{}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		tstReq := req.(*testRequest)
		if tstReq.Pagination == nil {
			t.Fatalf("invalid error: %s, %s - expected: nil, nil", tstReq.Pagination, err)
		}
		page := tstReq.Pagination
		if page.GetLimit() != 20 || page.GetOffset() != 10 || page.GetPageToken() != "ptoken" {
			t.Errorf("invalid pagination: %s - expected: %s", page, &query.Pagination{Limit: 20, Offset: 10, PageToken: "ptoken"})
		}
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err != nil {
		t.Fatalf("invalid error: %s, for CollectionOperationsInterceptor", err)
	}

	// invalid pagination testRequest
	hreq, err = http.NewRequest(http.MethodGet, "http://app.com?_limit=twenty&_offset=10", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), hreq)
	ctx = metadata.NewIncomingContext(context.Background(), md)
	invoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err == nil {
		t.Fatalf("unexpected nil error")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("unexpected non status error: %v", s)
	}
	if s.Code() != codes.InvalidArgument {
		t.Errorf("invalid status error code: %d", s.Code())
	}
}

func TestFieldSelection(t *testing.T) {
	// valid pagination testRequest
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_fields=name,address.street&someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), hreq)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &testRequest{}
	repl := &testResponse{}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		tstReq := req.(*testRequest)
		if tstReq.FieldSelection == nil {
			t.Fatalf("invalid error: %s, %s - expected: nil, nil", tstReq.FieldSelection, err)
		}
		expected := &query.FieldSelection{Fields: query.FieldSelectionMap{"name": &query.Field{Name: "name"}, "address": &query.Field{Name: "address", Subs: query.FieldSelectionMap{"street": &query.Field{Name: "street"}}}}}
		if !reflect.DeepEqual(tstReq.FieldSelection, expected) {
			t.Errorf("Unexpected result %v while expecting %v", tstReq.FieldSelection, expected)
		}
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err != nil {
		t.Fatalf("invalid error: %s, for CollectionOperationsInterceptor", err)
	}
}

func TestFilitering(t *testing.T) {
	// valid pagination testRequest
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_filter=(field1!=\"abc\" and field2==\"zxc\") and (field3 >= 7 or field4 < 9)", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), hreq)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &testRequest{}
	repl := &testResponse{}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		tstReq := req.(*testRequest)
		if tstReq.Filtering == nil {
			t.Fatalf("invalid error: %s, %s - expected: nil, nil", tstReq.Filtering, err)
		}
		expected := &query.Filtering{}

		a := &query.StringCondition{FieldPath: []string{"field1"}, Value: "abc", Type: query.StringCondition_EQ, IsNegative: true}
		b := &query.StringCondition{FieldPath: []string{"field2"}, Value: "zxc", Type: query.StringCondition_EQ, IsNegative: false}

		c := &query.NumberCondition{FieldPath: []string{"field3"}, Value: 7.0, Type: query.NumberCondition_GE, IsNegative: false}
		d := &query.NumberCondition{FieldPath: []string{"field4"}, Value: 9.0, Type: query.NumberCondition_LT, IsNegative: false}

		ab := &query.LogicalOperator{Type: query.LogicalOperator_AND, IsNegative: false}
		ab.SetLeft(a)
		ab.SetRight(b)

		cd := &query.LogicalOperator{Type: query.LogicalOperator_OR, IsNegative: false}
		cd.SetLeft(c)
		cd.SetRight(d)

		abcd := &query.LogicalOperator{Type: query.LogicalOperator_AND, IsNegative: false}
		abcd.SetLeft(ab)
		abcd.SetRight(cd)

		expected.SetRoot(abcd)

		if !reflect.DeepEqual(tstReq.Filtering, expected) {
			t.Errorf("Unexpected result %v while expecting %v", tstReq.Filtering, expected)
		}
		return nil
	}

	err = ClientUnaryInterceptor(ctx, hreq.Method, req, repl, nil, invoker)
	if err != nil {
		t.Fatalf("invalid error: %s, for CollectionOperationsInterceptor", err)
	}
}
