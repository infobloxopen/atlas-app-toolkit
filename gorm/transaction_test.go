package gorm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/infobloxopen/atlas-app-toolkit/v2/rpc/errdetails"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor_success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
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

func TestUnaryServerInterceptorTxn_success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	gdb, err := gorm.Open("postgres", db)
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
		t.Fatalf("failed to create sqlmock - %s", err)
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

			gdb, err := gorm.Open("postgres", db)
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

	gdb, err := gorm.Open("postgres", db)
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

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}
	txn := &Transaction{parent: gdb}
	txn.Begin()

	called := false
	hook := func(context.Context) { called = true; return }
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

	fdb, err := gorm.Open("postgres", db)
	fdb.Close()
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
			gdb, err := gorm.Open("postgres", db)
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
			gdb, err := gorm.Open("postgres", db)
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

// newMockGorm returns a gorm DB backed by sqlmock, closed when the test ends.
func newMockGorm(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock - %s", err)
	}
	t.Cleanup(func() { db.Close() })

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm db - %s", err)
	}

	return gdb, mock
}

// mockServerStream is the minimum grpc.ServerStream the stream interceptor needs: it only reads Context.
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context { return m.ctx }

// commitFailureCase describes one way a request can end with a failed commit.
type commitFailureCase struct {
	name string
	// commitErr is what the database returns for COMMIT.  When nil the driver never sees a
	// commit: the request's own cancellation makes database/sql roll the transaction back first.
	commitErr error
	// useBegin opens the transaction with Begin(), which is not tied to the request context.
	// Otherwise BeginFromContext is used, which is.
	useBegin bool
	// cancelRequest cancels the request context from inside the handler, as a client hanging
	// up would.
	cancelRequest bool
	// timeout gives the request context a deadline.
	timeout time.Duration
	// wait keeps the handler running afterwards, long enough for database/sql's background
	// rollback to finish and for Commit to return sql.ErrTxDone.
	wait     time.Duration
	wantCode codes.Code
}

var commitFailureCases = []commitFailureCase{
	// The commit error itself says why, and the request context is still live.
	{name: "commit fails with context canceled", commitErr: context.Canceled, wantCode: codes.Canceled},
	{name: "commit fails with context deadline exceeded", commitErr: context.DeadlineExceeded, wantCode: codes.DeadlineExceeded},
	{name: "commit fails with wrapped cancellation", commitErr: fmt.Errorf("commit: %w", context.Canceled), wantCode: codes.Canceled},

	// Controls: a failure that has nothing to do with the caller stays Internal.
	{name: "database fault", commitErr: errors.New("deadlock detected"), wantCode: codes.Internal},
	{name: "driver error", commitErr: driver.ErrBadConn, wantCode: codes.Internal},

	// The request ends and the commit error no longer says so, which is what really happens:
	// database/sql rolls back in the background and Commit may return sql.ErrTxDone.
	{name: "client hangs up and the commit races the rollback", cancelRequest: true, wantCode: codes.Canceled},
	{name: "client hangs up and the rollback finished first", cancelRequest: true, wait: 50 * time.Millisecond, wantCode: codes.Canceled},
	{name: "deadline passes and the rollback finished first", timeout: 20 * time.Millisecond, wait: 60 * time.Millisecond, wantCode: codes.DeadlineExceeded},
	{name: "client hangs up and the driver reports its own error", commitErr: driver.ErrBadConn, useBegin: true, cancelRequest: true, wantCode: codes.Canceled},
}

// runCommitFailure runs one case through the unary or the stream interceptor and returns the
// error the request ends with.
func runCommitFailure(t *testing.T, tc commitFailureCase, stream bool) error {
	t.Helper()

	gdb, mock := newMockGorm(t)
	mock.ExpectBegin()
	if tc.commitErr != nil {
		mock.ExpectCommit().WillReturnError(tc.commitErr)
	} else {
		mock.ExpectRollback()
	}

	ctx, cancel := context.WithCancel(context.Background())
	if tc.timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), tc.timeout)
	}
	defer cancel()

	// body is what a handler does: open the transaction, let the request end, and succeed.
	body := func(c context.Context) error {
		if tc.useBegin {
			txn, ok := FromContext(c)
			if !ok {
				t.Error("failed to extract transaction from context")
				return nil
			}
			txn.Begin()
		} else if _, err := BeginFromContext(c); err != nil {
			return err
		}
		if tc.cancelRequest {
			cancel()
		}
		time.Sleep(tc.wait)

		return nil
	}

	var err error
	if stream {
		err = StreamServerInterceptor(gdb)(nil, &mockServerStream{ctx: ctx}, nil, func(srv interface{}, s grpc.ServerStream) error {
			return body(s.Context())
		})
	} else {
		_, err = UnaryServerInterceptor(gdb)(ctx, nil, nil, func(c context.Context, req interface{}) (interface{}, error) {
			return nil, body(c)
		})
	}

	// When the request's own cancellation drives the outcome, the rollback happens in a
	// background goroutine, so there is no deterministic point at which to check the mock.
	if tc.commitErr != nil {
		if merr := mock.ExpectationsWereMet(); merr != nil {
			t.Errorf("failed to meet mock expectations - %s", merr)
		}
	}

	return err
}

// checkCommitFailure asserts what a failed commit must look like to the caller: the code follows
// the cause, the message stays the stable one, and the real error is still attached as the
// "gorm" detail, labelled with the same code.
func checkCommitFailure(t *testing.T, err error, tc commitFailureCase) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error from the failed commit")
	}

	st := status.Convert(err)
	if st.Code() != tc.wantCode {
		t.Errorf("expected code %v, got %v (%q)", tc.wantCode, st.Code(), st.Message())
	}
	if st.Message() != "failed to commit transaction" {
		t.Errorf("unexpected message %q", st.Message())
	}

	details := st.Details()
	if len(details) == 0 {
		t.Fatal("expected the gorm detail to be attached")
	}
	d, ok := details[0].(*errdetails.TargetInfo)
	if !ok {
		t.Fatalf("unexpected detail type %T", details[0])
	}
	if d.Code != int32(tc.wantCode) || d.Target != "gorm" {
		t.Errorf("invalid target info - code=%d, target=%q", d.Code, d.Target)
	}
	if tc.commitErr != nil && d.Message != tc.commitErr.Error() {
		t.Errorf("expected detail message %q, got %q", tc.commitErr.Error(), d.Message)
	}
}

func TestUnaryServerInterceptor_commitFailureCode(t *testing.T) {
	for _, tc := range commitFailureCases {
		t.Run(tc.name, func(t *testing.T) {
			checkCommitFailure(t, runCommitFailure(t, tc, false), tc)
		})
	}
}

func TestStreamServerInterceptor_commitFailureCode(t *testing.T) {
	for _, tc := range commitFailureCases {
		t.Run(tc.name, func(t *testing.T) {
			checkCommitFailure(t, runCommitFailure(t, tc, true), tc)
		})
	}
}
