package requestinfo

import (
	"context"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func MetadataAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	info, err := NewRequestInfo(r)
	if err != nil {
		return metadata.MD{}
	}

	mInfo, err := requestInfoToMap(info)
	if err != nil {
		return metadata.MD{}
	}

	return metadata.New(mInfo)
}
