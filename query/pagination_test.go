package query

import (
	"testing"
)

func TestParsePagination(t *testing.T) {
	// invalid limit
	_, err := ParsePagination("1s", "0", "ptoken")
	if err == nil {
		t.Error("unexpected nil error - expected: pagination: limit - invalid syntax")
	}
	if err.Error() != "pagination: limit - invalid syntax" {
		t.Errorf("invalid error: %s - expected: pagination: limit - invalid syntax", err)
	}

	// negative limit
	_, err = ParsePagination("-1", "0", "ptoken")
	if err == nil {
		t.Error("unexpected nil error - expected: pagination: limit - negative value")
	}
	if err.Error() != "pagination: limit - negative value" {
		t.Errorf("invalid error: %s - expected: pagination: limit - negative value", err)
	}

	// invalid offset
	_, err = ParsePagination("0", "0w", "ptoken")
	if err == nil {
		t.Error("unexpected nil error - expected: pagination: offset - invalid syntax")
	}
	if err.Error() != "pagination: offset - invalid syntax" {
		t.Errorf("invalid error: %s - expected: pagination: offset - invalid syntax", err)
	}

	// negative offset
	_, err = ParsePagination("0", "-1", "ptoken")
	if err == nil {
		t.Error("unexpected nil error - expected: pagination: offset - negative value")
	}
	if err.Error() != "pagination: offset - negative value" {
		t.Errorf("invalid error: %s - expected: pagination: offset - negative value", err)
	}

	// null offset
	p, err := ParsePagination("0", "null", "ptoken")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p.GetOffset() != 0 {
		t.Errorf("invalid offset: %d - expected: 0", p.GetOffset())
	}

	// first page
	p, err = ParsePagination("0", "0", "ptoken")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if !p.FirstPage() {
		t.Errorf("invalid value of first page: %v - expected: true", p.FirstPage())
	}
	p, err = ParsePagination("0", "100", "null")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if !p.FirstPage() {
		t.Errorf("invalid value of first page: %v - expected: true", p.FirstPage())
	}

	// default limit
	if p.DefaultLimit(1000) != 1000 {
		t.Errorf("invalid default limit: %d - expected: 1000", p.DefaultLimit(1000))
	}

	// valid pagination
	p, err = ParsePagination("1000", "100", "ptoken")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if p.GetLimit() != 1000 {
		t.Errorf("invalid limit: %d - expected: 1000", p.GetLimit())
	}
	if p.GetOffset() != 100 {
		t.Errorf("invalid offset: %d - expected: 100", p.GetOffset())
	}
	if p.GetPageToken() != "ptoken" {
		t.Errorf("invalid page token: %q - expected: ptoken", p.GetPageToken())
	}
}

func TestPageInfo(t *testing.T) {
	p := new(PageInfo)
	if p.NoMore() {
		t.Errorf("invalid value of NoMore: %v - expected: false", p.NoMore())
	}
	p.SetLastOffset()
	if !p.NoMore() {
		t.Errorf("invalid value of NoMore: %v - expected: true", p.NoMore())
	}

	p = new(PageInfo)
	if p.NoMore() {
		t.Errorf("invalid value of NoMore: %v - expected: false", p.NoMore())
	}
	p.SetLastToken()
	if !p.NoMore() {
		t.Errorf("invalid value of NoMore: %v - expected: true", p.NoMore())
	}
}
