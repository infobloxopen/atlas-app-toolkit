package example

import (
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

func TestFilteringPermissionsValidation(t *testing.T) {
	data := map[string]string{
		"first_name==\"Jan\"":      "Operation EQ does not allowed for 'first_name'",
		"User.first_name!=\"Sam\"": "Operation EQ does not allowed for 'User.first_name'",
		"first_name~\"Sam.*\"":     "",

		"middle_name==\"Jan\"":       "",
		"middle_name!=\"Sam\"":       "",
		"User.middle_name~\"Sam.*\"": "",

		"last_name~\"Jan\"":         "Operation MATCH does not allowed for 'last_name'",
		"User.last_name!~\"Sam\"":   "Operation MATCH does not allowed for 'User.last_name'",
		"last_name==\"Sam.*\"":      "",
		"User.last_name!=\"Sam.*\"": "",

		"age==18":     "Operation EQ does not allowed for 'age'",
		"age>=18":     "",
		"User.age<18": "",

		"height>=180":     "",
		"height>180":      "Operation GT does not allowed for 'height'",
		"User.height<180": "Operation LT does not allowed for 'User.height'",
	}

	for param, expected := range data {
		response := ""
		f, err := query.ParseFiltering(param)
		if err != nil {
			t.Fatalf("Invalid filtering data '%s'", param)
		}
		resp := Validate(f, nil, "/example.TstService/List")
		if resp != nil {
			response = resp.Error()
		}
		if expected != response {
			t.Errorf("Error, for filtering data '%s' expected '%s', got '%s'", param, expected, response)
		}

	}
}

func TestFilteringPermissionsValidationForRead(t *testing.T) {
	data := []string{
		"first_name==\"Jan\"",
		"last_name~\"Jan\"",
		"age==18",
		"height>=180",
		"height<180",
	}

	for _, param := range data {
		f, err := query.ParseFiltering(param)
		if err != nil {
			t.Fatalf("Invalid filtering data '%s'", param)
		}
		resp := Validate(f, nil, "/example.TstService/Read")
		if resp != nil {
			t.Errorf("Error, for filtering data '%s' got '%s'", param, resp.Error())
		}

	}
}

func TestSortingPermissionsValidation(t *testing.T) {
	data := map[string]string{
		"first_name, height, middle_name":                "pagination doesn't allowd for 'middle_name'",
		"User.first_name, User.height, User.middle_name": "pagination doesn't allowd for 'User.middle_name'",
	}

	for param, expected := range data {
		response := ""
		p, err := query.ParseSorting(param)
		if err != nil {
			t.Fatalf("Invalid paging data '%s'", param)
		}
		resp := Validate(nil, p, "/example.TstService/List")
		if resp != nil {
			response = resp.Error()
		}
		if expected != response {
			t.Errorf("Error, for paging data '%s' expected '%s', got '%s'", param, expected, response)
		}
	}
}
