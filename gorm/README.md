# GORM

This package contains a set of utilities for famous [GORM](http://gorm.io/) library. If used together they can significantly help you to enable persistency in your application.

## Collection Operators

The package provides some helpers which are able to apply collection operators defined in [query](../query) package to a GORM [query](http://gorm.io/docs/query.html#Query).

### Applying query.Filtering

```golang
...
// given that Person is a protobuf message and PersonORM is a corresponding GORM model
db, assoc, err = gorm.ApplyFiltering(ctx, db, filtering, &PersonORM{}, &Person{})
if err != nil {
    ...
}
// Join required associations(for nested field support)
db, err = gorm.JoinAssociations(ctx, db, assoc, &PersonORM{})
if err != nil {
    ...
}
var people []Person
db.Find(&people)
...
```
### Applying query.Sorting

```golang
...
db, assoc, err = gorm.ApplySorting(ctx, db, sorting, &PersonORM{}, &Person{})
if err != nil {
    ...
}
// Join required associations(for nested field support)
db, err = gorm.JoinAssociations(ctx, db, assoc, &PersonORM{})
if err != nil {
    ...
}
var people []Person
db.Find(&people)
...
```
### Applying query.Pagination

```golang
...
db = gorm.ApplyPagination(ctx, db, pagination, &PersonORM{}, &Person{})
if err != nil {
    ...
}
var people []Person
db.Find(&people)
...
```
### Applying query.FieldSelection

```golang
...
db, err = gorm.ApplyFieldSelection(ctx, db, fields, &PersonORM{}, &Person{})
if err != nil {
    ...
}
var people []Person
db.Find(&people)
...
```

### Applying everything

```golang
...
db, err = gorm.ApplyCollectionOperators(ctx, db, &PersonORM{}, &Person{}, filtering, sorting, pagination, fields)
if err != nil {
    ...
}
var people []Person
db.Find(&people)
...
```


## Transaction Management

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

## Migration version validation

The toolkit does not require any specific method for database provisioning and setup.
However, if [golang-migrate](https://github.com/golang-migrate/migrate) or the [infobloxopen fork](https://github.com/infobloxopen/migrate) of it is used, a couple helper functions are provided [here](version.go) for verifying that the database version matches a required version without having to import the entire migration package.
