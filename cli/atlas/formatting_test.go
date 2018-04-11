package main

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
		path     string
		expected string
		err      error
	}{
		{
			"go/src/ProjectRoot",
			"ProjectRoot",
			nil,
		},
		{
			"go/src/github.com/secret_project",
			"github.com/secret_project",
			nil,
		},
		{
			"/Users/john/go/src/github.com/infobloxopen/helloWorld",
			"github.com/infobloxopen/helloWorld",
			nil,
		},
		{
			"/Users/john/go",
			"",
			errInvalidProjectRoot,
		},
		{
			"go/src",
			"",
			errInvalidProjectRoot,
		},
	}
	for _, test := range tests {
		root, err := ProjectRoot(test.path)
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
