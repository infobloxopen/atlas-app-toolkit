package v2

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/infobloxopen/atlas-app-toolkit/v2/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUnaryServerInterceptor_success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
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

func TestUnaryServerInterceptorTxn_success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := NewTransaction(gdb)
	interceptor := UnaryServerInterceptorTxn(&txn)
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
		t.Fatalf("failed to create sqlmock - %s", err)
	}

	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(errors.New("handler"))

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
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
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
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
	if len(details) == 0 {
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

// mockServerStream implements grpc.ServerStream for testing
type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SendMsg(interface{}) error    { return nil }
func (m *mockServerStream) RecvMsg(interface{}) error    { return nil }

func TestStreamServerInterceptor(t *testing.T) {
	tests := []struct {
		name                string
		usePreConfiguredTxn bool
		setupMock           func(mock sqlmock.Sqlmock)
		handler             func(t *testing.T) grpc.StreamHandler
		expectError         bool
		expectedCode        codes.Code
		expectedMessage     string
		checkDetails        bool
	}{
		{
			name:                "success with StreamServerInterceptor",
			usePreConfiguredTxn: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			handler: func(t *testing.T) grpc.StreamHandler {
				return func(srv any, stream grpc.ServerStream) error {
					txn, ok := FromContext(stream.Context())
					if !ok {
						t.Error("failed to extract transaction from context")
					}
					return txn.Begin().Error
				}
			},
			expectError: false,
		},
		{
			name:                "success with StreamServerInterceptorTxn",
			usePreConfiguredTxn: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			handler: func(t *testing.T) grpc.StreamHandler {
				return func(srv any, stream grpc.ServerStream) error {
					txn, ok := FromContext(stream.Context())
					if !ok {
						t.Error("failed to extract transaction from context")
					}
					return txn.Begin().Error
				}
			},
			expectError: false,
		},
		{
			name:                "error handling with rollback",
			usePreConfiguredTxn: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback().WillReturnError(errors.New("handler"))
			},
			handler: func(t *testing.T) grpc.StreamHandler {
				return func(srv any, stream grpc.ServerStream) error {
					txn, ok := FromContext(stream.Context())
					if !ok {
						t.Error("failed to extract transaction from context")
					}
					txn.Begin()
					return status.Error(codes.InvalidArgument, "handler")
				}
			},
			expectError:     true,
			expectedCode:    codes.InvalidArgument,
			expectedMessage: "handler",
		},
		{
			name:                "error details attachment",
			usePreConfiguredTxn: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("internal"))
			},
			handler: func(t *testing.T) grpc.StreamHandler {
				return func(srv interface{}, stream grpc.ServerStream) error {
					txn, ok := FromContext(stream.Context())
					if !ok {
						t.Error("failed to extract transaction from context")
					}
					txn.Begin()
					// Don't set an error on the transaction itself, 
					// let the commit fail to trigger the error details path
					return nil
				}
			},
			expectError:     true,
			expectedCode:    codes.Internal,
			expectedMessage: "failed to commit transaction",
			checkDetails:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock - %s", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Logf("failed to close db: %v", err)
				}
			}()

			tt.setupMock(mock)

			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db - %s", err)
			}

			var interceptor grpc.StreamServerInterceptor
			if tt.usePreConfiguredTxn {
				txn := NewTransaction(gdb)
				interceptor = StreamServerInterceptorTxn(&txn)
			} else {
				interceptor = StreamServerInterceptor(gdb)
			}

			stream := &mockServerStream{ctx: context.Background()}
			err = interceptor(nil, stream, nil, tt.handler(t))

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}

				st := status.Convert(err)
				if st.Code() != tt.expectedCode {
					t.Errorf("expected code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMessage {
					t.Errorf("expected message %q, got %q", tt.expectedMessage, st.Message())
				}

				if tt.checkDetails {
					details := st.Details()
					if len(details) == 0 {
						t.Fatal("expected details but got none")
					}

					d, ok := details[0].(*errdetails.TargetInfo)
					if !ok {
						t.Fatal("expected TargetInfo details")
					}
					if d.Code != int32(codes.Internal) || d.Message != "internal" || d.Target != "gorm" {
						t.Errorf("invalid target info - code=%d, message=%q, target=%q", d.Code, d.Message, d.Target)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error - %s", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("failed to meet mock expectations - %s", err)
			}
		})
	}
}

func TestTransaction_Begin(t *testing.T) {
	tests := []struct {
		desc     string
		withOpts bool
	}{
		{
			desc:     "begin without options",
			withOpts: false,
		},
		{
			desc:     "begin with options",
			withOpts: true,
		},
	}
	begin := func(txn *Transaction, withOpts bool) {
		switch withOpts {
		case true:
			opt := &sql.TxOptions{Isolation: sql.LevelSerializable}
			txn.BeginWithOptions(opt)
		case false:
			txn.Begin()
		}
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock - %s", err)
			}
			mock.ExpectBegin()

			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
			txn := &Transaction{parent: gdb}

			// test singleton behavior
			begin(txn, test.withOpts)
			if txn.current == nil {
				t.Fatal("failed to begin transaction")
			}
			prev := txn.current
			begin(txn, test.withOpts)
			if txn.current != prev {
				t.Fatal("transaction does not behaves like singleton")
			}

			// test begin behavior: no error
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("failed to begin transaction - %s", err)
			}
		})
	}
}

func TestTransaction_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := &Transaction{parent: gdb}
	ctx := context.Background()
	// test current transaction is nil
	if err := txn.Commit(ctx); err != nil {
		t.Errorf("unexpected error %s", err)
	}

	txn.Begin()
	if err := txn.Commit(ctx); err != nil {
		t.Errorf("failed to commit transaction - %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to commit transaction - %s", err)
	}

	if txn.current != nil {
		t.Error("failed to reset current gorm instance - txn.current is not nil")
	}
}

func TestTransaction_AfterCommitHook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := &Transaction{parent: gdb}
	txn.Begin()

	called := false
	hook := func(context.Context) { called = true }
	txn.AddAfterCommitHook(hook)
	ctx := context.Background()
	if err := txn.Commit(ctx); err != nil {
		t.Errorf("failed to commit transaction - %s", err)
	}
	if !called {
		t.Errorf("did not fire the hook")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failed to commit transaction - %s", err)
	}

}
func TestTransaction_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectRollback()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
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

	fdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	sqlDB, _ := fdb.DB()
	if err := sqlDB.Close(); err != nil {
		t.Logf("failed to close db: %v", err)
	}
	txn = &Transaction{parent: gdb}

	txn.Begin()
	if err := txn.Rollback(); !reflect.DeepEqual(err, status.Error(codes.Unavailable, "Database connection not available")) {
		t.Errorf("Did not receive proper error for broken DB - %s", err)
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

func beginFromContext(ctx context.Context, withOpts bool) (*gorm.DB, error) {
	switch withOpts {
	case true:
		opt := &sql.TxOptions{Isolation: sql.LevelSerializable}
		return BeginWithOptionsFromContext(ctx, opt)
	case false:
		return BeginFromContext(ctx)
	}
	return nil, nil
}

func TestBeginFromContext_Good(t *testing.T) {
	tests := []struct {
		desc     string
		withOpts bool
	}{
		{
			desc:     "begin without options",
			withOpts: false,
		},
		{
			desc:     "begin with options",
			withOpts: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock - %s", err)
			}
			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db - %s", err)
			}
			mock.ExpectBegin()

			// Case: All good
			ctxtxn := &Transaction{parent: gdb}
			ctx = NewContext(ctx, ctxtxn)

			txn1, err := beginFromContext(ctx, test.withOpts)
			if txn1 == nil {
				t.Error("Did not receive a transaction from context")
			}
			if err != nil {
				t.Error("Received an error beginning transaction")
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("failed to begin transaction - %s", err)
			}

			// Case: Transaction begin is idempotent
			txn2, err := beginFromContext(ctx, test.withOpts)
			if txn2 != txn1 {
				t.Error("Got a different txn than was opened before")
			}
			if err != nil {
				t.Error("Received an error opening transaction")
			}
		})
	}
}

func TestBeginFromContext_Bad(t *testing.T) {
	tests := []struct {
		desc            string
		withOpts        bool
		contextCanceled bool
	}{
		{
			desc:     "begin without options",
			withOpts: false,
		},
		{
			desc:     "begin with options",
			withOpts: true,
		},
		{
			desc:            "canceled context without context",
			withOpts:        true,
			contextCanceled: true,
		},
		{
			desc:            "canceled context with options",
			withOpts:        false,
			contextCanceled: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			if test.contextCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			// Case: Transaction missing from context
			txn1, err := beginFromContext(ctx, test.withOpts)
			if err != ErrCtxTxnMissing {
				t.Error("Did not receive a CtxTxnError when no context transaction was present")
			}
			if txn1 != nil {
				t.Error("Got some txn returned when nil was expected")
			}

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock - %s", err)
			}
			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db - %s", err)
			}
			mock.ExpectBegin().WillReturnError(errors.New(""))

			// Case: Transaction fails to open
			txn2, err := beginFromContext(NewContext(ctx, &Transaction{parent: gdb}), test.withOpts)
			if txn2 != nil {
				t.Error("Got some txn returned when nil was expected")
			}
			if err == nil {
				t.Error("Did not receive an error when transaction begin returned error")
			}

			// Case: DB Missing from Transaction in Context
			txn3, err := beginFromContext(NewContext(ctx, &Transaction{}), test.withOpts)
			if txn3 != nil {
				t.Error("Got some txn returned when nil was expected")
			}
			if err != ErrCtxTxnNoDB {
				t.Error("Did not receive an error opening a txn with nil DB")
			}
		})
	}
}
