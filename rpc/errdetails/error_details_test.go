package errdetails

import (
	"encoding/json"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestOkCode(t *testing.T) {
	ti := New(codes.OK, "", "")

	if ti.GetCode() != int32(codes.OK) {
		t.Error("code is not OK")
	}

	// test MarshalJSON
	data, err := json.Marshal(ti)
	if err != nil {
		t.Error(err)
	}
	if string(data) != `{"code":"OK"}` {
		t.Errorf("invalid code: %s", data)
	}

	// test UnmarshalJSON
	var tu TargetInfo
	if err := json.Unmarshal(data, &tu); err != nil {
		t.Error(err)
	}

	if ti.GetCode() != tu.GetCode() {
		t.Errorf("invalid code", tu.GetCode())
	}

	// code is not set
	ti.Reset()
	if err := json.Unmarshal([]byte(`{}`), ti); err != nil {
		t.Error(err)
	}
	if ti.GetCode() != int32(codes.OK) {
		t.Errorf("invalid code: %s", ti.GetCode())
	}

	// code is null
	ti.Reset()
	if err := json.Unmarshal([]byte(`{"code": null}`), ti); err != nil {
		t.Error(err)
	}
	if ti.GetCode() != int32(codes.OK) {
		t.Errorf("invalid code: %s", ti.GetCode())
	}

	// code is empty string
	ti.Reset()
	if err := json.Unmarshal([]byte(`{"code": ""}`), ti); err != nil {
		t.Error(err)
	}
	if ti.GetCode() != int32(codes.OK) {
		t.Errorf("invalid code: %s", ti.GetCode())
	}
}

func TestUnimplementedCode(t *testing.T) {
	ti := New(codes.Unimplemented, "", "")

	if ti.GetCode() != int32(codes.Unimplemented) {
		t.Error("code is not Unimplemented")
	}

	// test MarshalJSON
	data, err := json.Marshal(ti)
	if err != nil {
		t.Error(err)
	}
	if string(data) != `{"code":"NOT_IMPLEMENTED"}` {
		t.Errorf("invalid code: %s", data)
	}

	// test UnmarshalJSON
	var tu TargetInfo
	if err := json.Unmarshal(data, &tu); err != nil {
		t.Error(err)
	}

	if ti.GetCode() != tu.GetCode() {
		t.Errorf("invalid code", tu.GetCode())
	}

	ti.Reset()
	if err := json.Unmarshal([]byte(`{"code": "NOT_IMPLEMENTED"}`), ti); err != nil {
		t.Error(err)
	}
	if ti.GetCode() != int32(codes.Unimplemented) {
		t.Errorf("invalid code: %s", ti.GetCode())
	}
}

func TestUnknownCode(t *testing.T) {
	ti := New(codes.Unknown, "", "")

	if ti.GetCode() != int32(codes.Unknown) {
		t.Error("code is not Unimplemented")
	}

	// test MarshalJSON
	data, err := json.Marshal(ti)
	if err != nil {
		t.Error(err)
	}
	if string(data) != `{"code":"UNKNOWN"}` {
		t.Errorf("invalid code: %s", data)
	}

	// test UnmarshalJSON
	var tu TargetInfo
	if err := json.Unmarshal(data, &tu); err != nil {
		t.Error(err)
	}

	if ti.GetCode() != tu.GetCode() {
		t.Errorf("invalid code", tu.GetCode())
	}

	ti.Reset()
	if err := json.Unmarshal([]byte(`{"code": "NEW_CODE"}`), ti); err != nil {
		t.Error(err)
	}
	if ti.GetCode() != int32(codes.Unknown) {
		t.Errorf("invalid code: %s", ti.GetCode())
	}
}
