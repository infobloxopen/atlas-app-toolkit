package gateway

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type testRequest struct {
	Sorting    *query.Sorting
	Pagination *query.Pagination
}

type testResponse struct {
	PageInfo *query.PageInfo
}

func TestWithCollectionOperatorSorting(t *testing.T) {
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_order_by=name asc, age desc", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}
	md := MetadataAnnotator(context.Background(), hreq)

	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := &testRequest{Sorting: nil}
	interceptor := UnaryServerInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		criteria := CtxGetSorting(ctx).GetCriterias()
		if len(criteria) != 2 {
			t.Fatalf("invalid number of sort criteria: %d - expected: %d", len(criteria), 2)
		}

		if c := criteria[0]; c.Tag != "name" || c.Order != query.SortCriteria_ASC {
			t.Errorf("invalid sort criteria: %v - expected: %v", c, query.SortCriteria{"name", query.SortCriteria_ASC})
		}

		if c := criteria[1]; c.Tag != "age" || c.Order != query.SortCriteria_DESC {
			t.Errorf("invalid sort criteria: %v - expected: %v", c, query.SortCriteria{"age", query.SortCriteria_DESC})
		}

		return &testResponse{}, nil
	}

	_, err = interceptor(ctx, req, nil, handler)
	if err != nil {
		t.Fatalf("failed to attach sorting to testRequest: %s", err)
	}
}

func TestWithCollectionOperatorPagination(t *testing.T) {
	hreq, err := http.NewRequest(http.MethodGet, "http://app.com?_limit=10&_offset=20", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), hreq)

	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := &testRequest{Pagination: nil}
	interceptor := UnaryServerInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		page := CtxGetPagination(ctx)

		if page.GetLimit() != 10 {
			t.Errorf("invalid pagination limit: %d - expected: 10", page.GetLimit())
		}
		if page.GetOffset() != 20 {
			t.Errorf("invalid pagination offset: %d - expected: 20", page.GetOffset())
		}

		return &testResponse{}, nil
	}

	_, err = interceptor(ctx, req, nil, handler)
	if err != nil {
		t.Fatalf("failed to attach sorting to testRequest: %s", err)
	}
}
