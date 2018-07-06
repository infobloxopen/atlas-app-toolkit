package gateway

import (
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

func TestUnsetOp(t *testing.T) {
	page := new(query.PageInfo)
	res := &testResponse{PageInfo: &query.PageInfo{Offset: 30, Size: 10}}

	if err := unsetOp(res, page); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if page.GetOffset() != 30 {
		t.Errorf("invalid repsponse offset: %d - expected: 30", page.GetOffset())
	}
	if page.GetSize() != 10 {
		t.Errorf("invalid repsponse size: %d - expected: 10", page.GetSize())
	}

	// nil operator
	err := unsetOp(res, nil)
	if err == nil {
		t.Fatalf("unexpected non error result - expected: %s", "operator is not a pointer - invalid")
	}
	if err.Error() != "operator is not a pointer - invalid" {
		t.Errorf("invalid error: %s - expected: %s", err, "operator is not a pointer - invalid")
	}

	// nil response
	err = unsetOp(nil, nil)
	if err == nil {
		t.Fatalf("unexpected non error result - expected: %s", "response is not a pointer - invalid")
	}
	if err.Error() != "response is not a pointer - invalid" {
		t.Errorf("invalid error: %s - expected: %s", err, "response is not a pointer - invalid")
	}

	// non struct response
	var i int
	err = unsetOp(&i, nil)
	if err == nil {
		t.Fatalf("unexpected non error result - expected: %s", "response value is not a struct - int")
	}
	if err.Error() != "response value is not a struct - int" {
		t.Errorf("invalid error: %s - expected: %s", err, "response value is not a struct - int")
	}
}
