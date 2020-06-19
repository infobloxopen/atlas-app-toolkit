package tracing

import (
	"testing"
)

func Test_truncatePayload(t *testing.T) {
	tests := []struct {
		in      []byte
		out     []byte
		outFlag bool

		limit int
	}{
		{
			in:      []byte("Hello World"),
			out:     []byte("Hello World"),
			outFlag: false,
			limit:   10000000,
		},
		{
			in:      []byte("Hello World"),
			out:     []byte("He..."),
			outFlag: true,
			limit:   5,
		},
		{
			in:      []byte("Hello"),
			out:     []byte("Hello"),
			outFlag: false,
			limit:   5,
		},
	}

	for _, tt := range tests {
		out, reduced := truncatePayload(tt.in, tt.limit)
		if tt.outFlag != reduced {
			t.Errorf("Unexpected result expected %t, got %t", tt.outFlag, reduced)
		}

		if string(out) != string(tt.out) {
			t.Errorf("Unexpected result\n\texpected %q\n\tgot %q", tt.out, out)
		}
	}
}

func Test_obfuscate(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "HelloWorld",
			out: "He...",
		},
		{
			in:  "",
			out: "...",
		},
		{
			in:  "H",
			out: "...",
		},
	}

	for _, tt := range tests {
		out := obfuscate(tt.in)
		if out != tt.out {
			t.Errorf("Unexpected result\n\texpected %q\n\tgot %q", tt.out, out)
		}
	}
}
