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
	Id        int64
	Name      string
	Age       int
	SubPerson SubPerson `gorm:"foreignkey:PersonId;association_foreignkey:Id"`
}

type SubPerson struct {
	Id       int64
	Name     string
	PersonId int64
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

	req, err := http.NewRequest("GET", "http://test.com?_fields=id,name,sub_person&_filter=age<=25 and sub_person.name=='Mike'&_order_by=age,sub_person.name desc&_limit=2&_offset=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	md := gateway.MetadataAnnotator(nil, req)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		gormDB, mock := setUp(t)

		rq := req.(*testRequest)
		gormDB, err = ApplyCollectionOperators(gormDB, &Person{}, rq.Filtering, rq.Sorting, rq.Pagination, rq.FieldSelection)
		if err != nil {
			t.Fatal(err)
		}

		mock.ExpectQuery(fixedFullRe("SELECT \"people\".* FROM \"people\" LEFT JOIN sub_people ON people.id = sub_people.person_id WHERE (((people.age <= $1) AND (sub_people.name = $2))) ORDER BY people.age,sub_people.name desc LIMIT 2 OFFSET 1")).WithArgs(25.0, "Mike").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(111, "Mike"))

		mock.ExpectQuery(fixedFullRe("SELECT * FROM  \"sub_people\" WHERE (\"person_id\" IN ($1))")).WithArgs(111)

		var actual []Person
		gormDB.Find(&actual)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
		return nil
	}

	err = gateway.ClientUnaryInterceptor(ctx, req.Method, &testRequest{}, &testResponse{}, nil, invoker)
	if err != nil {
		t.Fatal("no error returned")
	}
}
