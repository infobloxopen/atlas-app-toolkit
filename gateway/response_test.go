package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	gateway_test "github.com/infobloxopen/atlas-app-toolkit/gateway/internal"
)

type response struct {
	Error   []map[string]interface{} `json:"error,omitempty"`
	Result  []*gateway_test.User     `json:"users"`
	Success map[string]interface{}   `json:"success"`
}

func TestForwardResponseMessage(t *testing.T) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.Encode(map[string]interface{}{"code": 201, "status": CodeName(Created)})
	md := runtime.ServerMetadata{
		HeaderMD: metadata.Pairs(
			"grpcgateway-status-code", CodeName(Created),
		),
		TrailerMD: metadata.Pairs(
			"success-1", "message:deleted 1 item",
			"success-1", fmt.Sprintf("fields:%q", string(b.Bytes())),
			"success-5", "message:created 1 item",
			"success-5", fmt.Sprintf("fields:%q", string(b.Bytes())),
		),
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	rw := httptest.NewRecorder()
	ForwardResponseMessage(ctx, nil, &runtime.JSONBuiltin{}, rw, nil, &gateway_test.Result{Users: []*gateway_test.User{{Name: "Poe", Age: 209}, {Name: "Hemingway", Age: 119}}})

	if rw.Code != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", rw.Code, http.StatusCreated)
	}

	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}

	mdSt := "Grpc-Metadata-Grpcgateway-Status-Code"
	if h := rw.Header().Get(mdSt); h != "" {
		t.Errorf("got %s: %s", mdSt, h)
	}

	v := &response{}
	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %s", err)
	}

	if v.Success["status"] != CodeName(Created) {
		t.Errorf("invalid status string: %s - expected: %s", v.Success["status"], CodeName(Created))
	}

	if v.Success["code"].(float64) != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", v.Success["code"], http.StatusCreated)
	}

	if v.Success["message"] != "created 1 item" {
		t.Errorf("invalid status message: %s - expected: %s", v.Success["message"], "created 1 item")
	}

	if l := len(v.Result); l != 2 {
		t.Fatalf("invalid number of items in response result: %d - expected: %d", l, 2)
	}

	poe, hemingway := v.Result[0], v.Result[1]
	if poe.Name != "Poe" || poe.Age != 209 {
		t.Errorf("invalid result item: %+v - expected: %+v", poe, &gateway_test.User{Name: "Poe", Age: 209})
	}

	if hemingway.Name != "Hemingway" || hemingway.Age != 119 {
		t.Errorf("invalid result item: %+v - expected: %+v", hemingway, &gateway_test.User{Name: "Hemingway", Age: 119})
	}
}

func TestForwardResponseMessageWithDetailsIncluded(t *testing.T) {
	IncludeStatusDetails(true)
	defer IncludeStatusDetails(false)
	md := runtime.ServerMetadata{
		HeaderMD: metadata.Pairs(
			"grpcgateway-status-code", CodeName(Created),
		),
		TrailerMD: metadata.Pairs(
			"success-1", "message:deleted 1 item",
			"success-5", "message:created 1 item",
		),
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	rw := httptest.NewRecorder()
	ForwardResponseMessage(ctx, nil, &runtime.JSONBuiltin{}, rw, nil, &gateway_test.Result{Users: []*gateway_test.User{{Name: "Poe", Age: 209}, {Name: "Hemingway", Age: 119}}})

	if rw.Code != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", rw.Code, http.StatusCreated)
	}

	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}

	mdSt := "Grpc-Metadata-Grpcgateway-Status-Code"
	if h := rw.Header().Get(mdSt); h != "" {
		t.Errorf("got %s: %s", mdSt, h)
	}

	v := &response{}
	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %s", err)
	}

	if v.Success["status"] != CodeName(Created) {
		t.Errorf("invalid status string: %s - expected: %s", v.Success["status"], CodeName(Created))
	}

	if v.Success["code"].(float64) != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", v.Success["code"], http.StatusCreated)
	}

	if v.Success["message"] != "created 1 item" {
		t.Errorf("invalid status message: %s - expected: %s", v.Success["message"], "created 1 item")
	}

	if l := len(v.Result); l != 2 {
		t.Fatalf("invalid number of items in response result: %d - expected: %d", l, 2)
	}

	poe, hemingway := v.Result[0], v.Result[1]
	if poe.Name != "Poe" || poe.Age != 209 {
		t.Errorf("invalid result item: %+v - expected: %+v", poe, &gateway_test.User{Name: "Poe", Age: 209})
	}

	if hemingway.Name != "Hemingway" || hemingway.Age != 119 {
		t.Errorf("invalid result item: %+v - expected: %+v", hemingway, &gateway_test.User{Name: "Hemingway", Age: 119})
	}
}

func TestForwardResponseMessageWithErrorsAndDetailsIncluded(t *testing.T) {
	IncludeStatusDetails(true)
	defer IncludeStatusDetails(false)
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.Encode(map[string]interface{}{"bar": 2})
	md := runtime.ServerMetadata{
		HeaderMD: metadata.Pairs(
			"grpcgateway-status-code", CodeName(Created),
		),
		TrailerMD: metadata.Pairs(
			"error-1", "message:err message",
			"error-1", fmt.Sprintf("fields:%q", string(b.Bytes()))),
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	rw := httptest.NewRecorder()
	ForwardResponseMessage(ctx, nil, &runtime.JSONBuiltin{}, rw, nil, &gateway_test.Result{Users: []*gateway_test.User{{Name: "Poe", Age: 209}, {Name: "Hemingway", Age: 119}}})

	if rw.Code != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", rw.Code, http.StatusCreated)
	}

	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}

	mdSt := "Grpc-Metadata-Grpcgateway-Status-Code"
	if h := rw.Header().Get(mdSt); h != "" {
		t.Errorf("got %s: %s", mdSt, h)
	}

	v := &response{}
	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %s", err)
	}

	if len(v.Error) != 1 {
		t.Errorf("did not contain expected error in response")
	}

	if v.Success["status"] != CodeName(Created) {
		t.Errorf("invalid status string: %s - expected: %s", v.Success["status"], CodeName(Created))
	}

	if v.Success["code"].(float64) != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", v.Success["code"], http.StatusCreated)
	}

	if v.Error[0]["bar"].(float64) != 2 {
		t.Errorf("unexpected err field: %d - expected: %d", v.Error[0]["bar"], 2)
	}

	if v.Error[0]["message"] != "err message" {
		t.Errorf("invalid status message: %s - expected: %s", v.Error[0]["message"], "err message")
	}

	if l := len(v.Result); l != 2 {
		t.Fatalf("invalid number of items in response result: %d - expected: %d", l, 2)
	}

	poe, hemingway := v.Result[0], v.Result[1]
	if poe.Name != "Poe" || poe.Age != 209 {
		t.Errorf("invalid result item: %+v - expected: %+v", poe, &gateway_test.User{Name: "Poe", Age: 209})
	}

	if hemingway.Name != "Hemingway" || hemingway.Age != 119 {
		t.Errorf("invalid result item: %+v - expected: %+v", hemingway, &gateway_test.User{Name: "Hemingway", Age: 119})
	}
}

func TestForwardResponseMessageWithNil(t *testing.T) {
	ctx := runtime.NewServerMetadataContext(context.Background(), runtime.ServerMetadata{})

	rw := httptest.NewRecorder()
	ForwardResponseMessage(
		ctx, nil, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
		}, rw, nil,
		&gateway_test.UserWithPtrResult{Result: &gateway_test.UserWithPtr{PtrValue: nil}},
	)

	var v map[string]interface{}

	if err := json.Unmarshal(rw.Body.Bytes(), &v); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %s", err)
	}
	fmt.Printf("%+v %s", v, rw.Body.Bytes())

	if len(v["result"].(map[string]interface{})) != 1 {
		t.Errorf("invalid result item: %+v - expected %+v", v["result"], map[string]interface{}{})
	}
}

func TestForwardResponseMessageWithSuccessField(t *testing.T) {
	ctx := runtime.NewServerMetadataContext(context.Background(), runtime.ServerMetadata{})

	rw := httptest.NewRecorder()
	ForwardResponseMessage(
		ctx, nil, &runtime.JSONBuiltin{}, rw, nil,
		&gateway_test.BadResult{Success: []*gateway_test.User{{Name: "Poe", Age: 209}, {Name: "Hemingway", Age: 119}}},
	)

	var v map[string][]*gateway_test.User
	if err := json.Unmarshal(rw.Body.Bytes(), &v); err != nil {
		t.Fatalf("failed to unmarshal response: %s", err)
	}
	l, ok := v["success"]
	if !ok {
		t.Fatal("invalid response: missing 'success' field")
	}
	if len(l) != 2 {
		t.Fatalf("invalid number of items in response: %d - expected: %d", len(l), 2)
	}
	if u := l[0]; u.Name != "Poe" || u.Age != 209 {
		t.Errorf("invalid response item: %+v - expected: %+v", u, &gateway_test.User{Name: "Poe", Age: 209})
	}
	if u := l[1]; u.Name != "Hemingway" || u.Age != 119 {
		t.Errorf("invalid response item: %+v - expected: %+v", u, &gateway_test.User{Name: "Hemingway", Age: 119})
	}
}

func TestForwardResponseStream(t *testing.T) {
	md := runtime.ServerMetadata{
		HeaderMD: metadata.Pairs(
			runtime.MetadataPrefix+"status-message", "returned 1 item",
		),
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	rw := httptest.NewRecorder()

	count := 0
	items := []*gateway_test.Result{
		{Users: []*gateway_test.User{{Name: "Poe", Age: 209}}},
		{Users: []*gateway_test.User{{Name: "Hemingway", Age: 119}}},
	}
	recv := func() (protoreflect.ProtoMessage, error) {
		if count < len(items) {
			i := items[count]
			count++
			return i, nil
		}
		return nil, io.EOF
	}

	ForwardResponseStream(ctx, nil, &runtime.JSONBuiltin{}, rw, nil, recv)

	// if not set explicitly should be set by default
	if rw.Code != http.StatusPartialContent {
		t.Errorf("invalid http status code:%d - expected: %d", rw.Code, http.StatusPartialContent)
	}
	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}
	if te := rw.Header().Get("Transfer-Encoding"); te != "chunked" {
		t.Errorf("invalid transfer-encoding: %s - expected: %s", te, "chunked")
	}

	dec := json.NewDecoder(rw.Body)

	var rv *gateway_test.Result
	// test Poe
	if err := dec.Decode(&rv); err != nil {
		t.Fatalf("failed to unmarshal response chunked result: %s", err)
	}
	if len(rv.Users) != 1 {
		t.Fatalf("invalid number of items in chuncked result: %d - expected: %d", len(rv.Users), 1)
	}
	if u := rv.Users[0]; u.Name != "Poe" || u.Age != 209 {
		t.Errorf("invalid item from chuncked result: %+v - expected: %+v", u, &gateway_test.User{Name: "Poe", Age: 209})
	}

	// test Hemingway
	if err := dec.Decode(&rv); err != nil {
		t.Fatalf("failed to unmarshal response chunked result: %s", err)
	}
	if len(rv.Users) != 1 {
		t.Fatalf("invalid number of items in chuncked result: %d - expected: %d", len(rv.Users), 1)
	}
	if u := rv.Users[0]; u.Name != "Hemingway" || u.Age != 119 {
		t.Errorf("invalid item from chuncked result: %+v - expected: %+v", u, &gateway_test.User{Name: "Hemingway", Age: 119})
	}
}
