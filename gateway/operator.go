package gateway

import (
	"context"
	"net/http"
	"strconv"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	filterQueryKey           = "_filter"
	sortQueryKey             = "_order_by"
	fieldsQueryKey           = "_fields"
	limitQueryKey            = "_limit"
	offsetQueryKey           = "_offset"
	pageTokenQueryKey        = "_page_token"
	pageInfoSizeMetaKey      = "status-page-info-size"
	pageInfoOffsetMetaKey    = "status-page-info-offset"
	pageInfoPageTokenMetaKey = "status-page-info-page_token"

	query_url = "query_url"
)

// MetadataAnnotator is a function for passing metadata to a gRPC context
// It must be mainly used as ServeMuxOption for gRPC Gateway 'ServeMux'
// See: 'WithMetadata' option.
//
// MetadataAnnotator stores request URL in gRPC metadata from incoming HTTP кequest
func MetadataAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	mdmap := make(map[string]string)
	mdmap[query_url] = req.URL.String()
	return metadata.New(mdmap)
}

// SetPagination sets page info to outgoing gRPC context.
// Deprecated: Please add `infoblox.api.PageInfo` as part of gRPC message and do not rely on outgoing gRPC context.
func SetPageInfo(ctx context.Context, p *query.PageInfo) error {
	m := make(map[string]string)

	if pt := p.GetPageToken(); pt != "" {
		m[pageInfoPageTokenMetaKey] = pt
	}

	if o := p.GetOffset(); o != 0 && p.NoMore() {
		m[pageInfoOffsetMetaKey] = "null"
	} else if o != 0 {
		m[pageInfoOffsetMetaKey] = strconv.FormatUint(uint64(o), 10)
	}

	if s := p.GetSize(); s != 0 {
		m[pageInfoSizeMetaKey] = strconv.FormatUint(uint64(s), 10)
	}

	return grpc.SetHeader(ctx, metadata.New(m))
}
