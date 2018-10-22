# GORM Package

##### Transaction Management

We provide transaction management by offering `gorm.Transaction` wrapper and `gorm.UnaryServerInterceptor`.
The `gorm.Transaction` works as a singleton to prevent an application of creating more than one transaction instance per incoming request.
The `gorm.UnaryServerInterceptor` performs management on transactions.
Interceptor creates new transaction on each incoming request and commits it if request finishes without error, otherwise transaction is aborted.
The created transaction is stored in `context.Context` and passed to the request handler as usual.
**NOTE** Client is responsible to call `txn.Begin()` to open transaction.

```go
// add gorm interceptor to the chain
  server := grpc.NewServer(
    grpc.UnaryInterceptor(
      grpc_middleware.ChainUnaryServer( // middleware chain
        ...
        gorm.UnaryServerInterceptor(), // transaction management
        ...
      ),
    ),
  )
```

```go
import (
	"github.com/infobloxopen/atlas-app-toolkit/gorm"
)

func (s *MyService) MyMethod(ctx context.Context, req *MyMethodRequest) (*MyMethodResponse, error) {
	// extract gorm transaction from context
	txn, ok := gorm.FromContext(ctx)
	if !ok {
		return panic("transaction is not opened") // don't panic in production!
	}
	// start transaction
	gormDB := txn.Begin()
	if err := gormDB.Error; err != nil {
		return nil, err
	}
	// do stuff with *gorm.DB
	return &MyMethodResponse{...}, nil
}
```