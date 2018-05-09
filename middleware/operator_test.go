package middleware

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/collections"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
)

type request struct {
	Sorting    *collections.Sorting
	Pagination *collections.Pagination
}

type response struct {
	PageInfo *collections.PageInfo
}

func TestWithCollectionOperatorSorting(t *testing.T) {
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_order_by=name asc, age desc", nil)
	if err != nil {
		t.Fatalf("failed to build new http request: %s", err)
	}
	md := gateway.MetadataAnnotator(context.Background(), hreq)

	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := &request{Sorting: nil}
	interceptor := WithCollectionOperator()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		msg := req.(*request)
		criteria := msg.Sorting.GetCriterias()
		if len(criteria) != 2 {
			t.Fatalf("invalid number of sort criteria: %d - expected: %d", len(criteria), 2)
		}

		if c := criteria[0]; c.Tag != "name" || c.Order != collections.SortCriteria_ASC {
			t.Errorf("invalid sort criteria: %v - expected: %v", c, collections.SortCriteria{"name", collections.SortCriteria_ASC})
		}

		if c := criteria[1]; c.Tag != "age" || c.Order != collections.SortCriteria_DESC {
			t.Errorf("invalid sort criteria: %v - expected: %v", c, collections.SortCriteria{"age", collections.SortCriteria_DESC})
		}

		return &response{}, nil
	}

	_, err = interceptor(ctx, req, nil, handler)
	if err != nil {
		t.Fatalf("failed to attach sorting to request: %s", err)
	}
}

func TestWithCollectionOperatorPagination(t *testing.T) {
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_limit=10&_offset=20", nil)
	if err != nil {
		t.Fatalf("failed to build new http request: %s", err)
	}

	md := gateway.MetadataAnnotator(context.Background(), hreq)

	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := &request{Pagination: nil}
	interceptor := WithCollectionOperator()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		msg := req.(*request)
		page := msg.Pagination

		if page.GetLimit() != 10 {
			t.Errorf("invalid pagination limit: %d - expected: 10", page.GetLimit())
		}
		if page.GetOffset() != 20 {
			t.Errorf("invalid pagination offset: %d - expected: 20", page.GetOffset())
		}

		return &response{}, nil
	}

	_, err = interceptor(ctx, req, nil, handler)
	if err != nil {
		t.Fatalf("failed to attach sorting to request: %s", err)
	}
}

func TestUnsetOp(t *testing.T) {
	page := new(collections.PageInfo)
	res := &response{PageInfo: &collections.PageInfo{Offset: 30, Size: 10}}

	if err := unsetOp(res, page); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if page.GetOffset() != 30 {
		t.Errorf("invalid repsponse offset: %d - expected: 30", page.GetOffset())
	}
	if page.GetSize() != 10 {
		t.Errorf("invalid repsponse size: %d - expected: 10", page.GetSize())
	}

	// nil operator
	err := unsetOp(res, nil)
	if err == nil {
		t.Fatalf("unexpected non error result - expected: %s", "operator is not a pointer - invalid")
	}
	if err.Error() != "operator is not a pointer - invalid" {
		t.Errorf("invalid error: %s - expected: %s", err, "operator is not a pointer - invalid")
	}

	// nil response
	err = unsetOp(nil, nil)
	if err == nil {
		t.Fatalf("unexpected non error result - expected: %s", "response is not a pointer - invalid")
	}
	if err.Error() != "response is not a pointer - invalid" {
		t.Errorf("invalid error: %s - expected: %s", err, "response is not a pointer - invalid")
	}

	// non struct response
	var i int
	err = unsetOp(&i, nil)
	if err == nil {
		t.Fatalf("unexpected non error result - expected: %s", "response value is not a struct - int")
	}
	if err.Error() != "response value is not a struct - int" {
		t.Errorf("invalid error: %s - expected: %s", err, "response value is not a struct - int")
	}
}
