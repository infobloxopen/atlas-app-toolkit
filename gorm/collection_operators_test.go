package gorm

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Person struct {
	ID   int64
	Name string
	Age  int
}

func fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

func setUp(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	var gormDB *gorm.DB
	gormDB, err = gorm.Open("postgres", db)
	if err != nil {
		t.Fatal(err)
	}

	return gormDB, mock
}

type testRequest struct {
	Sorting        *query.Sorting
	Pagination     *query.Pagination
	Filtering      *query.Filtering
	FieldSelection *query.FieldSelection
}

type testResponse struct {
	PageInfo *query.PageInfo
}

func TestApplyCollectionOperators(t *testing.T) {

	req, err := http.NewRequest("GET", "http://test.com?_fields=name&_filter=age<=25&_order_by=age desc&_limit=2&_offset=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	md := gateway.MetadataAnnotator(nil, req)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		gormDB, mock := setUp(t)

		rq := req.(*testRequest)
		gormDB, err = ApplyCollectionOperators(gormDB, rq.Filtering, rq.Sorting, rq.Pagination, rq.FieldSelection)
		if err != nil {
			t.Fatal(err)
		}

		mock.ExpectQuery(fixedFullRe("SELECT name FROM \"people\" WHERE ((age <= $1)) ORDER BY age desc LIMIT 2 OFFSET 1")).WithArgs(25.0)

		var actual []Person
		gormDB.Find(&actual)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
		return nil
	}

	err = gateway.ClientUnaryInterceptor(ctx, req.Method, &testRequest{}, &testRequest{}, nil, invoker)
	if err != nil {
		t.Fatal("no error returned")
	}
}
