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

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

type user struct {
	Name string `json:"user"`
	Age  int    `json:"age"`
}

type result struct {
	Users []*user `json:"users"`
}

type userWithPtr struct {
	PtrValue *wrappers.Int64Value `json:"ptr_value"`
}

func (m *userWithPtr) Reset()         {}
func (m *userWithPtr) ProtoMessage()  {}
func (m *userWithPtr) String() string { return "" }

type userWithPtrResult struct {
	Results *userWithPtr `json:"results"`
}

func (m *userWithPtrResult) Reset()         {}
func (m *userWithPtrResult) ProtoMessage()  {}
func (m *userWithPtrResult) String() string { return "" }

func (m *result) Reset()         {}
func (m *result) ProtoMessage()  {}
func (m *result) String() string { return "" }

type badresult struct {
	Success []*user `json:"success"`
}

func (m *badresult) Reset()         {}
func (m *badresult) ProtoMessage()  {}
func (m *badresult) String() string { return "" }

type response struct {
	Error   []map[string]interface{} `json:"error,omitempty"`
	Result  []*user                  `json:"users"`
	Success map[string]interface{}   `json:"success"`
}

func TestForwardResponseMessage(t *testing.T) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.Encode(map[string]interface{}{"code": CodeName(Created), "status": 201})
	md := runtime.ServerMetadata{
		HeaderMD: metadata.Pairs(
			"status-code", CodeName(Created),
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
	ForwardResponseMessage(ctx, nil, &runtime.JSONBuiltin{}, rw, nil, &result{Users: []*user{{"Poe", 209}, {"Hemingway", 119}}})

	if rw.Code != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", rw.Code, http.StatusCreated)
	}

	if ct := rw.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("invalid content-type: %s - expected: %s", ct, "application/json")
	}

	v := &response{}
	if err := json.Unmarshal(rw.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %s", err)
	}

	if v.Success["code"] != CodeName(Created) {
		t.Errorf("invalid status code: %s - expected: %s", v.Success["code"], CodeName(Created))
	}

	if v.Success["status"].(float64) != http.StatusCreated {
		t.Errorf("invalid http status code: %d - expected: %d", v.Success["status"], http.StatusCreated)
	}

	if v.Success["message"] != "created 1 item" {
		t.Errorf("invalid status message: %s - expected: %s", v.Success["message"], "created 1 item")
	}

	if l := len(v.Result); l != 2 {
		t.Fatalf("invalid number of items in response result: %d - expected: %d", l, 2)
	}

	poe, hemingway := v.Result[0], v.Result[1]
	if poe.Name != "Poe" || poe.Age != 209 {
		t.Errorf("invalid result item: %+v - expected: %+v", poe, &user{"Poe", 209})
	}

	if hemingway.Name != "Hemingway" || hemingway.Age != 119 {
		t.Errorf("invalid result item: %+v - expected: %+v", hemingway, &user{"Hemingway", 119})
	}
}

func TestForwardResponseMessageWithNil(t *testing.T) {
	ctx := runtime.NewServerMetadataContext(context.Background(), runtime.ServerMetadata{})

	rw := httptest.NewRecorder()
	ForwardResponseMessage(
		ctx, nil, &runtime.JSONPb{OrigName: true, EmitDefaults: true}, rw, nil,
		&userWithPtrResult{Results: &userWithPtr{PtrValue: nil}},
	)

	var v map[string]interface{}

	if err := json.Unmarshal(rw.Body.Bytes(), &v); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %s", err)
	}

	if len(v["Results"].(map[string]interface{})) != 1 {
		t.Errorf("invalid result item: %+v - expected %+v", v["Results"], map[string]interface{}{})
	}
}

func TestForwardResponseMessageWithSuccessField(t *testing.T) {
	ctx := runtime.NewServerMetadataContext(context.Background(), runtime.ServerMetadata{})

	rw := httptest.NewRecorder()
	ForwardResponseMessage(
		ctx, nil, &runtime.JSONBuiltin{}, rw, nil,
		&badresult{Success: []*user{{"Poe", 209}, {"Hemingway", 119}}},
	)

	var v map[string][]*user
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
		t.Errorf("invalid response item: %+v - expected: %+v", u, &user{"Poe", 209})
	}
	if u := l[1]; u.Name != "Hemingway" || u.Age != 119 {
		t.Errorf("invalid response item: %+v - expected: %+v", u, &user{"Hemingway", 119})
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
	items := []*result{
		{[]*user{{"Poe", 209}}},
		{[]*user{{"Hemingway", 119}}},
	}
	recv := func() (proto.Message, error) {
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

	var rv *result
	// test Poe
	if err := dec.Decode(&rv); err != nil {
		t.Fatalf("failed to unmarshal response chunked result: %s", err)
	}
	if len(rv.Users) != 1 {
		t.Fatalf("invalid number of items in chuncked result: %d - expected: %d", len(rv.Users), 1)
	}
	if u := rv.Users[0]; u.Name != "Poe" || u.Age != 209 {
		t.Errorf("invalid item from chuncked result: %+v - expected: %+v", u, &user{"Poe", 209})
	}

	// test Hemingway
	if err := dec.Decode(&rv); err != nil {
		t.Fatalf("failed to unmarshal response chunked result: %s", err)
	}
	if len(rv.Users) != 1 {
		t.Fatalf("invalid number of items in chuncked result: %d - expected: %d", len(rv.Users), 1)
	}
	if u := rv.Users[0]; u.Name != "Hemingway" || u.Age != 119 {
		t.Errorf("invalid item from chuncked result: %+v - expected: %+v", u, &user{"Hemingway", 119})
	}
}
