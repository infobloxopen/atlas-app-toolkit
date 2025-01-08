package gorm

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ctxKey is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxKey int

// txnKey is the key for `*Transaction` values in `context.Context`.
// It is unexported; clients use NewContext and FromContext
// instead of using this key directly.
var txnKey ctxKey

var (
	ErrCtxTxnMissing     = errors.New("Database transaction for request missing in context")
	ErrCtxTxnNoDB        = errors.New("Transaction in context, but DB is nil")
	ErrCtxTxnOptMismatch = errors.New("Transaction in context, but Txn opts are mismatched")
	ErrCtxDBOptMismatch  = errors.New("Transaction in context, but DB opts are mismatched")
)

// This is used to define various options to set/get the current database setup in the context
type dbType int

const (
	dbNotSet dbType = iota
	dbReadOnly
	dbReadWrite
)

// NewContext returns a new Context that carries value txn.
func NewContext(parent context.Context, txn *Transaction) context.Context {
	return context.WithValue(parent, txnKey, txn)
}

// FromContext returns the *Transaction value stored in ctx, if any.
func FromContext(ctx context.Context) (txn *Transaction, ok bool) {
	txn, ok = ctx.Value(txnKey).(*Transaction)
	return
}

// Transaction serves as a wrapper around `*gorm.DB` instance.
// It works as a singleton to prevent an application of creating more than one
// transaction instance per incoming request.
type Transaction struct {
	mu              sync.Mutex
	parent          *gorm.DB
	parentRO        *gorm.DB
	current         *gorm.DB
	currentOpts     databaseOptions
	afterCommitHook []func(context.Context)
}

type databaseOptions struct {
	database dbType
	txOpts   *sql.TxOptions
}

type DatabaseOption func(*databaseOptions)

// WithRODB returns closure to set the readOnlyReplica flag
func WithRODB(readOnlyReplica bool) DatabaseOption {
	return func(ops *databaseOptions) {
		if readOnlyReplica == false {
			ops.database = dbReadWrite
		} else {
			ops.database = dbReadOnly
		}
	}
}

// WithTxOptions returns a closure to set the TxOptions
func WithTxOptions(opts *sql.TxOptions) DatabaseOption {
	return func(ops *databaseOptions) {
		ops.txOpts = opts
	}
}
func toDatabaseOptions(options ...DatabaseOption) *databaseOptions {
	opts := &databaseOptions{}
	for _, op := range options {
		op(opts)
	}
	return opts
}

func NewTransaction(db *gorm.DB) Transaction {
	return Transaction{parent: db}
}

func (t *Transaction) AddAfterCommitHook(hooks ...func(context.Context)) {
	t.afterCommitHook = append(t.afterCommitHook, hooks...)
}

// getReadOnlyDBInstance returns the read only db txn if RO DB available otherwise it returns read/write db txn
func getReadOnlyDBTxn(ctx context.Context, opts *databaseOptions, txn *Transaction) (*gorm.DB, error) {
	var db *gorm.DB
	switch {
	case txn.parentRO == nil:
		return getReadWriteDBTxn(ctx, opts, txn)
	case opts.txOpts != nil && txn.currentOpts.txOpts != nil:
		if *opts.txOpts != *txn.currentOpts.txOpts {
			return nil, ErrCtxTxnOptMismatch
		}
	case opts.txOpts != nil:
		// We should error in two cases 1. We should error if read-only DB requested with read-write txn
		// 2. If no txn options provided in previous call but provided in subsequent call
		if opts.txOpts.ReadOnly == false || txn.currentOpts.database != dbNotSet {
			return nil, ErrCtxTxnOptMismatch
		}
		txnOpts := *opts.txOpts
		txn.currentOpts.txOpts = &txnOpts
	}
	if txn.current != nil {
		return txn.current, nil
	}
	db = txn.beginReadOnlyWithContextAndOptions(ctx, txn.currentOpts.txOpts)
	if db.Error != nil {
		return nil, db.Error
	}
	if txn.currentOpts.database == dbNotSet {
		txn.currentOpts.database = dbReadOnly
	}
	return db, nil
}

// getReadWriteDBTxn returns the read/write db txn
func getReadWriteDBTxn(ctx context.Context, opts *databaseOptions, txn *Transaction) (*gorm.DB, error) {
	var db *gorm.DB
	switch {
	case txn.parent == nil:
		return nil, ErrCtxTxnNoDB
	case opts.txOpts != nil && txn.currentOpts.txOpts != nil:
		if *opts.txOpts != *txn.currentOpts.txOpts {
			return nil, ErrCtxTxnOptMismatch
		}
	case opts.txOpts != nil:
		// We should return error If no txn options provided in previous call but provided in subsequent call
		if txn.currentOpts.database != dbNotSet {
			return nil, ErrCtxTxnOptMismatch
		}
		txnOpts := *opts.txOpts
		txn.currentOpts.txOpts = &txnOpts
	}
	if txn.current != nil {
		return txn.current, nil
	}
	db = txn.beginWithContextAndOptions(ctx, txn.currentOpts.txOpts)
	if db.Error != nil {
		return nil, db.Error
	}
	if txn.currentOpts.database == dbNotSet {
		txn.currentOpts.database = dbReadWrite
	}
	return db, nil
}

// BeginFromContext will return read only db txn if readOnlyReplica flag is set otherwise it will extract transaction wrapper from context and start new transaction
// If readOnlyReplica flag is set and read only db is not available then it will check if a txn with read-write db is already in use then it will return a txn from ctx otherwise it will start a new txn with read/write db and return
// As result new instance of `*gorm.DB` will be returned.
// Error will be returned in case either transaction or db connection info is missing in context.
// Gorm specific error can be checked by `*gorm.DB.Error`.
func BeginFromContext(ctx context.Context, options ...DatabaseOption) (*gorm.DB, error) {
	txn, ok := FromContext(ctx)
	if !ok {
		return nil, ErrCtxTxnMissing
	}
	opts := toDatabaseOptions(options...)
	switch opts.database {
	case dbReadOnly:
		if txn.currentOpts.database == dbReadWrite && txn.parentRO != nil {
			return nil, ErrCtxDBOptMismatch
		}
		return getReadOnlyDBTxn(ctx, opts, txn)
	case dbReadWrite:
		if txn.currentOpts.database == dbReadOnly {
			return nil, ErrCtxDBOptMismatch
		}
		return getReadWriteDBTxn(ctx, opts, txn)
	default:
		// This is the case to handle when no database options provided
		if txn.currentOpts.database == dbReadOnly {
			return getReadOnlyDBTxn(ctx, opts, txn)
		}
		return getReadWriteDBTxn(ctx, opts, txn)
	}
}

// BeginWithOptionsFromContext will extract transaction wrapper from context and start new transaction,
// options can be specified to control isolation level for transaction.
// As result new instance of `*gorm.DB` will be returned.
// Error will be returned in case either transaction or db connection info is missing in context.
// Gorm specific error can be checked by `*gorm.DB.Error`.
func BeginWithOptionsFromContext(ctx context.Context, opts *sql.TxOptions) (*gorm.DB, error) {
	txn, ok := FromContext(ctx)
	if !ok {
		return nil, ErrCtxTxnMissing
	}
	if txn.parent == nil {
		return nil, ErrCtxTxnNoDB
	}
	db := txn.beginWithContextAndOptions(ctx, opts)
	if db.Error != nil {
		return nil, db.Error
	}
	return db, nil
}

// Begin starts new transaction by calling `*gorm.DB.Begin()`
// Returns new instance of `*gorm.DB` (error can be checked by `*gorm.DB.Error`)
func (t *Transaction) Begin() *gorm.DB {
	return t.beginWithContext(context.Background())
}

func (t *Transaction) beginWithContext(ctx context.Context) *gorm.DB {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.current == nil {
		t.current = t.parent.BeginTx(ctx, nil)
	}

	return t.current
}

// BeginWithOptions starts new transaction by calling `*gorm.DB.BeginTx()`
// Returns new instance of `*gorm.DB` (error can be checked by `*gorm.DB.Error`)
func (t *Transaction) BeginWithOptions(opts *sql.TxOptions) *gorm.DB {
	return t.beginWithContextAndOptions(context.Background(), opts)
}

// beginReadOnlyWithContextAndOptions will start a new transaction by calling `*gorm.DB.BeginTx() if no current transaction exist
func (t *Transaction) beginReadOnlyWithContextAndOptions(ctx context.Context, opts *sql.TxOptions) *gorm.DB {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.current == nil {
		t.current = t.parentRO.BeginTx(ctx, opts)
	}

	return t.current
}

func (t *Transaction) beginWithContextAndOptions(ctx context.Context, opts *sql.TxOptions) *gorm.DB {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.current == nil {
		t.current = t.parent.BeginTx(ctx, opts)
	}

	return t.current
}

// Rollback terminates transaction by calling `*gorm.DB.Rollback()`
// Reset current transaction and returns an error if any.
func (t *Transaction) Rollback() error {
	if t.current == nil {
		return nil
	}
	if reflect.ValueOf(t.current.CommonDB()).IsNil() {
		return status.Error(codes.Unavailable, "Database connection not available")
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current.Rollback()
	err := t.current.Error
	t.current = nil
	return err
}

// Commit finishes transaction by calling `*gorm.DB.Commit()`
// Reset current transaction and returns an error if any.
func (t *Transaction) Commit(ctx context.Context) error {
	if t.current == nil || reflect.ValueOf(t.current.CommonDB()).IsNil() {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current.Commit()
	err := t.current.Error
	if err == nil {
		for i := range t.afterCommitHook {
			t.afterCommitHook[i](ctx)
		}
	}
	t.current = nil
	return err
}

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor that manages
// a `*Transaction` instance.
// New *Transaction instance is created before grpc.UnaryHandler call.
// Client is responsible to call `txn.Begin()` to open transaction.
// If call of grpc.UnaryHandler returns with an error the transaction
// is aborted, otherwise committed.
func UnaryServerInterceptor(db *gorm.DB, readOnlyDB ...*gorm.DB) grpc.UnaryServerInterceptor {
	txn := &Transaction{parent: db}
	if len(readOnlyDB) > 0 {
		dbRO := readOnlyDB[0]
		if dbRO != nil {
			txn.parentRO = dbRO
		}
	}
	return UnaryServerInterceptorTxn(txn)
}

func UnaryServerInterceptorTxn(txn *Transaction) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Deep copy is necessary as a tansaction should be created per request.
		txn := &Transaction{parent: txn.parent, parentRO: txn.parentRO, afterCommitHook: txn.afterCommitHook}
		defer func() {
			// simple panic handler
			if perr := recover(); perr != nil {
				// we do not try to safe the world -
				// just attempt to close our transaction
				// re-raise panic and let someone to handle it
				txn.Rollback()
				panic(perr)
			}

			var terr error
			if err != nil {
				terr = txn.Rollback()
			} else {
				if terr = txn.Commit(ctx); terr != nil {
					err = status.Error(codes.Internal, "failed to commit transaction")
				}
			}

			if terr == nil {
				return
			}
			// Catch the status: UNAVAILABLE error that Rollback might return
			if _, ok := status.FromError(terr); ok {
				err = terr
				return
			}

			st := status.Convert(err)
			st, serr := st.WithDetails(errdetails.New(codes.Internal, "gorm", terr.Error()))
			// do not override error if failed to attach details
			if serr == nil {
				err = st.Err()
			}
			return
		}()

		ctx = NewContext(ctx, txn)
		resp, err = handler(ctx, req)

		return resp, err
	}
}
