package requestid

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func NewRequestIDAnnotator() func(context.Context, *http.Request) metadata.MD {
	return func(ctx context.Context, req *http.Request) metadata.MD {
		if req.Header.Get(DefaultRequestIDKey) != "" {
			md := make(metadata.MD)
			md.Set(DefaultRequestIDKey, req.Header.Get(DefaultRequestIDKey))
			return md
		} else {
			return nil
		}
	}
}
