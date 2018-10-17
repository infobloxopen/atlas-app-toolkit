package gorm

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type Person struct {
	Id        int64
	Name      string
	Age       int
	ParentId  int64
	Parent    Parent        `gorm:"foreignkey:ParentId;association_foreignkey:Id"`
	SubPerson SubPerson     `gorm:"foreignkey:PersonId;association_foreignkey:Id"`
	Items     []OrderedItem `gorm:"foreignkey:PersonId;association_foreignkey:Id" atlas:"position:Position"`
}

type Parent struct {
	Id   int64
	Name string
}

type SubPerson struct {
	Id       int64
	Name     string
	PersonId int64
}

type OrderedItem struct {
	Id       int64
	Position int
	PersonId int64
}

type PersonProto struct {
}

func (*PersonProto) Reset() {
}

func (*PersonProto) ProtoMessage() {
}

func (*PersonProto) String() string {
	return "Person"
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

	req, err := http.NewRequest("GET", "http://test.com?_fields=id,name,sub_person,items&_filter=age<=25 and sub_person.name=='Mike'&_order_by=age,sub_person.name,parent.name desc&_limit=2&_offset=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	md := gateway.MetadataAnnotator(nil, req)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		gormDB, mock := setUp(t)

		rq := req.(*testRequest)
		gormDB, err = ApplyCollectionOperators(ctx, gormDB, &Person{}, &PersonProto{}, rq.Filtering, rq.Sorting, rq.Pagination, rq.FieldSelection)
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(`^SELECT "people".\* FROM "people" LEFT JOIN [[sub_people sub_person ON people.id = sub_person.person_id LEFT JOIN parents parent ON people.parent_id = parent.id]|[parents parent ON people.parent_id = parent.id LEFT JOIN sub_people sub_person ON people.id = sub_person.person_id]] WHERE (((people.age <= \$1) AND (sub_person.name = \$2))) ORDER BY people.age,sub_person.name,parent.name desc LIMIT 2 OFFSET 1$`).WithArgs(25.0, "Mike").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(111, "Mike"))

		mock.ExpectQuery(fixedFullRe("SELECT * FROM  \"ordered_items\" WHERE (\"person_id\" IN ($1)) ORDER BY \"position\"")).WithArgs(111).
			WillReturnRows(sqlmock.NewRows([]string{"id", "position", "person_id"}))
		mock.ExpectQuery(fixedFullRe("SELECT * FROM  \"sub_people\" WHERE (\"person_id\" IN ($1))")).WithArgs(111).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "position"}))

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
