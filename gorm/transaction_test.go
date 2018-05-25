package gorm

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor_success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("faliled to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}

	interceptor := UnaryServerInterceptor(gdb)
	_, err = interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		txn, ok := FromContext(ctx)
		if !ok {
			t.Error("failed to extract transaction from context")
		}

		return nil, txn.Begin().Error
	})
	if err != nil {
		t.Errorf("unexpected error - %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to manage transaction on success response - %s", err)
	}
}

func TestUnaryServerInterceptor_error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("faliled to create sqlmock - %s", err)
	}

	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(errors.New("handler"))

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}

	interceptor := UnaryServerInterceptor(gdb)
	_, err = interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		txn, ok := FromContext(ctx)
		if !ok {
			t.Error("failed to extract transaction from context")
		}
		txn.Begin()
		return nil, status.Error(codes.InvalidArgument, "handler")
	})

	if st := status.Convert(err); st.Message() != "handler" || st.Code() != codes.InvalidArgument {
		t.Fatalf("unexpected error - %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to manage transaction on error response - %s", err)
	}
}

func TestUnaryServerInterceptor_details(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("faliled to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}

	interceptor := UnaryServerInterceptor(gdb)
	_, err = interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		txn, ok := FromContext(ctx)
		if !ok {
			t.Error("failed to extract transaction from context")
		}
		txn.Begin()
		txn.current.Error = errors.New("internal")
		return nil, nil
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to manage transaction on error response - %s", err)
	}

	details := status.Convert(err).Details()
	if details == nil || len(details) == 0 {
		t.Fatalf("empty details")
	}

	d, ok := details[0].(*errdetails.TargetInfo)
	if !ok {
		t.Fatal("unknown type of details")
	}
	if d.Code != int32(codes.Internal) || d.Message != "internal" || d.Target != "gorm" {
		t.Errorf("invalid targer info - %s", d)
	}
}

func TestTransaction_Begin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("faliled to create sqlmock - %s", err)
	}
	mock.ExpectBegin()

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := &Transaction{parent: gdb}

	// test singleton behavior
	txn.Begin()
	if txn.current == nil {
		t.Fatal("failed to begin transaction")
	}
	prev := txn.current
	txn.Begin()
	if txn.current != prev {
		t.Fatal("transaction does not behaves like singleton")
	}

	// test begin behavior: no error
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to begin transaction - %s", err)
	}
}

func TestTransaction_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("faliled to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := &Transaction{parent: gdb}

	// test current transaction is nil
	if err := txn.Commit(); err != nil {
		t.Errorf("unexpected error %s", err)
	}

	txn.Begin()
	if err := txn.Commit(); err != nil {
		t.Errorf("failed to commit transaction - %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to commit transaction - %s", err)
	}

	if txn.current != nil {
		t.Error("failed to reset current gorm instance - txn.current is not nil")
	}
}

func TestTransaction_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("faliled to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectRollback()

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := &Transaction{parent: gdb}

	// test current transaction is nil
	if err := txn.Rollback(); err != nil {
		t.Errorf("unexpected error %s", err)
	}

	txn.Begin()
	if err := txn.Rollback(); err != nil {
		t.Errorf("failed to rollback transaction - %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to rollback transaction - %s", err)
	}

	if txn.current != nil {
		t.Error("failed to reset current gorm instance - txn.current is not nil")
	}
}

func TestContext(t *testing.T) {
	ctx := context.Background()
	_, ok := FromContext(ctx)
	if ok {
		t.Error("false positive value FromContext")
	}
	txn := &Transaction{}
	ctx = NewContext(ctx, txn)

	ftxn, ok := FromContext(ctx)
	if !ok {
		t.Error("failed to extract transaction from context")
	}
	if ftxn != txn {
		t.Error("unknown transaction instance")
	}
}
