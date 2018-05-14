package query

import (
	"testing"
)

func TestSortCriteria(t *testing.T) {
	c := SortCriteria{"name", SortCriteria_ASC}
	if !c.IsAsc() {
		t.Errorf("invalid sort order: IsAsc = %v - expected: %v", c.IsAsc(), true)
	}
	if c.GoString() != "name ASC" {
		t.Errorf("invalid string representation: %v - expected: %s", c, "name ASC")
	}

	c = SortCriteria{"age", SortCriteria_DESC}
	if !c.IsDesc() {
		t.Errorf("invalid sort order: IsDesc = %v - expected: %v", c.IsDesc(), true)
	}
	if c.GoString() != "age DESC" {
		t.Errorf("invalid string representation: %v - expected: %s", c, "age DESC")
	}
}

func TestParseSorting(t *testing.T) {
	s, err := ParseSorting("name")
	if err != nil {
		t.Fatalf("failed to parse sort parameters: %s", err)
	}
	if len(s.GetCriterias()) != 1 {
		t.Fatalf("invalid number of sort criterias: %d - expected: %d", len(s.GetCriterias()), 1)
	}
	if c := s.GetCriterias()[0]; !c.IsAsc() || c.Tag != "name" {
		t.Errorf("invalid sort criteria: %v - expected: %v", c, SortCriteria{"name", SortCriteria_ASC})
	}

	s, err = ParseSorting("name desc, age")
	if err != nil {
		t.Fatalf("failed to parse sort parameters: %s", err)
	}
	if len(s.GetCriterias()) != 2 {
		t.Fatalf("invalid number of sort criterias: %d - expected: %d", len(s.GetCriterias()), 2)
	}
	if c := s.GetCriterias()[0]; !c.IsDesc() || c.Tag != "name" {
		t.Errorf("invalid sort criteria: %v - expected: %v", c, SortCriteria{"name", SortCriteria_DESC})
	}
	if c := s.GetCriterias()[1]; !c.IsAsc() || c.Tag != "age" {
		t.Errorf("invalid sort criteria: %v - expected: %v", c, SortCriteria{"age", SortCriteria_ASC})
	}
	if s.GoString() != "name DESC, age ASC" {
		t.Errorf("invalid sorting: %v - expected: %s", s, "name DESC, age ASC")
	}

	_, err = ParseSorting("name dask")
	if err == nil {
		t.Fatal("expected error - got nil")
	}
	if err.Error() != "invalid sort order - \"dask\" in \"name dask\"" {
		t.Errorf("invalid error message: %s - expected: %s", err, "invalid sort order - \"dask\" in \"name dask\"")
	}
}
