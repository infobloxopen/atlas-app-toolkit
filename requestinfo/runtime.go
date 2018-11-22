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
		//In some cases ResourceId can be empty(ListOperation, CreateOperation)
	}

	//If we don't have operation Name in metadata or don't have such operation will set UnknownOperation type
	var op string
	if op, found = gateway.Header(ctx, operationTypeMetaKey); !found {
		info.OperationType = UnknownOperation
	}

	if info.OperationType, found = operationNameToType[op]; !found {
		info.OperationType = UnknownOperation
	}

	return info, nil
}
