package requestinfo

import "fmt"

var (
	ErrHTTPRequestIsMissing   = fmt.Errorf("request is missing")
	ErrInvalidHTTPRequestPath = fmt.Errorf("invalid HTTP request path")
	ErrInvalidOperation       = fmt.Errorf("invalid operation")
	ErrAppNameIsMissing       = fmt.Errorf("application name is missing in the context object")
	ErrResourceTypeIsMissing  = fmt.Errorf("resource type is missing in the context object")
	ErrResourceIdIsMissing    = fmt.Errorf("resource identifier is missing in the context object")
	ErrOperationNameIsMissing = fmt.Errorf("operation name is missing in the context object")
)
