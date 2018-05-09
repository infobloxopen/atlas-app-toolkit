package templates

import (
	"testing"
)

func TestServiceName(t *testing.T) {
	var tests = []struct {
		appname  string
		expected string
		err      error
	}{
		{"ddi.dns.config", "DdiDnsConfig", nil},
		{"ddi-dns_config&007!", "DdiDnsConfig007", nil},
		{"addressbook", "Addressbook", nil},
		{"addressBook", "AddressBook", nil},
		{"addressøB00k", "AddressB00k", nil},
		{"", "", errEmptyServiceName},
		{"4ddressBook", "4ddressBook", errInvalidFirstRune},
	}
	for _, test := range tests {
		name, err := ServiceName(test.appname)
		if err != test.err {
			t.Log(err)
			t.Errorf("Unexpected formatting error: %s - expected %s", name, test.expected)
		}
		if name != test.expected {
			t.Errorf("Unexpected service name: %s - expected %s", name, test.expected)
		}
	}
}

func TestServerURL(t *testing.T) {
	var tests = []struct {
		appname  string
		expected string
		err      error
	}{
		{"ddi.dns.config", "ddi-dns-config", nil},
		{"ddi-dns_config&007!", "ddi-dns-config-007", nil},
		{"addressbook", "addressbook", nil},
		{"addressBook", "addressbook", nil},
		{"addressøB00k", "address-b00k", nil},
		{"", "", errEmptyServiceName},
		{"4ddressbook", "4ddressbook", nil},
	}
	for _, test := range tests {
		name, err := ServerURL(test.appname)
		if err != test.err {
			t.Errorf("Unexpected formatting error: %v - expected %v", err, test.err)
		}
		if name != test.expected {
			t.Errorf("Unexpected service name: %s - expected %s", name, test.expected)
		}
	}
}

func TestProjectRoot(t *testing.T) {
	var tests = []struct {
		gopath   string
		workdir  string
		expected string
		err      error
	}{
		{
			gopath:   "/Users/person/go",
			workdir:  "/Users/person/go/src",
			expected: "",
			err:      nil,
		},
		{
			gopath:   "/Users/person/go",
			workdir:  "/Users/person/go/src/SomeProject",
			expected: "SomeProject",
			err:      nil,
		},
		{
			gopath:   "/Users/person/go_projects",
			workdir:  "/Users/person/go_projects/src/SomeProject",
			expected: "SomeProject",
			err:      nil,
		},
		{
			gopath:   "go",
			workdir:  "go/src/github.com/SomeProject",
			expected: "github.com/SomeProject",
			err:      nil,
		},
		{
			gopath:   "/Users/person/go",
			workdir:  "/Users/person/go/src/github.com/infobloxopen/SomeProject",
			expected: "github.com/infobloxopen/SomeProject",
			err:      nil,
		},
		{
			gopath:   "/Users/person/go",
			workdir:  "/Users/python",
			expected: "",
			err:      errInvalidProjectRoot,
		},
		{
			gopath:   "/Users/person/go",
			workdir:  "/Users/person/hi/src/SomeProject",
			expected: "",
			err:      errInvalidProjectRoot,
		},
		{
			gopath:   "/Users/person/go",
			workdir:  "/Users/person/go",
			expected: "",
			err:      errInvalidProjectRoot,
		},
		{
			gopath:   "",
			workdir:  "/Users/person/go/src/SomeProject",
			expected: "",
			err:      errMissingGOPATH,
		},
	}
	for _, test := range tests {
		root, err := ProjectRoot(test.gopath, test.workdir)
		if root != test.expected {
			t.Errorf("Unexpected service name: %s - expected %s", root, test.expected)
		}
		if err != test.err {
			t.Errorf("Unexpected formatting error: %v - expected %v", err, test.err)
		}
	}
}

func TestIsSpecial(t *testing.T) {
	var tests = []struct {
		r        rune
		expected bool
	}{
		{'a', false},
		{'z', false},
		{'A', false},
		{'Z', false},
		{'0', false},
		{'9', false},
		{'_', true},
		{'.', true},
		{'&', true},
		{'/', true},
		{' ', true},
		{'Σ', true},
	}
	for _, test := range tests {
		if result := isSpecial(test.r); result != test.expected {
			t.Errorf("Unexpected alphanumeric result: %t - expected %t", result, test.expected)
		}
	}
}

func TestDatabaseName(t *testing.T) {
	var tests = []struct {
		appname  string
		expected string
		err      error
	}{
		{
			appname:  "ddi.dns.config",
			expected: "ddi_dns_config",
			err:      nil,
		},
		{
			appname:  "contacts",
			expected: "contacts",
			err:      nil,
		},
		{
			appname:  "_contacts",
			expected: "",
			err:      errInvalidFirstRune,
		},
		{
			appname:  "some&.!@app!name",
			expected: "some_app_name",
			err:      nil,
		},
	}
	for _, test := range tests {
		db, err := DatabaseName(test.appname)
		if err != test.err {
			t.Errorf("Unexpected formatting error: %s - expected %s", err, test.err)
		}
		if db != test.expected {
			t.Errorf("Unexpected service name: %s - expected %s", db, test.expected)
		}
	}
}
