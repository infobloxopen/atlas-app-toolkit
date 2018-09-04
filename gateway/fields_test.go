package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/query"
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

	req, _ := http.NewRequest("GET", "http://example.com?_fields=y", nil)

	ctx := context.Background()
	retainFields(ctx, req, indata)

	if !reflect.DeepEqual(indata, expdata) {
		t.Errorf("Unexpected result %v while expecting %v", indata, expdata)
	}

}

func TestRetainSingleResult(t *testing.T) {
	data := `
	{
		"result":
		  {
			"x": "3",
			"y": "4",
			"z": "5"
		  }
	 }`

	expected := `
	 {
		 "result":
		   {
			 "y": "4"
		   }
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

	req, _ := http.NewRequest("GET", "http://example.com?_fields=y", nil)

	ctx := context.Background()
	retainFields(ctx, req, indata)

	if !reflect.DeepEqual(indata, expdata) {
		t.Errorf("Unexpected result %v while expecting %v", indata, expdata)
	}
}

func TestRetainArray(t *testing.T) {
	data := `
	{
		"result": [
		  {
			"x": "1",
			"y": [
				{
				 "a": "2", 
				 "b": "3"
				},
				{
				 "a": "22", 
				 "b": "33"
				}
			 ]
		  },
		  {
			"x": "4",
			"y": [
				{
				 "a": "5", 
				 "b": "6"
				}
			 ],
			"z": "5"
		  }
		]
	 }`

	expected := `
	 {
		 "result": [
		   {
			 "y": [
				{
				 "b":"3"
				},
				 {
				 "b":"33"
				 }
			  ]
		   },
		   {
			 "y": [
				{
				 "b":"6"
				}
			 ]
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

	req, _ := http.NewRequest("GET", "http://example.com?_fields=y.b", nil)

	ctx := context.Background()
	retainFields(ctx, req, indata)

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
			"arr":[
			  {"one":"v1",
			   "two":"v2",
			   "three":"v3"
		      },
			  {"one":"v11",
			   "two":"v22",
			   "three":"v33"
		      }
			],
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
			"arr":[
			  {"one":"v1",
			   "two":"v2",
			   "three":"v3"
		      },
			  {"one":"v11",
			   "two":"v22",
			   "three":"v33"
		      }
			],
		   "e":"eee",
		   "r":"rrr"
		},
		"z":"zzz",
		"q":"qqq"
     }
	`)

	ensureRetain(t, data, "a.b.c,a.b.d,a.e,z,a.arr.one", `
		{
			"a":{
			   "b":{
				  "c":"ccc",
				  "d":"ddd"
				  },
			"arr":[
			  {"one":"v1"
			  },
			  {"one":"v11"
			  }
			],
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
				"b": {}
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
	doRetainFields(indata, flds.GetFields())

	if !reflect.DeepEqual(indata, expdata) {
		t.Errorf("Filtering input %s on fields %s returned %v while expecting %v", input, fields, indata, expdata)
		return
	}
}
