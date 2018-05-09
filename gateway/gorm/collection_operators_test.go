package gorm

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/jinzhu/gorm"
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

func TestApplyCollectionOperators(t *testing.T) {
	gormDB, mock := setUp(t)

	req, err := http.NewRequest("GET", "http://test.com?_fields=name&_filter=age<=25&_order_by=age desc&_limit=2&_offset=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(fixedFullRe("SELECT name FROM \"people\" WHERE ((age <= $1)) ORDER BY age desc LIMIT 2 OFFSET 1")).WithArgs(25.0)

	md := gateway.MetadataAnnotator(nil, req)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	gormDB, err = ApplyCollectionOperators(gormDB, ctx)
	if err != nil {
		t.Fatal(err)
	}

	var actual []Person
	gormDB.Find(&actual)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
