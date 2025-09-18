# GORM v2 Transaction Package

This package provides GORM v2 compatible transaction management utilities for gRPC services.


```go
import (
    gormv2 "github.com/infobloxopen/atlas-app-toolkit/gorm/v2"
    "gorm.io/gorm"
)

// Create a transaction interceptor
db, err := gorm.Open(...)
interceptor := gormv2.UnaryServerInterceptor(db)

// Use in your handler
func (s *Server) MyHandler(ctx context.Context, req *Request) (*Response, error) {
    // Begin transaction
    tx, err := gormv2.BeginFromContext(ctx)
    if err != nil {
        return nil, err
    }
    
    // Use transaction...
    
    return response, nil
}
```

### API Compatibility

The API is designed to be compatible with the GORM v1 version while using GORM v2 under the hood.
