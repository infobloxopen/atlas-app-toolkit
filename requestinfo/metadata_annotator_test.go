package requestinfo

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func generateRequestInfo(op OperationType, resourceId string) RequestInfo {
	info := RequestInfo{
		Identifier: resource.Identifier{
			ApplicationName: "app-name",
			ResourceType:    "resource-type",
			ResourceId:      resourceId,
		},
		OperationType: op,
	}
	return info
}
func generateRequestInfoMap(op, resourceId string) map[string]string {
	mdmap := map[string]string{
		runtime.MetadataPrefix + appNameMetaKey:       "app-name",
		runtime.MetadataPrefix + resourceTypeMetaKey:  "resource-type",
		runtime.MetadataPrefix + resourceIdMetaKey:    resourceId,
		runtime.MetadataPrefix + operationTypeMetaKey: op,
	}
	return mdmap
}

type Test struct {
}

func TestRequestInfo(t *testing.T) {
	// Positive test-cases:
	// Note: in a running service we have urls like "http://127.0.0.1:8080/ddi-ipam-ipamsvc/v1/ipam/ip_space"
	// But some middleware truncates the req.URL.EscapedPath to "ipam/ip_space", so here in UTs we have fake URLs:
	// Test Create

	testCases := []struct {
		request *http.Request
		result  RequestInfo
		err     error
	}{
		{
			request: MustRequest("POST", "http://127.0.0.1:8080/app-name/resource-type", nil),
			result:  generateRequestInfo(CreateOperation, ""),
			err:     nil,
		},
		{
			request: MustRequest("PATCH", "http://127.0.0.1:8080/app-name/resource-type/resource-id", nil),
			result:  generateRequestInfo(UpdateOperation, "resource-id"),
			err:     nil,
		},
		{
			request: MustRequest("PUT", "http://127.0.0.1:8080/app-name/resource-type/resource-id", nil),
			result:  generateRequestInfo(ReplaceOperation, "resource-id"),
			err:     nil,
		},
		{
			request: MustRequest("GET", "http://127.0.0.1:8080/app-name/resource-type/resource-id", nil),
			result:  generateRequestInfo(ReadOperation, "resource-id"),
			err:     nil,
		},
		{
			request: MustRequest("GET", "http://127.0.0.1:8080/app-name/resource-type", nil),
			result:  generateRequestInfo(ListOperation, ""),
			err:     nil,
		},
		{
			request: MustRequest("DELETE", "http://127.0.0.1:8080/app-name/resource-type/resource-id", nil),
			result:  generateRequestInfo(DeleteOperation, "resource-id"),
			err:     nil,
		},
		{
			request: MustRequest("GET", "http://127.0.0.1:8080/none", nil),
			result:  RequestInfo{},
			err:     ErrInvalidHTTPRequestPath,
		},
	}

	for testNum, test := range testCases {
		res, err := NewRequestInfo(test.request)
		if err != nil && test.err == nil {
			t.Errorf("Error not equal in test num %d expected nil, getted %s", testNum+1, err.Error())
			t.Fail()
			continue
		}

		if err == nil && err != test.err {
			t.Errorf("Error not equal in test num %d expected %s, getted nil", testNum+1, test.err)
			t.Fail()
			continue
		}

		if !reflect.DeepEqual(res, test.result) {
			t.Errorf("Result not equal in test num %d expected %+v, result %+v", testNum+1, test.result, res)
			t.Fail()
		}
	}
}

func TestRequestInfoToMap(t *testing.T) {
	// Positive test-cases:
	testCases := []struct {
		params RequestInfo
		result map[string]string
		err    error
	}{
		{
			params: generateRequestInfo(CreateOperation, ""),
			result: generateRequestInfoMap(operationTypeToName[CreateOperation], ""),
			err:    nil,
		},
		{
			params: generateRequestInfo(UpdateOperation, "resource-id"),
			result: generateRequestInfoMap(operationTypeToName[UpdateOperation], "resource-id"),
			err:    nil,
		},
		{
			params: generateRequestInfo(ReplaceOperation, "resource-id"),
			result: generateRequestInfoMap(operationTypeToName[ReplaceOperation], "resource-id"),
			err:    nil,
		},
		{
			params: generateRequestInfo(ReadOperation, "resource-id"),
			result: generateRequestInfoMap(operationTypeToName[ReadOperation], "resource-id"),
			err:    nil,
		},
		{
			params: generateRequestInfo(ReadOperation, ""),
			result: generateRequestInfoMap(operationTypeToName[ReadOperation], ""),
			err:    nil,
		},
		{
			params: generateRequestInfo(DeleteOperation, "resource-id"),
			result: generateRequestInfoMap(operationTypeToName[DeleteOperation], "resource-id"),
			err:    nil,
		},
	}

	for testNum, test := range testCases {
		res, err := requestInfoToMap(test.params)
		if err != nil && test.err == nil {
			t.Errorf("Error not equal in test num %d expected nil, getted %s", testNum+1, err.Error())
			t.Fail()
			continue
		}

		if err == nil && err != test.err {
			t.Errorf("Error not equal in test num %d expected %s, getted nil", testNum+1, test.err)
			t.Fail()
			continue
		}

		if !reflect.DeepEqual(res, test.result) {
			t.Errorf("Result not equal in test num %d expected %+v, result %+v", testNum+1, test.result, res)
		}
	}
}

func TestRequestInfoMetadataAnnotator(t *testing.T) {
	// Test Read (GET with object UUID);
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/app-name/resource-type/resource-id", nil)
	expMetadata := metadata.New(generateRequestInfoMap(operationTypeToName[ReadOperation], "resource-id"))
	md := MetadataAnnotator(context.Background(), req)
	if !reflect.DeepEqual(expMetadata, md) {
		t.Fail()
	}
}

func MustRequest(method, url string, body io.Reader) *http.Request {
	if req, err := http.NewRequest(method, url, body); err == nil {
		return req
	} else {
		panic(err)
	}
}
