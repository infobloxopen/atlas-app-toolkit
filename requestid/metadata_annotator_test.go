package requestid

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"google.golang.org/grpc/metadata"
)

func generateRequestID(requestId string) map[string]string {
	mdmap := map[string]string{
		DefaultRequestIDKey: requestId,
	}
	return mdmap
}

func TestRequestIDMetadataAnnotator(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/myrequest/", nil)
	req.Header.Add("Request-Id", "my-request")
	f := NewRequestIDAnnotator()
	expMetadata := metadata.New(generateRequestID("my-request"))
	resultMD := f(context.Background(), req)
	if !reflect.DeepEqual(expMetadata, resultMD) {
		t.Errorf("Result: %v Expected: %v", resultMD, expMetadata)
	}
}
