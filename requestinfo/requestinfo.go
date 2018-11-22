package requestinfo

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"net/http"
	"strings"
)

const (
	CreateOperation OperationType = iota
	ReplaceOperation
	UpdateOperation
	DeleteOperation
	ReadOperation
	ListOperation
	UnknownOperation
)

const (
	appNameMetaKey       = "request_info_app_name"
	resourceTypeMetaKey  = "request_info_resource_type"
	resourceIdMetaKey    = "request_info_resource_id"
	operationTypeMetaKey = "request_info_operation_type"
)

var operationNameToType = map[string]OperationType{
	"Create":  CreateOperation,
	"Replace": ReplaceOperation,
	"Update":  UpdateOperation,
	"Delete":  DeleteOperation,
	"Read":    ReadOperation,
	"List":    ListOperation,
	"Unknown": UnknownOperation,
}

var operationTypeToName = map[OperationType]string{
	CreateOperation:  "Create",
	ReplaceOperation: "Replace",
	UpdateOperation:  "Update",
	DeleteOperation:  "Delete",
	ReadOperation:    "Read",
	ListOperation:    "List",
	UnknownOperation: "Unknown",
}

type OperationType int

func (o OperationType) String() string {
	if result, ok := operationTypeToName[o]; ok {
		return result
	}

	return operationTypeToName[UnknownOperation]
}

type RequestInfo struct {
	Identifier    resource.Identifier
	OperationType OperationType
}

func NewRequestInfo(r *http.Request) (RequestInfo, error) {
	info := RequestInfo{}
	if r == nil {
		return info, ErrHTTPRequestIsMissing
	}

	path := strings.TrimPrefix(r.URL.EscapedPath(), "/")
	parts := strings.Split(path, "/")
	hasId := len(parts) == 3
	if len(parts) < 2 || len(parts) > 3 {
		return RequestInfo{}, ErrInvalidHTTPRequestPath
	}

	var appName, rType, rID string
	appName, rType = parts[0], parts[1]
	if hasId {
		rID = parts[2]
	}

	info.Identifier = resource.Identifier{
		ApplicationName: appName,
		ResourceType:    rType,
		ResourceId:      rID,
	}

	info.OperationType = GetOperationTypeFromRequest(r, hasId)

	return info, nil
}

func GetOperationTypeFromRequest(r *http.Request, hasIdentifier bool) OperationType {
	switch r.Method {
	case "GET":
		if hasIdentifier {
			return ReadOperation
		}

		return ListOperation

	case "POST":
		return CreateOperation

	case "PUT":
		return ReplaceOperation

	case "PATCH":
		return UpdateOperation

	case "DELETE":
		return DeleteOperation
	}

	return UnknownOperation
}

func requestInfoToMap(reqInfo RequestInfo) (map[string]string, error) {
	metadata := make(map[string]string, 0)
	metadata[runtime.MetadataPrefix+appNameMetaKey] = reqInfo.Identifier.ApplicationName
	metadata[runtime.MetadataPrefix+resourceTypeMetaKey] = reqInfo.Identifier.ResourceType
	metadata[runtime.MetadataPrefix+resourceIdMetaKey] = reqInfo.Identifier.ResourceId

	var ok bool
	if metadata[runtime.MetadataPrefix+operationTypeMetaKey], ok = operationTypeToName[reqInfo.OperationType]; !ok {
		metadata[runtime.MetadataPrefix+operationTypeMetaKey] = operationTypeToName[UnknownOperation]
	}

	return metadata, nil
}
