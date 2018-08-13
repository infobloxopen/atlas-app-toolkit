package util

import "testing"

// TestUnaryServerInterceptor_ValidationErrors will run mock validation errors to see if it parses correctly.
func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		// Test cases
		{
			"Testing CamelCase",
			"CamelCase",
			"camel_case",
		},
		{
			"Testing AnotherCamel123",
			"AnotherCamel123",
			"another_camel123",
		},
		{
			"Testing testCase",
			"testCase",
			"test_case",
		},
		{
			"Testing testcase",
			"testcase",
			"testcase",
		},
		{
			"Testing JSONData",
			"TestCaseUUID",
			"test_case_uuid",
		},
		{
			"Testing JSONData",
			"JSONData",
			"json_data",
		},
	}
	for _, tt := range tests {
		actual := CamelToSnake(tt.actual)
		expected := tt.expected
		if actual != expected {
			t.Errorf("CamelToSnake failed for test %s, expected: \"%s\", actual: \"%s\"", tt.name, expected, actual)
		}

	}
}
