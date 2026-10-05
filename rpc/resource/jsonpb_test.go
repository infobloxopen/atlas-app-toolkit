package resource

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/golang/protobuf/jsonpb"
)

func TestIdentifier_MarshalJSONPB(t *testing.T) {
	tcases := []struct {
		Identifier         *Identifier
		ExpectedJSONString string
	}{
		{
			&Identifier{
				ApplicationName: "app",
				ResourceType:    "resource",
				ResourceId:      "res1",
			},
			`"app/resource/res1"`,
		},
		{
			&Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			`"null"`,
		},
		{
			&Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: `a"b`},
			`"app/res/a\"b"`,
		},
		{
			&Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: `a\b`},
			`"app/res/a\\b"`,
		},
		{
			&Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "a\nb"},
			`"app/res/a\nb"`,
		},
		{
			&Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "a&b<c>"},
			`"app/res/a&b<c>"`,
		},
	}

	var (
		marshaler = &jsonpb.Marshaler{}
		buffer    = &bytes.Buffer{}
	)

	for _, tc := range tcases {
		buffer.Reset()

		if err := marshaler.Marshal(buffer, tc.Identifier); err != nil {
			t.Errorf("failed to marshal identifier %s - %s", tc.Identifier, err)
		}

		if s := buffer.String(); s != tc.ExpectedJSONString {
			t.Errorf("ivalid identifier %s, expected %s", s, tc.ExpectedJSONString)
		}
	}
}

func TestIdentifier_UnmarshalJSONPB(t *testing.T) {
	const (
		noErr = iota
		typeErr
		syntaxErr
	)

	tcases := []struct {
		name string
		data string
		want Identifier
		err  int
		// directOnly rows are rejected by encoding/json and jsonpb before
		// UnmarshalJSONPB is reached, so only the direct call is meaningful.
		directOnly bool
	}{
		{name: "full identifier", data: `"app/resource/res1"`, want: Identifier{ApplicationName: "app", ResourceType: "resource", ResourceId: "res1"}},
		{name: "type and id", data: `"resource/res1"`, want: Identifier{ResourceType: "resource", ResourceId: "res1"}},
		{name: "id only", data: `"res1"`, want: Identifier{ResourceId: "res1"}},
		{name: "extra segments stay in id", data: `"a/b/c/d"`, want: Identifier{ApplicationName: "a", ResourceType: "b", ResourceId: "c/d"}},
		{name: "null literal", data: `null`},
		{name: "quoted null", data: `"null"`},
		{name: "empty string", data: `""`},
		{name: "empty data", data: ``, directOnly: true},
		{name: "surrounding whitespace", data: ` "app/res/id1" `, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "id1"}},

		{name: "escaped slash", data: `"app\/res\/id1"`, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "id1"}},
		{name: "unicode escaped slash", data: `"app\u002fres\u002fid1"`, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "id1"}},
		{name: "unicode escape", data: `"app/res/\u00e9"`, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "\u00e9"}},
		{name: "escaped quote", data: `"app/res/a\"b"`, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: `a"b`}},
		{name: "escaped backslash", data: `"app/res/a\\b"`, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: `a\b`}},
		{name: "escaped newline", data: `"app/res/a\nb"`, want: Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "a\nb"}},

		{name: "array", data: `["app/resource/id1","app/resource/id2"]`, err: typeErr},
		{name: "empty array", data: `[]`, err: typeErr},
		{name: "number", data: `12345`, err: typeErr},
		{name: "negative number", data: `-1`, err: typeErr},
		{name: "float", data: `1.5`, err: typeErr},
		{name: "object", data: `{"application_name":"app","resource_type":"res","resource_id":"id1"}`, err: typeErr},
		{name: "empty object", data: `{}`, err: typeErr},
		{name: "true", data: `true`, err: typeErr},
		{name: "false", data: `false`, err: typeErr},

		{name: "unterminated string", data: `"app/res/id1`, err: syntaxErr, directOnly: true},
		{name: "unquoted string", data: `app/res/id1`, err: syntaxErr, directOnly: true},
		{name: "two values", data: `"a" "b"`, err: syntaxErr, directOnly: true},
		{name: "truncated null", data: `nul`, err: syntaxErr, directOnly: true},
		{name: "open bracket", data: `[`, err: syntaxErr, directOnly: true},
	}

	check := func(t *testing.T, got Identifier, err error, wantErr int, want Identifier) {
		t.Helper()
		var te *json.UnmarshalTypeError
		var se *json.SyntaxError
		switch wantErr {
		case noErr:
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got.ApplicationName != want.ApplicationName || got.ResourceType != want.ResourceType || got.ResourceId != want.ResourceId {
				t.Errorf("got %q/%q/%q, want %q/%q/%q", got.ApplicationName, got.ResourceType, got.ResourceId,
					want.ApplicationName, want.ResourceType, want.ResourceId)
			}
		case typeErr:
			if !errors.As(err, &te) {
				t.Errorf("expected *json.UnmarshalTypeError, got %v", err)
			}
		case syntaxErr:
			if !errors.As(err, &se) {
				t.Errorf("expected *json.SyntaxError, got %v", err)
			}
		}
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("direct", func(t *testing.T) {
				var id Identifier
				err := id.UnmarshalJSONPB(nil, []byte(tc.data))
				check(t, id, err, tc.err, tc.want)
			})
			if tc.directOnly {
				return
			}
			t.Run("jsonpb", func(t *testing.T) {
				var id Identifier
				err := (&jsonpb.Unmarshaler{}).Unmarshal(strings.NewReader(tc.data), &id)
				if tc.err != noErr {
					if err == nil {
						t.Error("expected error, got nil")
					}
					return
				}
				check(t, id, err, noErr, tc.want)
			})
			t.Run("json", func(t *testing.T) {
				var wrapper struct {
					ID Identifier `json:"id"`
				}
				err := json.Unmarshal([]byte(`{"id":`+tc.data+`}`), &wrapper)
				check(t, wrapper.ID, err, tc.err, tc.want)
			})
		})
	}
}

func TestIdentifier_UnmarshalJSONPB_ErrorMessage(t *testing.T) {
	tcases := []struct {
		name string
		data []byte
	}{
		{"multi-byte characters across the cut", []byte(`["` + strings.Repeat("é", 100) + `"]`)},
		{"invalid UTF-8 input", []byte("[\xff\xfe]")},
	}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			err := (&Identifier{}).UnmarshalJSONPB(nil, tc.data)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !utf8.ValidString(err.Error()) {
				t.Errorf("error text is not valid UTF-8: %q", err.Error())
			}
			if !strings.Contains(err.Error(), "expected a string") {
				t.Errorf("unexpected error text: %q", err.Error())
			}
		})
	}
}

func TestIdentifier_JSONRoundTrip(t *testing.T) {
	ids := []struct {
		name       string
		resourceID string
	}{
		{"plain", "id1"},
		{"path in id", "a/b/c"},
		{"quote", `a"b`},
		{"backslash", `a\b`},
		{"backslash then n", `a\nb`},
		{"newline", "a\nb"},
		{"tab", "a\tb"},
		{"control character", "a\x01b"},
		{"html characters", "a&b<c>"},
		{"unicode", "é-日本"},
		{"key injection attempt", `x","admin":true,"y":"z`},
	}

	for _, tc := range ids {
		t.Run(tc.name, func(t *testing.T) {
			orig := &Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: tc.resourceID}

			encoders := map[string]func() ([]byte, error){
				"MarshalJSONPB": func() ([]byte, error) { return orig.MarshalJSONPB(nil) },
				"jsonpb": func() ([]byte, error) {
					var buf bytes.Buffer
					err := (&jsonpb.Marshaler{}).Marshal(&buf, orig)
					return buf.Bytes(), err
				},
				"json": func() ([]byte, error) { return json.Marshal(orig) },
			}
			for name, encode := range encoders {
				out, err := encode()
				if err != nil {
					t.Fatalf("%s: marshal failed: %s", name, err)
				}
				var s string
				if err := json.Unmarshal(out, &s); err != nil {
					t.Fatalf("%s: output %q is not a single JSON string: %s", name, out, err)
				}

				var back Identifier
				if err := back.UnmarshalJSONPB(nil, out); err != nil {
					t.Fatalf("%s: unmarshal of %q failed: %s", name, out, err)
				}
				if back.ApplicationName != orig.ApplicationName || back.ResourceType != orig.ResourceType || back.ResourceId != orig.ResourceId {
					t.Errorf("%s: round trip changed %q/%q/%q into %q/%q/%q", name,
						orig.ApplicationName, orig.ResourceType, orig.ResourceId,
						back.ApplicationName, back.ResourceType, back.ResourceId)
				}
			}
		})
	}
}

func TestTruncateBytes(t *testing.T) {
	tcases := []struct {
		name   string
		data   []byte
		maxLen int
		want   string
	}{
		{"empty", nil, 4, ""},
		{"shorter than limit", []byte("abc"), 5, "abc"},
		{"exactly the limit", []byte("abc"), 3, "abc"},
		{"ascii cut", []byte("abcdef"), 3, "abc..."},
		{"cut on a rune boundary", []byte("éé"), 2, "é..."},
		{"cut inside a rune backs off", []byte("éé"), 3, "é..."},
		{"only continuation bytes back off to zero", []byte{0x80, 0x80, 0x80, 0x80}, 2, "..."},
		{"invalid bytes replaced when not cut", []byte("a\xffb"), 10, "a\uFFFDb"},
		{"invalid bytes replaced when cut", []byte("a\xffbcd"), 4, "a\uFFFDbc..."},
	}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := truncateBytes(tc.data, tc.maxLen); got != tc.want {
				t.Errorf("truncateBytes(%q, %d) = %q, want %q", tc.data, tc.maxLen, got, tc.want)
			}
		})
	}
}
