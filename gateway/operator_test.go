package gateway

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

func TestSorting(t *testing.T) {
	// sort parameters is not specified
	req, err := http.NewRequest(http.MethodGet, "http://app.com?someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), req)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	s, err := Sorting(ctx)
	if err != nil || s != nil {
		t.Fatalf("invalid error: %s, %s - expected: nil, nil", s, err)
	}

	// invalid sort parameters
	req, err = http.NewRequest(http.MethodGet, "http://app.com?_order_by=name dasc, age desc&someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), req)
	ctx = metadata.NewIncomingContext(context.Background(), md)

	_, err = Sorting(ctx)
	if err == nil {
		t.Fatal("no error returned")
	}
	if s, ok := status.FromError(err); !ok {
		t.Fatal("no status error retunred")
	} else if s.Code() != codes.InvalidArgument {
		t.Errorf("invalid status code: %s - expected: %s", s.Code(), codes.InvalidArgument)
	}

	// valid sort parameters
	req, err = http.NewRequest(http.MethodGet, "http://app.com?_order_by=name asc, age desc&someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), req)
	ctx = metadata.NewIncomingContext(context.Background(), md)

	s, err = Sorting(ctx)
	if err != nil {
		t.Fatalf("failed to extract sorting parameters from context: %s", err)
	}

	if len(s.GetCriterias()) != 2 {
		t.Fatalf("invalid number of sort criterias: %d - expected: 2", len(s.GetCriterias()))
	}
	if c := s.GetCriterias(); c[0].GoString() != "name ASC" || c[0].Tag != "name" || c[0].Order != query.SortCriteria_ASC {
		t.Errorf("invalid sort criteria: %v - expected: %v", c[0], query.SortCriteria{"name", query.SortCriteria_ASC})
	}
	if c := s.GetCriterias(); c[1].GoString() != "age DESC" || c[1].Tag != "age" || c[1].Order != query.SortCriteria_DESC {
		t.Errorf("invalid sort criteria: %v - expected: %v", c[1], query.SortCriteria{"age", query.SortCriteria_DESC})
	}
}

func TestPagination(t *testing.T) {
	// valid pagination testRequest
	req, err := http.NewRequest(http.MethodGet, "http://app.com?_limit=20&_offset=10&_page_token=ptoken", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), req)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	page, err := Pagination(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if page.GetLimit() != 20 || page.GetOffset() != 10 || page.GetPageToken() != "ptoken" {
		t.Errorf("invalid pagination: %s - expected: %s", page, &query.Pagination{Limit: 20, Offset: 10, PageToken: "ptoken"})
	}

	// invalid pagination testRequest
	req, err = http.NewRequest(http.MethodGet, "http://app.com?_limit=twenty&_offset=10", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), req)
	ctx = metadata.NewIncomingContext(context.Background(), md)

	_, err = Pagination(ctx)
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

func TestNewPaginationContext(t *testing.T) {
	ctx := context.Background()

	p, err := Pagination(ctx)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if p.GetOffset() != 0 || p.DefaultLimit() != query.DefaultLimit {
		t.Errorf("invalid pagination instance %v", p)
	}

	// change pagination
	p = &query.Pagination{}
	p.Offset, p.Limit = 19, 86
	ctx = NewPaginationContext(ctx, p)

	p, err = Pagination(ctx)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	if p.Offset != 19 || p.Limit != 86 {
		t.Errorf("invalid pagination instance %v", p)
	}
}
