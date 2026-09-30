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

func TestIdentifier_UnmarhsalJSONPB(t *testing.T) {
	tcases := []struct {
		JSONData           string
		ExpectedIdentifier *Identifier
	}{
		{
			`"app/resource/res2"`,
			&Identifier{
				ApplicationName: "app",
				ResourceType:    "resource",
				ResourceId:      "res2",
			},
		},
		{
			`"null"`,
			&Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
		},
	}

	var (
		unmarshaler = &jsonpb.Unmarshaler{}
		buffer      = &bytes.Buffer{}
	)

	for _, tc := range tcases {
		buffer.Reset()
		buffer.WriteString(tc.JSONData)
		id := &Identifier{}

		if err := unmarshaler.Unmarshal(buffer, id); err != nil {
			t.Errorf("failed to unmarshal identifier %s", err)
		}

		if id.String() != tc.ExpectedIdentifier.String() {
			t.Errorf("invalid identifier %s, expected %s", id, tc.ExpectedIdentifier)
		}
	}
}

func TestIdentifier_UnmarshalJSONPB_InvalidTypes(t *testing.T) {
	tcases := []struct {
		Name     string
		JSONData string
	}{
		{"array", `["app/resource/id1","app/resource/id2"]`},
		{"empty_array", `[]`},
		{"number", `12345`},
		{"negative_number", `-1`},
		{"float_number", `1.5`},
		{"object", `{"application_name":"app","resource_type":"res","resource_id":"id1"}`},
		{"empty_object", `{}`},
		{"boolean_true", `true`},
		{"boolean_false", `false`},
	}

	for _, tc := range tcases {
		t.Run(tc.Name, func(t *testing.T) {
			id := &Identifier{}
			err := id.UnmarshalJSONPB(nil, []byte(tc.JSONData))
			if err == nil {
				t.Errorf("expected error for input %s, got nil", tc.JSONData)
			}
		})
	}
}

func TestIdentifier_UnmarshalJSONPB_ValidInputs(t *testing.T) {
	tcases := []struct {
		Name               string
		JSONData           string
		ExpectedIdentifier *Identifier
	}{
		{
			"valid resource identifier",
			`"app/resource/res1"`,
			&Identifier{ApplicationName: "app", ResourceType: "resource", ResourceId: "res1"},
		},
		{
			"partial identifier with only resource id",
			`"res1"`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: "res1"},
		},
		{
			"partial identifier with type and id",
			`"resource/res1"`,
			&Identifier{ApplicationName: "", ResourceType: "resource", ResourceId: "res1"},
		},
		{
			"null literal",
			`null`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
		{
			"quoted null",
			`"null"`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
		{
			"empty string",
			`""`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
		{
			"escaped slashes",
			`"app\/res\/id1"`,
			&Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "id1"},
		},
		{
			"surrounding whitespace",
			` "app/res/id1" `,
			&Identifier{ApplicationName: "app", ResourceType: "res", ResourceId: "id1"},
		},
		{
			"empty data",
			``,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.Name, func(t *testing.T) {
			id := &Identifier{}
			err := id.UnmarshalJSONPB(nil, []byte(tc.JSONData))
			if err != nil {
				t.Errorf("unexpected error for input %s: %s", tc.JSONData, err)
			}
			if id.String() != tc.ExpectedIdentifier.String() {
				t.Errorf("got %s, expected %s", id, tc.ExpectedIdentifier)
			}
		})
	}
}

func TestIdentifier_UnmarshalJSONPB_InvalidSyntax(t *testing.T) {
	for _, in := range []string{`"app/res/id1`, `"a" "b"`, `nul`, `[`} {
		if err := (&Identifier{}).UnmarshalJSONPB(nil, []byte(in)); err == nil {
			t.Errorf("expected error for input %s, got nil", in)
		}
	}
}

func TestIdentifier_UnmarshalJSONPB_WrapsDecoderError(t *testing.T) {
	var typeErr *json.UnmarshalTypeError
	if err := (&Identifier{}).UnmarshalJSONPB(nil, []byte(`123`)); !errors.As(err, &typeErr) {
		t.Errorf("expected *json.UnmarshalTypeError, got %v", err)
	}

	var syntaxErr *json.SyntaxError
	if err := (&Identifier{}).UnmarshalJSONPB(nil, []byte(`"app/res/id1`)); !errors.As(err, &syntaxErr) {
		t.Errorf("expected *json.SyntaxError, got %v", err)
	}
}

func TestIdentifier_UnmarshalJSONPB_ErrorMessageIsValidUTF8(t *testing.T) {
	long := `["` + strings.Repeat("é", 100) + `"]`
	// Every cut position, so each possible split inside a 2-byte rune is hit.
	for n := 1; n <= 70; n++ {
		if got := truncateBytes([]byte(long), n); !utf8.ValidString(got) {
			t.Fatalf("truncateBytes(_, %d) = %q, not valid UTF-8", n, got)
		}
	}

	err := (&Identifier{}).UnmarshalJSONPB(nil, []byte(long))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !utf8.ValidString(err.Error()) {
		t.Errorf("error text is not valid UTF-8: %q", err.Error())
	}

	if err := (&Identifier{}).UnmarshalJSONPB(nil, []byte("[\xff\xfe]")); err == nil || !utf8.ValidString(err.Error()) {
		t.Errorf("expected error with valid UTF-8 text for invalid input bytes, got %v", err)
	}
}

func TestIdentifier_UnmarshalViaDecoders_RejectsNonString(t *testing.T) {
	for _, in := range []string{`123`, `true`, `[]`, `{}`} {
		if err := jsonpb.Unmarshal(strings.NewReader(in), &Identifier{}); err == nil {
			t.Errorf("jsonpb.Unmarshal(%s): expected error, got nil", in)
		}

		var wrapper struct {
			ID *Identifier `json:"id"`
		}
		if err := json.Unmarshal([]byte(`{"id":`+in+`}`), &wrapper); err == nil {
			t.Errorf("json.Unmarshal of {\"id\":%s}: expected error, got nil", in)
		}
	}

	var wrapper struct {
		ID *Identifier `json:"id"`
	}
	if err := json.Unmarshal([]byte(`{"id":"app/res/id1"}`), &wrapper); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got := wrapper.ID.String(); got != "app/res/id1" {
		t.Errorf("got %s, expected app/res/id1", got)
	}
}
