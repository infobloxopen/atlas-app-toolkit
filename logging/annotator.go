package logging

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// HTTP Headers
const logLevelHeaderKey = "log-level"
const logFlagHeaderKey = "log-trace-key"

// Metadata keys used to pass from gateway to service
const logLevelMetaKey = "log-level"
const logFlagMetaKey = "log-trace-key"

// Name of field to be logged if included
const logFlagFieldName = "log-trace-key"

// Annotator is a function that reads the http headers of incoming requests
// searching for special logging arguments
func Annotator(ctx context.Context, req *http.Request) metadata.MD {
	md := make(metadata.MD)
	if lvl := req.Header.Get(logLevelHeaderKey); lvl != "" {
		md[logLevelMetaKey] = []string{lvl}
	}
	if flag := req.Header.Get(logFlagHeaderKey); flag != "" {
		md[logFlagMetaKey] = []string{flag}
	}

	return md
}
