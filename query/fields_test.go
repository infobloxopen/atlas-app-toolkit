package query

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	validateParse(t, ParseFieldSelection(""), nil)

	expected := FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a"}}}
	validateParse(t, ParseFieldSelection("a"), &expected)
	validateParse(t, ParseFieldSelection("a", "?"), &expected)

	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b"}, "c": &Field{Name: "c"}}}}}
	validateParse(t, ParseFieldSelection("a.b,a.c"), &expected)
	validateParse(t, ParseFieldSelection("a?b,a?c", "?"), &expected)
	validateParse(t, ParseFieldSelection("a-b,a-c", "-", "?"), &expected)

	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b"}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x"}}}
	validateParse(t, ParseFieldSelection("a.b,a.c,x"), &expected)

	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x"}}}
	validateParse(t, ParseFieldSelection("a.b.v,a.c,x"), &expected)
	validateParse(t, ParseFieldSelection("a,a.b,a.b.v,a.c,x"), &expected)
}

func validateParse(t *testing.T, result *FieldSelection, expected *FieldSelection) {
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected parse result %v while expecting %v", result, expected)
	}
}

func TestGoString(t *testing.T) {
	validateGoString(t, "a,b,c.x,c.y")
	validateGoString(t, "a,b,c.x,c.y.z")
	validateGoString(t, "q.w,e,a,b,c.x,c.y.z,c.y.r")
}

func validateGoString(t *testing.T, data string) {
	original := map[string]bool{}
	for _, x := range strings.Split(data, ",") {
		original[x] = true
	}

	flds := ParseFieldSelection(data)
	fldsStr := flds.GoString()
	result := map[string]bool{}
	for _, x := range strings.Split(fldsStr, ",") {
		result[x] = true
	}

	if !reflect.DeepEqual(result, original) {
		t.Errorf("Unexpected fields-to-string conversion result for %s", data)
	}
}

func TestAdd(t *testing.T) {
	flds := &FieldSelection{}
	flds.Add("test")
	expected := FieldSelection{Fields: FieldSelectionMap{"test": &Field{Name: "test"}}}
	validateParse(t, flds, &expected)

	flds = ParseFieldSelection("a.b,x")

	flds.Add("")
	flds.Add("x")
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b"}}}, "x": &Field{Name: "x"}}}
	validateParse(t, flds, &expected)

	flds.Add("a.c")
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b"}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x"}}}
	validateParse(t, flds, &expected)

	flds.Add("a.b.v")
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x"}}}
	validateParse(t, flds, &expected)

	flds.Add("x.y.z")
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{"y": &Field{Name: "y", Subs: FieldSelectionMap{"z": &Field{Name: "z"}}}}}}}
	validateParse(t, flds, &expected)

	flds.Add("t.i")
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{"y": &Field{Name: "y", Subs: FieldSelectionMap{"z": &Field{Name: "z"}}}}}, "t": &Field{Name: "t", Subs: FieldSelectionMap{"i": &Field{Name: "i"}}}}}
	validateParse(t, flds, &expected)

	flds.Add("k")
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{"y": &Field{Name: "y", Subs: FieldSelectionMap{"z": &Field{Name: "z"}}}}}, "t": &Field{Name: "t", Subs: FieldSelectionMap{"i": &Field{Name: "i"}}}, "k": &Field{Name: "k"}}}
	validateParse(t, flds, &expected)
}

func TestDelete(t *testing.T) {
	flds := ParseFieldSelection("a.b.v,a.c,x.y.z,t.i")
	isDel := flds.Delete("q")
	if isDel == true {
		t.Error("Unexpected delete result")
	}
	isDel = flds.Delete("")
	if isDel == true {
		t.Error("Unexpected delete result")
	}
	isDel = flds.Delete("t.o")
	if isDel == true {
		t.Error("Unexpected delete result")
	}
	isDel = flds.Delete("t.i.o")
	if isDel == true {
		t.Error("Unexpected delete result")
	}

	expected := FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{"y": &Field{Name: "y", Subs: FieldSelectionMap{"z": &Field{Name: "z"}}}}}, "t": &Field{Name: "t", Subs: FieldSelectionMap{"i": &Field{Name: "i"}}}}}
	validateParse(t, flds, &expected)

	isDel = flds.Delete("t")
	if isDel == false {
		t.Error("Unexpected delete result")
	}
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{"y": &Field{Name: "y", Subs: FieldSelectionMap{"z": &Field{Name: "z"}}}}}}}
	validateParse(t, flds, &expected)

	isDel = flds.Delete("x.y")
	if isDel == false {
		t.Error("Unexpected delete result")
	}
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{}}}}
	validateParse(t, flds, &expected)

	isDel = flds.Delete("a.b.v")
	if isDel == false {
		t.Error("Unexpected delete result")
	}
	expected = FieldSelection{Fields: FieldSelectionMap{"a": &Field{Name: "a", Subs: FieldSelectionMap{"b": &Field{Name: "b", Subs: FieldSelectionMap{}}, "c": &Field{Name: "c"}}}, "x": &Field{Name: "x", Subs: FieldSelectionMap{}}}}
	validateParse(t, flds, &expected)

	isDel = flds.Delete("a")
	if isDel == false {
		t.Error("Unexpected delete result")
	}
	expected = FieldSelection{Fields: FieldSelectionMap{"x": &Field{Name: "x", Subs: FieldSelectionMap{}}}}
	validateParse(t, flds, &expected)
}

func TestGet(t *testing.T) {
	flds := ParseFieldSelection("a.b.v,a.c,x.y.z,t.i")
	validateGet(t, flds, "q", nil)
	validateGet(t, flds, "", nil)
	validateGet(t, flds, "t.o", nil)
	validateGet(t, flds, "t.i.o", nil)

	expected := &Field{Name: "i"}
	validateGet(t, flds, "t.i", expected)

	expected = &Field{Name: "b", Subs: FieldSelectionMap{"v": &Field{Name: "v"}}}
	validateGet(t, flds, "a.b", expected)

	expected = &Field{Name: "x", Subs: FieldSelectionMap{"y": &Field{Name: "y", Subs: FieldSelectionMap{"z": &Field{Name: "z"}}}}}
	validateGet(t, flds, "x", expected)
}

func validateGet(t *testing.T, flds *FieldSelection, field string, expected *Field) {
	fld := flds.Get(field)
	if !reflect.DeepEqual(fld, expected) {
		t.Errorf("Unexpected get result for %s", field)
	}
}
