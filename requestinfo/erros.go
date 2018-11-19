package requestinfo

import "fmt"

var (
	ErrHTTPRequestIsMissing   = fmt.Errorf("request is missing")
	ErrInvalidHTTPRequestPath = fmt.Errorf("invalid HTTP request path")
	ErrAppNameIsMissing       = fmt.Errorf("application name is missing in the context object")
	ErrResourceTypeIsMissing  = fmt.Errorf("resource type is missing in the context object")
)
