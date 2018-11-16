package requestinfo

import (
	"context"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
)

func FromContext(ctx context.Context) (RequestInfo, error) {
	info := RequestInfo{}

	var found bool
	if info.Identifier.ApplicationName, found = gateway.Header(ctx, appNameMetaKey); !found {
		return RequestInfo{}, ErrAppNameIsMissing
	}

	if info.Identifier.ResourceType, found = gateway.Header(ctx, resourceTypeMetaKey); !found {
		return RequestInfo{}, ErrResourceTypeIsMissing
	}

	if info.Identifier.ResourceId, found = gateway.Header(ctx, resourceIdMetaKey); !found {
		// Note: if resource Id is not available it should be stored as empty string in the context:
		return RequestInfo{}, ErrResourceIdIsMissing
	}

	var op string
	if op, found = gateway.Header(ctx, operationTypeMetaKey); !found {
		return RequestInfo{}, ErrOperationNameIsMissing
	}

	if info.OperationType, found = operationNameToType[op]; !found {
		return RequestInfo{}, ErrInvalidOperation
	}

	return info, nil
}
