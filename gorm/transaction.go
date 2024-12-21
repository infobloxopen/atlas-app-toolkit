package gorm

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/jinzhu/gorm"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ctxKey is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxKey int

// readOnlyDBKey is an unexported type and used to define a key for storing read only db instance in the context.
// This prevents collisions with keys defined in other package
type readOnlyDBKey int

// txnKey is the key for `*Transaction` values in `context.Context`.
// It is unexported; clients use NewContext and FromContext
// instead of using this key directly.
var txnKey ctxKey

// roDBKey is the key used for storing read-only db instance in the context.
// It is unexported; clients use BeginFromContext with options to get the read only db instance
// instead of using this key directly.
var roDBKey readOnlyDBKey

var (
	ErrCtxTxnMissing = errors.New("Database transaction for request missing in context")
	ErrCtxTxnNoDB    = errors.New("Transaction in context, but DB is nil")
	ErrNoReadOnlyDB  = errors.New("No read-only DB")
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
	current         *gorm.DB
	readOnly        bool
	afterCommitHook []func(context.Context)
}

type databaseOptions struct {
	readOnlyReplica bool
}

type DatabaseOption func(*databaseOptions)

// WithRODB returns clouser to set the readOnlyReplica flag
func WithRODB(readOnlyReplica bool) DatabaseOption {
	return func(ops *databaseOptions) {
		ops.readOnlyReplica = readOnlyReplica
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

// getReadOnlyDBInstance returns the read only database instance stored in the ctx
func getReadOnlyDBInstance(ctx context.Context) (*gorm.DB, error) {
	txn, ok := FromContext(ctx)
	if !ok {
		return nil, ErrCtxTxnMissing
	}
	dbRO, ok := ctx.Value(roDBKey).(*gorm.DB)
	if !ok {
		logger.Warnf("BeginFromContext: requested: read-only DB, returns: read-write DB, reason: read-only DB not available")
		if txn.parent == nil {
			return nil, ErrCtxTxnNoDB
		}
		db := txn.beginWithContext(ctx)
		if db.Error != nil {
			return nil, db.Error
		}
		return db, nil
	}
	txn.readOnly = true
	return dbRO, nil
}

// BeginFromContext will return read only db instance if readOnlyReplica flag is set otherwise it will extract transaction wrapper from context and start new transaction
// If readOnlyReplica flag is set and a txn with read-write db is already in use then it will return a txn from ctx rather than providing a read-only db instance.
// As result new instance of `*gorm.DB` will be returned.
// Error will be returned in case either transaction or db connection info is missing in context.
// Gorm specific error can be checked by `*gorm.DB.Error`.
func BeginFromContext(ctx context.Context, options ...DatabaseOption) (*gorm.DB, error) {
	txn, ok := FromContext(ctx)
	if !ok {
		return nil, ErrCtxTxnMissing
	}
	opts := toDatabaseOptions(options...)
	if opts.readOnlyReplica == true {
		if txn.current == nil {
			return getReadOnlyDBInstance(ctx)
		} else {
			logger.Warnf("BeginFromContext: requested: read-only DB, returns: read-write DB, reason: read-write DB txn in use")
			return txn.current, nil
		}
	} else if txn.readOnly == true {
		logger.Warnf("BeginFromContext: requested: read-write DB, returns: read-only DB, reason: txn set to read only")
		return getReadOnlyDBInstance(ctx)
	}
	if txn.parent == nil {
		return nil, ErrCtxTxnNoDB
	}
	db := txn.beginWithContext(ctx)
	if db.Error != nil {
		return nil, db.Error
	}
	return db, nil
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
	return UnaryServerInterceptorTxn(txn, readOnlyDB...)
}

func UnaryServerInterceptorTxn(txn *Transaction, readOnlyDB ...*gorm.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Deep copy is necessary as a tansaction should be created per request.
		txn := &Transaction{parent: txn.parent, afterCommitHook: txn.afterCommitHook}
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
		if len(readOnlyDB) > 0 {
			dbRO := readOnlyDB[0]
			if dbRO != nil {
				ctx = context.WithValue(ctx, roDBKey, dbRO)
			}
		}
		resp, err = handler(ctx, req)

		return resp, err
	}
}
