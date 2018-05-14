package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"google.golang.org/grpc/metadata"
)

func TestRetain(t *testing.T) {
	data := `
	{
		"result": [
		  {
			"x": "1",
			"y": "2"
		  },
		  {
			"x": "3",
			"y": "4",
			"z": "5"
		  }
		]
	 }`

	expected := `
	 {
		 "result": [
		   {
			 "y": "2"
		   },
		   {
			 "y": "4"
		   }
		 ]
	  }`

	var indata map[string]interface{}
	err := json.Unmarshal([]byte(data), &indata)
	if err != nil {
		t.Errorf("Error parsing test input %s", data)
		return
	}

	var expdata map[string]interface{}
	err = json.Unmarshal([]byte(expected), &expdata)
	if err != nil {
		t.Errorf("Error parsing test expected result %s", expected)
		return
	}

	md := runtime.ServerMetadata{
		HeaderMD: metadata.Pairs(
			runtime.MetadataPrefix+fieldsMetaKey, "y",
		),
	}
	ctx := runtime.NewServerMetadataContext(context.Background(), md)
	retainFields(ctx, nil, indata)

	if !reflect.DeepEqual(indata, expdata) {
		t.Errorf("Unexpected result %v while expecting %v", indata, expdata)
	}

}

func TestDoRetain(t *testing.T) {
	data := `
	{
		"a":{
		   "b":{
			  "c":"ccc",
			  "d":"ddd",
			  "x":"xxx"
		   },
		   "e":"eee",
		   "r":"rrr"
		},
		"z":"zzz",
		"q":"qqq"
	 }`

	ensureRetain(t, data, "", `
		{
			"a":{
				"b":{
				   "c":"ccc",
				   "d":"ddd",
				   "x":"xxx"
				},
				"e":"eee",
				"r":"rrr"
			 },
			 "z":"zzz",
			 "q":"qqq"
	     }
	`)

	ensureRetain(t, data, "a.b.c,a.b.d,a.e,z", `
		{
			"a":{
			   "b":{
				  "c":"ccc",
				  "d":"ddd"
			   },
			   "e":"eee"
			},
			"z":"zzz"
		 }
	`)

	ensureRetain(t, data, "a.b", `
		{
			"a":{
				"b":{
					"c":"ccc",
					"d":"ddd",
					"x":"xxx"
				 }
			}
		 }
	`)

	ensureRetain(t, data, "q", `
		{
			"q":"qqq"
		 }
	`)

	ensureRetain(t, data, "a.e,z", `
		{
			"a":{
			   "e":"eee"
			},
			"z":"zzz"
		 }
	`)

	ensureRetain(t, data, "a.mmm,vvv", `
		{
			"a":{}
		 }
	`)

	ensureRetain(t, data, "q.bbb", `
		{
			"q":"qqq"
		 }
	`)

	ensureRetain(t, data, "a.b.mmm", `
		{
			"a":{
				"b":{}
			}
		 }
	`)

}

func ensureRetain(t *testing.T, input, fields, expected string) {
	var indata map[string]interface{}
	err := json.Unmarshal([]byte(input), &indata)
	if err != nil {
		t.Errorf("Error parsing test input %s", input)
		return
	}

	var expdata map[string]interface{}
	err = json.Unmarshal([]byte(expected), &expdata)
	if err != nil {
		t.Errorf("Error parsing test expected result %s", expected)
		return
	}

	flds := query.ParseFieldSelection(fields)
	doRetainFields(indata, flds.Fields)

	if !reflect.DeepEqual(indata, expdata) {
		t.Errorf("Filtering input %s on fields %s returned %v while expecting %v", input, fields, indata, expdata)
		return
	}
}

func TestFieldSelection(t *testing.T) {
	// fields parameters is not specified
	req, err := http.NewRequest(http.MethodGet, "http://app.com?someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md := MetadataAnnotator(context.Background(), req)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	flds := FieldSelection(ctx)
	if flds != nil {
		t.Fatalf("unexpected fields result: %v, expected nil", flds)
	}

	// fields parameters is specified
	req, err = http.NewRequest(http.MethodGet, "http://app.com?_fields=name,address.street&someparam=1", nil)
	if err != nil {
		t.Fatalf("failed to build new http testRequest: %s", err)
	}

	md = MetadataAnnotator(context.Background(), req)
	ctx = metadata.NewIncomingContext(context.Background(), md)

	flds = FieldSelection(ctx)
	expected := &query.FieldSelection{Fields: query.FieldSelectionMap{"name": &query.Field{Name: "name"}, "address": &query.Field{Name: "address", Subs: query.FieldSelectionMap{"street": &query.Field{Name: "street"}}}}}
	if !reflect.DeepEqual(flds, expected) {
		t.Errorf("Unexpected result %v while expecting %v", flds, expected)
	}
}
